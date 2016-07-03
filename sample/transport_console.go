package sample

import (
	"io"
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
	return nil
}

func (sink *ConsoleSink) Close() {
	if err := sink.stream.Close(); err != nil {
		log.Println("Error closing stdout output:", err)
	}
}

func (sink *ConsoleSink) Header(header Header) error {
	sink.Stop()
	sink.stream = sink.Writer.Open(nopWriteCloser{os.Stdout}, sink.Marshaller)
	return sink.stream.Header(header)
}

func (sink *ConsoleSink) Sample(sample Sample, header Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	return sink.stream.Sample(sample)
}

type ConsoleSource struct {
	AbstractMetricSource
	Reader SampleReader
	stream *SampleInputStream
}

func (source *ConsoleSource) String() string {
	return "ConsoleSource"
}

func (source *ConsoleSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.stream = source.Reader.Open(os.Stdin, source.OutgoingSink)
	return golib.WaitErrFunc(wg, func() error {
		defer source.CloseSink(wg)
		err := source.stream.ReadNamedSamples("stdin")
		if isFileClosedError(err) {
			err = nil
		}
		return err
	})
}

func (source *ConsoleSource) Stop() {
	err := source.stream.Close()
	if err != nil && !isFileClosedError(err) {
		log.Println("Error closing stdin:", err)
	}
}

// ====== Helper type

type nopWriteCloser struct {
	io.WriteCloser
}

func (nopWriteCloser) Close() error {
	return nil
}
