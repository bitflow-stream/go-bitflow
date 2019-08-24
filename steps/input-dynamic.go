package steps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

var _ bitflow.SampleSource = new(DynamicSource)

const DynamicSourceEndpointType = "dynamic"

var DynamicSourceParameters = reg.RegisteredParameters{}.
	Optional("update-time", reg.Duration(), 2*time.Second)

type DynamicSource struct {
	bitflow.AbstractSampleSource
	URL          string
	FetchTimeout time.Duration
	Endpoints    *bitflow.EndpointFactory

	loop            golib.LoopTask
	wg              *sync.WaitGroup
	previousSources []string

	currentSource  bitflow.SampleSource
	sourceStopChan golib.StopChan
	sourceWg       *sync.WaitGroup
	sourceClosed   *golib.BoolCondition
}

func RegisterDynamicSource(factory *bitflow.EndpointFactory) {
	factory.CustomDataSources[DynamicSourceEndpointType] = func(urlStr string) (bitflow.SampleSource, error) {
		url, params, err := reg.ParseEndpointUrlParams(urlStr, DynamicSourceParameters)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse dynamic URL '%v': %v", urlStr, err)
		}

		// TODO support https and query parameters (collides with endpoint specification)
		url.Scheme = "http"
		url.RawQuery = ""

		return &DynamicSource{
			URL:          url.String(),
			Endpoints:    factory,
			FetchTimeout: params["update-time"].(time.Duration),
		}, nil
	}
}

func (s *DynamicSource) String() string {
	return fmt.Sprintf("Dynamic source (%v, updated every %v)", s.URL, s.FetchTimeout)
}

func (s *DynamicSource) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	s.wg = wg
	s.loop.Description = fmt.Sprintf("Loop of %v", s)
	s.loop.StopHook = s.stopSource
	s.loop.Loop = s.updateSource
	return s.loop.Start(wg)
}

func (s *DynamicSource) Close() {
	s.loop.Stop()
	s.CloseSink()
}

func (s *DynamicSource) updateSource(stop golib.StopChan) error {
	sources, err := s.loadSources()
	if err != nil {
		log.Errorf("%v: Failed to fetch sources: %v", s, err)
	} else if !stop.Stopped() {
		sort.Strings(sources)
		if !golib.EqualStrings(sources, s.previousSources) {
			source, err := s.Endpoints.CreateInput(sources...)
			if err != nil {
				log.Errorf("%v: Failed to created new data source: %v", s, err)
			} else {
				s.startSource(source)
				s.previousSources = sources
			}
		}
	}
	stop.WaitTimeout(s.FetchTimeout)
	return nil
}

func (s *DynamicSource) loadSources() ([]string, error) {
	log.Debugf("%v: Fetching sources from %v", s, s.URL)
	resp, err := http.Get(s.URL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-successful response code: %v", resp.StatusCode)
	}
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var sources []string
	err = json.Unmarshal(bodyData, &sources)
	return sources, err
}

func (s *DynamicSource) stopSource() {
	if s.currentSource != nil {
		// Initiate closing the source
		s.currentSource.Close()

		// Wait for all close conditions
		s.sourceStopChan.Wait()
		s.sourceClosed.Wait()
		s.sourceWg.Wait()

		// Clean up resources for garbage collection
		s.currentSource = nil
		s.sourceWg = nil
		s.sourceClosed = nil
		s.sourceStopChan = golib.StopChan{}
	}
}

func (s *DynamicSource) startSource(src bitflow.SampleSource) {
	s.stopSource()
	s.currentSource = src
	s.sourceWg = new(sync.WaitGroup)
	s.sourceClosed = golib.NewBoolCondition()

	src.SetSink(&closeNotifier{
		SampleProcessor: s.GetSink(),
		cond:            s.sourceClosed,
	})
	s.sourceStopChan = src.Start(s.sourceWg)

	s.wg.Add(1)
	go s.observeSourceStopChan(src, s.sourceStopChan, s.sourceClosed)

	log.Printf("%v: Started data source %v", s, src)
}

func (s *DynamicSource) observeSourceStopChan(source bitflow.SampleSource, stopChan golib.StopChan, sourceClosed *golib.BoolCondition) {
	defer s.wg.Done()

	stopChan.Wait()
	sourceClosed.Wait()

	if err := stopChan.Err(); err == nil {
		log.Printf("%v: Source finished: %v", s, source)
	} else {
		log.Warnf("%v: Source finished with error: %v, error: %v", s, source, err)
	}
}

type closeNotifier struct {
	bitflow.SampleProcessor
	cond *golib.BoolCondition
}

func (c *closeNotifier) Close() {
	c.cond.Broadcast()
}
