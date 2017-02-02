package bitflow

import (
	"io"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
)

// WriterSink implements MetricSink by writing all Headers and Samples to a single
// io.WriteCloser instance. An instance of SampleReader is used to write the data in parallel.
type WriterSink struct {
	AbstractMarshallingMetricSink
	Output      io.WriteCloser
	Description string

	stream *SampleOutputStream
}

// NewConsoleSink creates a MetricSink that writes to the standard output.
func NewConsoleSink() *WriterSink {
	return &WriterSink{
		Output:      os.Stdout,
		Description: "stdout",
	}
}

// String implements the MetricSink interface.
func (sink *WriterSink) String() string {
	return sink.Description + " printer"
}

// Start implements the MetricSink interface. No additional goroutines are
// spawned, only a log message is printed.
func (sink *WriterSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.WithField("format", sink.Marshaller).Println("Printing samples to " + sink.Description)
	sink.stream = sink.Writer.Open(sink.Output, sink.Marshaller)
	return nil
}

// Close implements the MetricSink interface. It flushes the remaining data
// to stdout and marks the stream as closed, but does not actually close the
// stdout.
func (sink *WriterSink) Close() {
	if err := sink.stream.Close(); err != nil {
		log.Errorf("%v: Error closing output: %v", sink, err)
	}
}

// Header implements the MetricSink interface by using a SampleStream to
// write the given Sample to the standard output.
func (sink *WriterSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	return sink.stream.Sample(sample, header)
}

// ReaderSource implements the MetricSource interface by reading Headers and
// Samples from an arbitrary io.ReadCloser instance. An instance of SampleReader is used
// to read the data in parallel.
type ReaderSource struct {
	AbstractUnmarshallingMetricSource
	Input       io.ReadCloser
	Description string

	stream *SampleInputStream
}

// NewConsoleSource creates a MetricSource that reads from the standard input.
func NewConsoleSource() *ReaderSource {
	return &ReaderSource{
		Input:       os.Stdin,
		Description: "stdin",
	}
}

// String implements the MetricSource interface.
func (source *ReaderSource) String() string {
	return source.Description + " reader"
}

// Start implements the MetricSource interface.
func (source *ReaderSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.stream = source.Reader.Open(source.Input, source.OutgoingSink)
	return golib.WaitErrFunc(wg, func() error {
		defer source.CloseSink(wg)
		err := source.stream.ReadNamedSamples(source.Description)
		if isFileClosedError(err) {
			err = nil
		}
		return err
	})
}

// Stop implements the MetricSource interface. It stops the underlying stream
// and prints any errors to the logger.
func (source *ReaderSource) Stop() {
	// TODO closing the os.Stdin stream does not cause the current Read()
	// invokation to return... This data source will hang until stdin is closed
	// from the outside, or the program is stopped forcefully.
	err := source.stream.Close()
	if err != nil && !isFileClosedError(err) {
		log.Errorf("%v: error closing output: %v", source, err)
	}
}

// ====== Helper type

type nopWriteCloser struct {
	io.WriteCloser
}

func (n nopWriteCloser) Close() error {
	return nil
}
