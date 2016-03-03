package metrics

import (
	"os"
	"sync"
)

type ConsoleSink struct {
	abstractSink
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup, marshaller Marshaller) error {
	sink.marshaller = marshaller
	return nil
}

func (sink *ConsoleSink) Header(header Header) error {
	sink.header = header
	return sink.marshaller.WriteHeader(header, os.Stdout)
}

func (sink *ConsoleSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	return sink.marshaller.WriteSample(sample, os.Stdout)
}

type ConsoleSource struct {
}

func (source *ConsoleSource) Start(wg *sync.WaitGroup, um Unmarshaller, sink MetricSink) error {
	simpleReadSamples(wg, "stdin", os.Stdin, um, sink)
	return nil
}
