package bitflow

import (
	"io"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/antongulenko/golib/gotermBox"
	log "github.com/sirupsen/logrus"
)

// ConsoleBoxSink implements the SampleSink interface by printing the received
// samples to the standard out. Contrary to the ConsoleSink, the screen is erased
// before printing a new sample, and the output is embedded in a box that shows
// the last lines of log output at the bottom. ConsoleBoxSink does not implement
// MarshallingSampleSink, because it uses its own, fixed marshaller.
//
// Multiple embedded fields provide access to configuration options.
//
// Init() must be called as early as possible when using ConsoleBoxSink, to make
// sure that all log messages are capture and none are overwritten by the box.
type ConsoleBoxSink struct {
	AbstractSampleProcessor
	gotermBox.CliLogBoxTask

	// ImmediateScreenUpdate causes the console box to be updated immediately
	// whenever a sample is received by this ConsoleBoxSink. Otherwise, the screen
	// will be updated in regular intervals based on the settings in CliLogBoxTask.
	ImmediateScreenUpdate bool

	// DontForwardSamples has the same semantics as AbstractSampleOutput.DontForwardSamples.
	DontForwardSamples bool

	lock       sync.Mutex
	lastSample *Sample
	lastHeader *Header
}

// String implements the SampleSink interface.
func (sink *ConsoleBoxSink) String() string {
	return "ConsoleBoxSink"
}

// Start implements the SampleSink interface. It starts a goroutine
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
	}.WriteSample(sample, header, sample.NumTags() > 0, out)
}

// Close implements the SampleSink interface. It stops the screen refresh
// goroutine.
func (sink *ConsoleBoxSink) Close() {
	sink.CliLogBoxTask.Stop()
}

// Stop shadows the Stop() method from gotermBox.CliLogBoxTask to make sure
// that this SampleSink is actually closed in the Close() method.
func (sink *ConsoleBoxSink) Stop() {
}

// Sample implements the SampleSink interface. The latest sample is stored
// and displayed on the console on the next screen refresh. Intermediate
// samples might get lost without being displayed.
func (sink *ConsoleBoxSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	sink.lock.Lock()
	sink.lastSample = sample
	sink.lastHeader = header
	if sink.ImmediateScreenUpdate {
		sink.TriggerUpdate()
	}
	sink.lock.Unlock()

	if sink.DontForwardSamples {
		return nil
	}
	return sink.OutgoingSink.Sample(sample, header)
}
