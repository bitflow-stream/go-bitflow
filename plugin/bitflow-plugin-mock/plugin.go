package main

import (
	"time"

	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/plugin"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
	log "github.com/sirupsen/logrus"
)

const (
	DataSourceType    = "mock"
	DataProcessorName = "mock"
)

func main() {
	log.Fatalln("This package is intended to be loaded as a plugin, not executed directly")
}

var defaultDataSource = RandomSampleGenerator{
	Interval:   300 * time.Millisecond,
	ErrorAfter: 2000,
	CloseAfter: 1000,
	TimeOffset: 0,
	ExtraTags: map[string]string{
		"plugin": "mock",
	},
	Header: []string{
		"num", "x", "y", "z",
	},
}

// The Symbol to be loaded
var Plugin plugin.BitflowPlugin = new(pluginImpl)

type pluginImpl struct {
}

func (*pluginImpl) Name() string {
	return "mock-plugin"
}

func (p *pluginImpl) Init(registry reg.ProcessorRegistry) error {
	plugin.LogPluginDataSource(p, DataSourceType)
	registry.Endpoints.CustomDataSources[bitflow.EndpointType(DataSourceType)] = func(query string) (bitflow.SampleSource, error) {
		params, err := plugin.ParseQueryParameters(query)
		if err != nil {
			return nil, err
		}
		generator := defaultDataSource
		err = generator.ParseParams(params)
		return &generator, err
	}

	plugin.LogPluginProcessor(p, DataProcessorName)
	registry.RegisterAnalysisParamsErr(DataProcessorName, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		var res MockSampleProcessor
		err := res.ParseParams(params)
		if err == nil {
			pipeline.Add(&res)
		}
		return err
	}, DataProcessorName, reg.RequiredParams("print", "error"))

	return nil
}
