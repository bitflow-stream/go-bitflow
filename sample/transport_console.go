package sample

import (
	"log"
	"os"
	"sync"

	"github.com/antongulenko/golib"
)

type ConsoleSink struct {
	abstractSink
}

func (sink *ConsoleSink) String() string {
	return "ConsoleSink"
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Printing", sink.marshaller, "samples")
	return nil
}

func (sink *ConsoleSink) Stop() {
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

func (source *ConsoleSource) String() string {
	return "ConsoleSource"
}

func (source *ConsoleSource) Start(wg *sync.WaitGroup) golib.StopChan {
	return simpleReadSamples(wg, "stdin", os.Stdin, source.Unmarshaller, source.Sink)
}

func (source *ConsoleSource) Stop() {
	_ = os.Stdin.Close() // Drop error
}
