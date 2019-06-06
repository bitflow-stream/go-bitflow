package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

type RandomSampleGenerator struct {
	bitflow.AbstractSampleSource

	Interval   time.Duration
	ErrorAfter int
	CloseAfter int
	Header     []string
	TimeOffset time.Duration
	ExtraTags  map[string]string

	task             golib.LoopTask
	samplesGenerated int
	rnd              *rand.Rand
}

var _ bitflow.SampleSource = new(RandomSampleGenerator)

var SampleGeneratorParameters = reg.RegisteredParameters{}.
	Required("interval", reg.Duration()).
	Required("error", reg.Int()).
	Required("close", reg.Int()).
	Required("offset", reg.Duration())

func (p *RandomSampleGenerator) SetValues(params map[string]interface{}) {
	p.Interval = params["interval"].(time.Duration)
	p.ErrorAfter = params["error"].(int)
	p.CloseAfter = params["close"].(int)
	p.TimeOffset = params["offset"].(time.Duration)
}

func (p *RandomSampleGenerator) String() string {
	return fmt.Sprintf("Random Samples (every %v, error after %v, close after %v, time offset %v)", p.Interval, p.ErrorAfter, p.CloseAfter, p.TimeOffset)
}

func (p *RandomSampleGenerator) Start(wg *sync.WaitGroup) golib.StopChan {
	p.task.StopHook = p.GetSink().Close
	p.rnd = rand.New(rand.NewSource(time.Now().Unix()))
	p.task.Loop = p.generate
	return p.task.Start(wg)
}

func (p *RandomSampleGenerator) Close() {
	p.task.Stop()
}

func (p *RandomSampleGenerator) generate(stopper golib.StopChan) error {
	if p.samplesGenerated >= p.CloseAfter {
		return golib.StopLoopTask
	}
	if p.samplesGenerated >= p.ErrorAfter {
		return fmt.Errorf("Mock data source: Automatic error after %v samples", p.samplesGenerated)
	}

	header := bitflow.Header{
		Fields: make([]string, len(p.Header)),
	}
	sample := bitflow.Sample{
		Values: make([]bitflow.Value, len(p.Header)),
		Time:   time.Now().Add(p.TimeOffset),
	}
	for i, field := range p.Header {
		header.Fields[i] = field
		if i == 0 {
			sample.Values[i] = bitflow.Value(p.samplesGenerated)
		} else {
			sample.Values[i] = bitflow.Value(p.rnd.Float64())
		}
	}
	for key, value := range p.ExtraTags {
		sample.SetTag(key, value)
	}

	p.samplesGenerated++
	if err := p.GetSink().Sample(&sample, &header); err != nil {
		return err
	}
	stopper.WaitTimeout(p.Interval)
	return nil
}
