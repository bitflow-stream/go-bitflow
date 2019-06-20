package steps

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type RestEndpointFactory struct {
	endpoints map[string]*RestEndpoint
}

func (s *RestEndpointFactory) GetEndpoint(endpointString string) *RestEndpoint {
	if s.endpoints == nil {
		s.endpoints = make(map[string]*RestEndpoint)
	}
	endpoint, ok := s.endpoints[endpointString]
	if !ok {
		endpoint = &RestEndpoint{
			engine:   golib.NewGinEngine(),
			endpoint: endpointString,
		}
		s.endpoints[endpointString] = endpoint
	}
	return endpoint
}

type RestEndpoint struct {
	engine    *gin.Engine
	paths     []string
	startOnce sync.Once
	endpoint  string
	stoppers  []golib.StopChan
}

func (endpoint *RestEndpoint) start() {
	endpoint.startOnce.Do(func() {
		go func() {
			// This routine cannot be interrupted
			log.Println("Listening for REST input on", endpoint.endpoint)
			for _, path := range endpoint.paths {
				log.Println(" -", path)
			}
			if err := endpoint.engine.Run(endpoint.endpoint); err != nil {
				for _, stopper := range endpoint.stoppers {
					stopper.StopErr(err)
				}
			}
		}()
	})
}

func (endpoint *RestEndpoint) serve(verb string, path string, logFile string, serve func(*gin.Context)) (err error) {
	pathStr := fmt.Sprintf("[%s] %s", verb, path)
	// TODO check if pathStr is already present in paths, raise error if so

	handlers := gin.HandlersChain{serve}
	if logFile != "" {
		handlers = append(handlers, golib.LogGinRequests(logFile, true, true))
	}

	defer func() {
		if p := recover(); p != nil {
			// Recovered from panic in Handle()
			err = fmt.Errorf("Failed to register REST path %v: %v", path, p)
		}
	}()
	endpoint.engine.Handle(verb, path, handlers...)
	endpoint.paths = append(endpoint.paths, pathStr)
	return
}

func (endpoint *RestEndpoint) NewDataSource(outgoingSampleBuffer int) *RestDataSource {
	stopper := golib.NewStopChan()
	endpoint.stoppers = append(endpoint.stoppers, stopper)
	return &RestDataSource{
		outgoing: make(chan []bitflow.SampleAndHeader, outgoingSampleBuffer),
		endpoint: endpoint,
		stop:     stopper,
	}
}

type RestDataSource struct {
	bitflow.AbstractSampleSource

	outgoing  chan []bitflow.SampleAndHeader
	stop      golib.StopChan
	closeOnce sync.Once
	endpoint  *RestEndpoint
	wg        *sync.WaitGroup
}

func (source *RestDataSource) Endpoint() string {
	return source.endpoint.endpoint
}

func (source *RestDataSource) Serve(verb string, path string, httpLogFile string, serve func(*gin.Context)) error {
	return source.endpoint.serve(verb, path, httpLogFile, serve)
}

func (source *RestDataSource) EmitSamplesTimeout(samples []bitflow.SampleAndHeader, timeout time.Duration) bool {
	if timeout > 0 {
		select {
		case source.outgoing <- samples:
			return true
		case <-time.After(timeout):
			return false
		}
	} else {
		source.outgoing <- samples
		return true
	}
}

func (source *RestDataSource) EmitSampleAndHeaderTimeout(sample bitflow.SampleAndHeader, timeout time.Duration) bool {
	return source.EmitSamplesTimeout([]bitflow.SampleAndHeader{sample}, timeout)
}

func (source *RestDataSource) EmitSampleTimeout(sample *bitflow.Sample, header *bitflow.Header, timeout time.Duration) bool {
	return source.EmitSampleAndHeaderTimeout(bitflow.SampleAndHeader{Sample: sample, Header: header}, timeout)
}

func (source *RestDataSource) EmitSamples(samples []bitflow.SampleAndHeader) {
	source.EmitSamplesTimeout(samples, 0)
}

func (source *RestDataSource) EmitSampleAndHeader(sample bitflow.SampleAndHeader) {
	source.EmitSampleAndHeaderTimeout(sample, 0)
}

func (source *RestDataSource) EmitSample(sample *bitflow.Sample, header *bitflow.Header) {
	source.EmitSampleTimeout(sample, header, 0)
}

func (source *RestDataSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.wg = wg
	wg.Add(1)
	go source.sinkSamples(wg)
	source.endpoint.start()
	return source.stop
}

func (source *RestDataSource) Close() {
	if source.wg == nil {
		return
	}
	source.closeOnce.Do(func() {
		source.stop.Stop()
		close(source.outgoing)
		source.AbstractSampleSource.CloseSinkParallel(source.wg)
	})
}

func (source *RestDataSource) String() string {
	return fmt.Sprintf("REST data source (%v, buffer %v)", source.endpoint.endpoint, cap(source.outgoing))
}

func (source *RestDataSource) sinkSamples(wg *sync.WaitGroup) {
	defer wg.Done()
	for samples := range source.outgoing {
		for _, sample := range samples {
			if err := source.GetSink().Sample(sample.Sample, sample.Header); err != nil {
				log.Errorf("%v: Error sinking sample: %v", source, err)
			}
		}
	}
}

type RestReplyHelpers struct {
}

func (h RestReplyHelpers) Reply(context *gin.Context, message string, statusCode int, contentType string) {
	if statusCode != http.StatusOK {
		log.Warnf("REST status %v: %v", statusCode, message)
	}

	context.Writer.Header().Set("Content-Type", contentType)
	context.Writer.WriteHeader(statusCode)
	_, err := context.Writer.WriteString(message + "\n")
	if err != nil {
		log.Errorf("REST: Error sending %v reply (len %v, content-type %v) to client: %v", statusCode, len(message), contentType, err)
	}
}

func (h RestReplyHelpers) ReplyCode(context *gin.Context, message string, statusCode int) {
	h.Reply(context, message, statusCode, "text/plain")
}

func (h RestReplyHelpers) ReplySuccess(context *gin.Context, message string) {
	h.ReplyCode(context, message, http.StatusOK)
}

func (h RestReplyHelpers) ReplyGenericError(context *gin.Context, errorMessage string) {
	h.ReplyCode(context, errorMessage, http.StatusBadRequest)
}

func (h RestReplyHelpers) ReplyError(context *gin.Context, errorMessage string) {
	h.ReplyGenericError(context, "Wrong HTTP request: "+errorMessage)
}
