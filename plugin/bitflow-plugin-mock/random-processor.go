package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow"
)

type MockSampleProcessor struct {
	bitflow.NoopProcessor
	ErrorAfter  int
	PrintModulo int

	task             golib.LoopTask
	samplesGenerated int
}

var _ bitflow.SampleProcessor = new(MockSampleProcessor)

func (p *MockSampleProcessor) String() string {
	return fmt.Sprintf("Mock processor (log every %v sample(s), error after %v)", p.PrintModulo, p.ErrorAfter)
}

func (p *MockSampleProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if p.samplesGenerated >= p.ErrorAfter {
		return fmt.Errorf("Mock processor: Automatic error after %v samples", p.samplesGenerated)
	}
	if p.PrintModulo > 0 && p.samplesGenerated%p.PrintModulo == 0 {
		log.Println("Mock processor processing sample nr", p.samplesGenerated)
	}
	p.samplesGenerated++
	return p.NoopProcessor.Sample(sample, header)
}

func (p *MockSampleProcessor) ParseParams(paramsIn map[string]string) error {
	// Make copy to avoid modifying value passed in from outside
	params := make(map[string]string, len(paramsIn))
	for key, value := range paramsIn {
		params[key] = value
	}

	var err error
	if intervalStr, ok := params["print"]; err == nil && ok {
		p.PrintModulo, err = strconv.Atoi(intervalStr)
		delete(params, "interval")
	}
	if errorStr, ok := params["error"]; err == nil && ok {
		p.ErrorAfter, err = strconv.Atoi(errorStr)
		delete(params, "error")
	}
	if err == nil && len(params) > 0 {
		err = fmt.Errorf("Unexpected parameters: %v", params)
	}
	return err
}
