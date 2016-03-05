package metrics

import (
	"log"
	"os"
	"sync"
)

type ConsoleSink struct {
	abstractSink
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup, marshaller Marshaller) error {
	log.Println("Printing", marshaller, "samples")
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
	unmarshallingMetricSource
}

func (source *ConsoleSource) Start(wg *sync.WaitGroup, sink MetricSink) error {
	simpleReadSamples(wg, "stdin", os.Stdin, source.Unmarshaller, sink)
	return nil
}
