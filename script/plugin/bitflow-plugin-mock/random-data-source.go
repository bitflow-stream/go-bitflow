package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
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

func (p *RandomSampleGenerator) ParseParams(paramsIn map[string]string) error {
	// Make copy to avoid modifying value passed in from outside
	params := make(map[string]string, len(paramsIn))
	for key, value := range paramsIn {
		params[key] = value
	}

	var err error
	if intervalStr, ok := params["interval"]; err == nil && ok {
		p.Interval, err = time.ParseDuration(intervalStr)
		delete(params, "interval")
	}
	if errorStr, ok := params["error"]; err == nil && ok {
		p.ErrorAfter, err = strconv.Atoi(errorStr)
		delete(params, "error")
	}
	if closeStr, ok := params["close"]; err == nil && ok {
		p.CloseAfter, err = strconv.Atoi(closeStr)
		delete(params, "close")
	}
	if offsetStr, ok := params["offset"]; err == nil && ok {
		p.TimeOffset, err = time.ParseDuration(offsetStr)
		delete(params, "offset")
	}
	if err == nil && len(params) > 0 {
		err = fmt.Errorf("Unexpected parameters: %v", params)
	}
	return err
}
