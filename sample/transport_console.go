package sample

import (
	"log"
	"os"
	"sync"

	"github.com/antongulenko/golib"
)

type ConsoleSink struct {
	AbstractMarshallingMetricSink
	stream *SampleOutputStream
}

func (sink *ConsoleSink) String() string {
	return "ConsoleSink"
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Printing", sink.Marshaller, "samples")
	sink.stream = sink.Writer.Open(os.Stdout, sink.Marshaller)
	return nil
}

func (sink *ConsoleSink) Close() {
	if err := sink.stream.Close(); err != nil {
		log.Println("Error closing stdout output:", err)
	}
	_ = os.Stdout.Close() // Drop error
}

func (sink *ConsoleSink) Header(header Header) error {
	return sink.stream.Header(header)
}

func (sink *ConsoleSink) Sample(sample Sample, header Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	return sink.stream.Sample(sample)
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
