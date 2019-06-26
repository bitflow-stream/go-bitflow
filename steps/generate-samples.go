package steps

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

var _ bitflow.SampleSource = new(GeneratorSource)

const GeneratorSourceEndpointType = "generate"

var GeneratorSourceParameters = reg.RegisteredParameters{}.
	Optional("interval", reg.Duration(), 500*time.Millisecond)

type GeneratorSource struct {
	bitflow.AbstractSampleSource

	loop      golib.LoopTask
	sleepTime time.Duration
}

func RegisterGeneratorSource(factory *bitflow.EndpointFactory) {
	factory.CustomDataSources[GeneratorSourceEndpointType] = func(urlStr string) (bitflow.SampleSource, error) {
		_, params, err := reg.ParseEndpointUrlParams(urlStr, GeneratorSourceParameters)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse URL '%v': %v", urlStr, err)
		}

		return &GeneratorSource{
			sleepTime: params["interval"].(time.Duration),
		}, nil
	}
}

func (s *GeneratorSource) String() string {
	return fmt.Sprintf("Sample generator (generated every %v)", s.sleepTime)
}

func (s *GeneratorSource) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	s.loop.Description = fmt.Sprintf("Loop of %v", s)
	s.loop.Loop = s.sendSample
	return s.loop.Start(wg)
}

func (s *GeneratorSource) Close() {
	s.loop.Stop()
	s.CloseSink()
}

func (s *GeneratorSource) sendSample(stop golib.StopChan) error {
	sample, header := s.generateSample()
	if err := s.GetSink().Sample(sample, header); err != nil {
		log.Errorf("%v: Error sinking sample: %v", s, err)
	}
	stop.WaitTimeout(s.sleepTime)
	return nil
}

func (s *GeneratorSource) generateSample() (*bitflow.Sample, *bitflow.Header) {
	sample := &bitflow.Sample{
		Time: time.Now(),
		Values: []bitflow.Value{
			bitflow.Value(math.NaN()),
			bitflow.Value(math.Inf(1)),
			bitflow.Value(math.Inf(-1)),
			-math.MaxFloat64,
			math.MaxFloat64,
		},
	}
	header := &bitflow.Header{
		Fields: []string{
			"nan", "plusInfinite", "minusInfinite", "min", "max",
		},
	}
	return sample, header
}
