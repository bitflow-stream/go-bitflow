package bitflow

import (
	"io"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
)

// ConsoleSink implements MetricSink by writing all Headers and Samples to the
// standard output. An instance of SampleReader is used to write the data in parllel.
type ConsoleSink struct {
	AbstractMarshallingMetricSink
	stream *SampleOutputStream
}

// String implements the MetricSink interface.
func (sink *ConsoleSink) String() string {
	return "ConsoleSink"
}

// Start implements the MetricSink interface. No additional goroutines are
// spawned, only a log message is printed.
func (sink *ConsoleSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.WithField("format", sink.Marshaller).Println("Printing samples")
	sink.stream = sink.Writer.Open(nopWriteCloser{os.Stdout}, sink.Marshaller)
	return nil
}

// Close implements the MetricSink interface by actually closing the
// standard output.
func (sink *ConsoleSink) Close() {
	if err := sink.stream.Close(); err != nil {
		log.Errorln("Error closing stdout output:", err)
	}
}

// Header implements the MetricSink interface by using a SampleStream to
// write the given Sample to the standard output.
func (sink *ConsoleSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	return sink.stream.Sample(sample, header)
}

// ConsoleSource implements the MetricSource interface by reading Headers and
// Samples from the standard input stream. An instance of SampleReader is used
// to read the data in parllel.
type ConsoleSource struct {
	AbstractMetricSource

	// Reader contains configuration variables that controll the unmarshalling
	// process. It must be configured before calling Start().
	Reader SampleReader

	stream *SampleInputStream
}

// String implements the MetricSource interface.
func (source *ConsoleSource) String() string {
	return "ConsoleSource"
}

// Start implements the MetricSource interface.
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

// Stop implements the MetricSource interface. It stops the underlying stream
// and prints any errors to the logger.
func (source *ConsoleSource) Stop() {
	err := source.stream.Close()
	if err != nil && !isFileClosedError(err) {
		log.Errorln("Error closing stdin:", err)
	}
}

// ====== Helper type

type nopWriteCloser struct {
	io.WriteCloser
}

func (nopWriteCloser) Close() error {
	return nil
}
