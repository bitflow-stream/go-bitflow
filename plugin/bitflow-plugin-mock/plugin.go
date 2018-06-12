package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline/plugin"
	"github.com/antongulenko/golib"
)

func main() {
	log.Fatalln("This package is intended to be loaded as a plugin, not executed directly")
}

// The Symbol to be loaded
var Plugin = MockPlugin{
	ChannelBuf: 10,
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

type MockPlugin struct {
	ChannelBuf int
	Interval   time.Duration
	ErrorAfter int
	CloseAfter int
	Header     []string
	TimeOffset time.Duration
	ExtraTags  map[string]string

	dataSink         plugin.PluginDataSink
	task             golib.LoopTask
	samplesGenerated int
	rnd              *rand.Rand
}

func (p *MockPlugin) Start(params map[string]string, dataSink plugin.PluginDataSink) {
	if err := p.parseParams(params); err != nil {
		dataSink.Error(err)
		return
	}
	p.dataSink = dataSink
	p.task.StopHook = func() {
		dataSink.Close()
	}
	p.rnd = rand.New(rand.NewSource(time.Now().Unix()))
	p.task.Loop = p.generate
	_ = p.task.Start(nil) // Using the StopHook replaces using the golib.StopChan
}

func (p *MockPlugin) Close() {
	p.task.Stop()
}

func (p *MockPlugin) generate(stopper golib.StopChan) error {
	if p.samplesGenerated >= p.CloseAfter {
		p.task.Stop()
		return nil
	}
	if p.samplesGenerated >= p.ErrorAfter {
		p.dataSink.Error(fmt.Errorf("Automatic error after %v samples", p.samplesGenerated))
		return nil
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

	p.dataSink.Sample(&sample, &header)
	p.samplesGenerated++
	stopper.WaitTimeout(p.Interval)
	return nil
}

func (p *MockPlugin) parseParams(paramsIn map[string]string) error {
	// Make copy to avoid modifying value passed in from outside
	params := make(map[string]string, len(paramsIn))
	for key, value := range paramsIn {
		params[key] = value
	}

	var err error
	if channelBufStr, ok := params["channel"]; ok {
		p.ChannelBuf, err = strconv.Atoi(channelBufStr)
		delete(params, "channel")
	}
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
