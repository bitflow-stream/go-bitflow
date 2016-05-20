package sample

import (
	"log"
	"os"
	"sync"

	"github.com/antongulenko/golib"
)

type ConsoleSink struct {
	AbstractMarshallingMetricSink
	Reader SampleReader
}

func (sink *ConsoleSink) String() string {
	return "ConsoleSink"
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Printing", sink.Marshaller, "samples")
	return nil
}

func (sink *ConsoleSink) Close() {
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
	Reader SampleReader
}

func (source *ConsoleSource) String() string {
	return "ConsoleSource"
}

func (source *ConsoleSource) Start(wg *sync.WaitGroup) golib.StopChan {
	return golib.WaitErrFunc(wg, func() error {
		err := source.Reader.ReadNamedSamples("stdin", os.Stdin, source.Unmarshaller, source.OutgoingSink)
		source.CloseSink(wg)
		return err
	})
}

func (source *ConsoleSource) Stop() {
	_ = os.Stdin.Close() // Drop error
}
