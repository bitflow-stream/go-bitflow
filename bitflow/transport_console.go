package bitflow

import (
	"io"
	"os"
	"sync"

	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

// WriterSink implements SampleSink by writing all Headers and Samples to a single
// io.WriteCloser instance. An instance of SampleWriter is used to write the data in parallel.
type WriterSink struct {
	AbstractMarshallingSampleOutput
	Output      io.WriteCloser
	Description string

	stream *SampleOutputStream
}

// NewConsoleSink creates a SampleSink that writes to the standard output.
func NewConsoleSink() *WriterSink {
	return &WriterSink{
		Output:      os.Stdout,
		Description: "stdout",
	}
}

func (sink *WriterSink) WritesToConsole() bool {
	return sink.Output == os.Stdout
}

// String implements the SampleSink interface.
func (sink *WriterSink) String() string {
	return sink.Description + " printer"
}

// Start implements the SampleSink interface. No additional goroutines are
// spawned, only a log message is printed.
func (sink *WriterSink) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	log.WithField("format", sink.Marshaller).Println("Printing samples to " + sink.Description)
	sink.stream = sink.Writer.Open(sink.Output, sink.Marshaller)
	return
}

// Close implements the SampleSink interface. It flushes the remaining data
// to the underlying io.WriteCloser and closes it.
func (sink *WriterSink) Close() {
	if err := sink.stream.Close(); err != nil {
		log.Errorf("%v: Error closing output: %v", sink, err)
	}
	sink.CloseSink()
}

// Header implements the SampleSink interface by using a SampleOutputStream to
// write the given Sample to the configured io.WriteCloser.
func (sink *WriterSink) Sample(sample *Sample, header *Header) error {
	err := sink.stream.Sample(sample, header)
	return sink.AbstractMarshallingSampleOutput.Sample(err, sample, header)
}

// ReaderSource implements the SampleSource interface by reading Headers and
// Samples from an arbitrary io.ReadCloser instance. An instance of SampleReader is used
// to read the data in parallel.
type ReaderSource struct {
	AbstractUnmarshallingSampleSource
	Input       io.ReadCloser
	Description string

	stream *SampleInputStream
}

// NewConsoleSource creates a SampleSource that reads from the standard input.
func NewConsoleSource() *ReaderSource {
	return &ReaderSource{
		Input:       os.Stdin,
		Description: "stdin",
	}
}

// String implements the SampleSource interface.
func (source *ReaderSource) String() string {
	return source.Description + " reader"
}

// Start implements the SampleSource interface by starting a SampleInputStream
// instance that reads from the given io.ReadCloser.
func (source *ReaderSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.stream = source.Reader.Open(source.Input, source.GetSink())
	return golib.WaitErrFunc(wg, func() error {
		defer source.CloseSinkParallel(wg)
		err := source.stream.ReadNamedSamples(source.Description)
		if IsFileClosedError(err) {
			err = nil
		}
		return err
	})
}

// Close implements the SampleSource interface. It stops the underlying stream
// and prints any errors to the logger.
func (source *ReaderSource) Close() {
	// TODO closing the os.Stdin stream does not cause the current Read()
	// invocation to return... This data source will hang until stdin is closed
	// from the outside, or the program is stopped forcefully.
	err := source.stream.Close()
	if err != nil && !IsFileClosedError(err) {
		log.Errorf("%v: error closing output: %v", source, err)
	}
}
