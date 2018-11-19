package plugin

import (
	"fmt"
	"net/url"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow"
	log "github.com/sirupsen/logrus"
)

type SampleSource struct {
	bitflow.AbstractSampleSource

	path    string
	symbol  string
	params  map[string]string
	plugin  SampleSourcePlugin
	stopper golib.StopChan
	wg      *sync.WaitGroup
}

func RegisterPluginDataSource(endpoints *bitflow.EndpointFactory) {
	endpoints.CustomDataSources["plugin"] = func(urlStr string) (bitflow.SampleSource, error) {
		endpoint, fullPath, params, err := ParseUrl(urlStr)
		if err == nil && endpoint != "" {
			return nil, fmt.Errorf("URL for plugin may not contain hostname: %v", urlStr)
		}
		if err != nil {
			return nil, err
		}
		path, symbol := filepath.Split(fullPath)
		if path == "" || symbol == "" {
			return nil, fmt.Errorf("URL for plugin must have path with at least two components: %v", urlStr)
		}
		path = path[:len(path)-1] // Strip trailing slash
		return NewPluginSource(path, symbol, params)
	}
}

func ParseUrl(urlStr string) (string, string, map[string]string, error) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", "", nil, fmt.Errorf("Failed to parse plugin URL: %v", err)
	}
	urlParams := parsedUrl.Query()
	params := make(map[string]string, len(urlParams))
	for key, value := range urlParams {
		paramVal := ""
		if len(value) == 1 {
			paramVal = value[0]
		} else if len(value) > 1 {
			return "", "", nil, fmt.Errorf("Multiple values for URL query key '%v': %v", key, value)
		}
		params[key] = paramVal
	}
	return parsedUrl.Host, parsedUrl.Path, params, nil
}

func NewPluginSource(path, symbolName string, params map[string]string) (*SampleSource, error) {
	res := &SampleSource{
		path:   path,
		symbol: symbolName,
		params: params,
	}
	openedPlugin, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	symbolObject, err := openedPlugin.Lookup(symbolName)
	if err != nil {
		return nil, err
	}
	sourcePlugin, ok := symbolObject.(SampleSourcePlugin)
	if !ok {
		return nil, fmt.Errorf(
			"Symbol '%v' from plugin '%v' has type %T instead of SampleSourcePlugin", symbolName, path, symbolObject)
	}
	res.plugin = sourcePlugin
	return res, nil
}

func (s *SampleSource) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Starting:", s.String())
	s.stopper = golib.NewStopChan()
	s.wg = wg
	s.plugin.Start(s.params, &pluginDataSink{s})
	return s.stopper
}

func (s *SampleSource) Close() {
	if p := s.plugin; p != nil {
		// Tell the plugin to close, it will then call Close() on the *pluginDataSink
		p.Close()
	}
}

func (s *SampleSource) String() string {
	return fmt.Sprintf("Plugin %v, symbol %v (parameters: %v)", s.path, s.symbol, s.params)
}

type pluginDataSink struct {
	source *SampleSource
}

func (s *pluginDataSink) Error(err error) {
	s.source.stopper.StopErr(err)
}

func (s *pluginDataSink) Close() {
	s.source.stopper.Stop()
	s.source.CloseSinkParallel(s.source.wg)
}

func (s *pluginDataSink) Sample(sample *bitflow.Sample, header *bitflow.Header) {
	if err := s.source.GetSink().Sample(sample, header); err != nil {
		s.Error(err)
	}
}
