package sample

import (
	"log"
	"os"
	"sync"

	"github.com/antongulenko/golib"
)

type ConsoleSink struct {
	AbstractMarshallingMetricSink
}

func (sink *ConsoleSink) String() string {
	return "ConsoleSink"
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Printing", sink.Marshaller, "samples")
	return nil
}

func (sink *ConsoleSink) Stop() {
}

func (sink *ConsoleSink) Header(header Header) error {
	return sink.Marshaller.WriteHeader(header, os.Stdout)
}

func (sink *ConsoleSink) Sample(sample Sample, header Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	return sink.Marshaller.WriteSample(sample, header, os.Stdout)
}

type ConsoleSource struct {
	AbstractUnmarshallingMetricSource
}

func (source *ConsoleSource) String() string {
	return "ConsoleSource"
}

func (source *ConsoleSource) Start(wg *sync.WaitGroup) golib.StopChan {
	return simpleReadSamples(wg, "stdin", os.Stdin, source.Unmarshaller, source.OutgoingSink)
}

func (source *ConsoleSource) Stop() {
	_ = os.Stdin.Close() // Drop error
}
