package plot

import (
	"fmt"
	"sort"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/steps"
)

func RegisterHttpPlotter(b reg.ProcessorRegistry) {
	create := func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
		windowSize := params["window"].(int)
		static := params["local-static"].(bool)
		p.Add(NewHttpPlotter(params["endpoint"].(string), windowSize, static))
		return nil
	}
	b.RegisterStep("http", create,
		"Serve HTTP-based plots about processed metrics values to the given HTTP endpoint").
		Required("endpoint", reg.String()).
		Optional("window", reg.Int(), 100).
		Optional("local-static", reg.Bool(), false)
}

func NewHttpPlotter(endpoint string, windowSize int, useLocalStatic bool) *HttpPlotter {
	return &HttpPlotter{
		data:           make(map[string]*steps.MetricWindow),
		Endpoint:       endpoint,
		WindowSize:     windowSize,
		UseLocalStatic: useLocalStatic,
	}
}

type HttpPlotter struct {
	bitflow.NoopProcessor

	Endpoint       string
	WindowSize     int
	UseLocalStatic bool

	data  map[string]*steps.MetricWindow
	names []string
}

func (p *HttpPlotter) Start(wg *sync.WaitGroup) golib.StopChan {
	go func() {
		// This routine cannot be interrupted gracefully
		if err := p.serve(); err != nil {
			p.Error(err)
		}
	}()
	return p.NoopProcessor.Start(wg)
}

func (p *HttpPlotter) String() string {
	endpoint := p.Endpoint
	if endpoint == "" {
		endpoint = "0.0.0.0:80"
	}
	return fmt.Sprintf("HTTP plotter on %v (window size %v)", endpoint, p.WindowSize)
}

func (p *HttpPlotter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.logSample(sample, header)
	return p.NoopProcessor.Sample(sample, header)
}

func (p *HttpPlotter) logSample(sample *bitflow.Sample, header *bitflow.Header) {
	for i, field := range header.Fields {
		if _, ok := p.data[field]; !ok {
			p.data[field] = steps.NewMetricWindow(p.WindowSize)
			p.names = append(p.names, field)
			sort.Strings(p.names)
		}
		p.data[field].Push(sample.Values[i])
	}
}

func (p *HttpPlotter) metricNames() []string {
	return p.names
}

func (p *HttpPlotter) metricData(metric string) []bitflow.Value {
	if data, ok := p.data[metric]; ok {
		return data.Data()
	} else {
		return []bitflow.Value{}
	}
}

func (p *HttpPlotter) allMetricData() map[string][]bitflow.Value {
	result := make(map[string][]bitflow.Value)
	for name, values := range p.data {
		result[name] = values.Data()
	}
	return result
}
