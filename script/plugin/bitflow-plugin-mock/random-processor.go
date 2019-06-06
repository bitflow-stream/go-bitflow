package main

import (
	"fmt"
	"log"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
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
