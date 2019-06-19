package main

import (
	"time"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/plugin"
	"github.com/bitflow-stream/go-bitflow/script/reg"
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
	registry.Endpoints.CustomDataSources[bitflow.EndpointType(DataSourceType)] = func(endpointUrl string) (bitflow.SampleSource, error) {
		_, params, err := reg.ParseEndpointUrlParams(endpointUrl, SampleGeneratorParameters)
		if err != nil {
			return nil, err
		}
		generator := defaultDataSource
		generator.SetValues(params)
		return &generator, err
	}

	plugin.LogPluginProcessor(p, DataProcessorName)
	registry.RegisterStep(DataProcessorName, func(pipeline *bitflow.SamplePipeline, params map[string]interface{}) error {
		pipeline.Add(&MockSampleProcessor{
			PrintModulo: params["print"].(int),
			ErrorAfter:  params["error"].(int),
		})
		return nil
	}, DataProcessorName).
		Required("print", reg.Int()).
		Required("error", reg.Int())

	return nil
}
