package bitflow

import (
	"io"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/golib"
	"github.com/antongulenko/golib/gotermBox"
)

// ConsoleBoxSink implements the MetricSink interface by printing the received
// samples to the standard out. Contrary to the ConsoleSink, the screen is erased
// before printing a new sample, and the output is embedded in a box that shows
// the last lines of log output at the bottom. ConsoleBoxSink does not implement
// MarshallingMetricSink, because it uses its own, fixed marshaller.
//
// Multiple fields provide access to configuration options.
type ConsoleBoxSink struct {
	AbstractMetricSink
	gotermBox.CliLogBoxTask

	lock       sync.Mutex
	lastSample *Sample
	lastHeader *Header
}

// String implements the MetricSink interface.
func (sink *ConsoleBoxSink) String() string {
	return "ConsoleBoxSink"
}

// Start implements the MetricSink interface. It starts a goroutine
// that regularly refreshes the screen to display the current sample values
// and latest log output lines.
func (sink *ConsoleBoxSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Printing samples to table")
	sink.CliLogBoxTask.Update = sink.updateBox
	return sink.CliLogBoxTask.Start(wg)
}

func (sink *ConsoleBoxSink) updateBox(out io.Writer, textWidth int) error {
	sink.lock.Lock()
	sample := sink.lastSample
	header := sink.lastHeader
	sink.lock.Unlock()
	if sample == nil || header == nil {
		return nil
	}
	return TextMarshaller{
		TextWidth: textWidth,
	}.WriteSample(sample, header, out)
}

// Close implements the MetricSink interface. It stops the screen refresh
// goroutine.
func (sink *ConsoleBoxSink) Close() {
	sink.CliLogBoxTask.Stop()
}

// Stop shadows the Stop() method from gotermBox.CliLogBoxTask to make sure
// that this MetricSink is actually closed in the Close() method.
func (sink *ConsoleBoxSink) Stop() {
}

// Sample implements the MetricSink interface. The latest sample is stored
// and displayed on the console on the next screen refresh. Intermediate
// samples might get lost without being displayed.
func (sink *ConsoleBoxSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	sink.lock.Lock()
	sink.lastSample = sample
	sink.lastHeader = header
	sink.lock.Unlock()
	return nil
}
