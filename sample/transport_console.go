package sample

import (
	"bufio"
	"log"
	"os"
	"sync"

	"github.com/antongulenko/golib"
)

type ConsoleSink struct {
	AbstractMarshallingMetricSink
	buf *bufio.Writer
}

func (sink *ConsoleSink) String() string {
	return "ConsoleSink"
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup) golib.StopChan {
	sink.buf = sink.BufferedWriter(os.Stdout)
	log.Println("Printing", sink.Marshaller, "samples")
	return nil
}

func (sink *ConsoleSink) Close() {
	if err := sink.buf.Flush(); err != nil {
		log.Println("Error flushing stdout:", err)
	}
	_ = os.Stdout.Close() // Drop error
}

func (sink *ConsoleSink) Header(header Header) error {
	return sink.Marshaller.WriteHeader(header, sink.buf)
}

func (sink *ConsoleSink) Sample(sample Sample, header Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	return sink.Marshaller.WriteSample(sample, header, sink.buf)
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
