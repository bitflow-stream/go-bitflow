package steps

import (
	"flag"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/antongulenko/golib/gotermBox"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	log "github.com/sirupsen/logrus"
)

const ConsoleBoxEndpoint = bitflow.EndpointType("box")

var (
	ConsoleBoxSettings = gotermBox.CliLogBox{
		NoUtf8:        false,
		LogLines:      10,
		MessageBuffer: 500,
	}
	ConsoleBoxUpdateInterval    = 500 * time.Millisecond
	ConsoleBoxMinUpdateInterval = 50 * time.Millisecond

	// ConsoleBoxOutputTestMode is a flag used by tests to suppress initialization routines
	// that are not testable. It is a hack to keep the EndpointFactory easy to use
	// while making it testable.
	ConsoleBoxOutputTestMode = false
)

func RegisterConsoleBoxOutput(e *bitflow.EndpointFactory) {
	var factory consoleBoxFactory
	e.CustomDataSinks[ConsoleBoxEndpoint] = factory.createConsoleBox
	e.CustomOutputFlags = append(e.CustomOutputFlags, factory.registerFlags)
}

type consoleBoxFactory struct {
	ConsoleBoxNoImmediateScreenUpdate bool
}

func (factory *consoleBoxFactory) registerFlags(f *flag.FlagSet) {
	f.BoolVar(&factory.ConsoleBoxNoImmediateScreenUpdate, "slow-screen-updates", false, fmt.Sprintf("For console box output, don't update the screen on every sample, but only in intervals of %v", ConsoleBoxUpdateInterval))
}

func (factory *consoleBoxFactory) createConsoleBox(target string) (bitflow.SampleProcessor, error) {
	if target != bitflow.StdTransportTarget {
		return nil, fmt.Errorf("Transport '%v' can only be defined with target '%v' (received '%v')", ConsoleBoxEndpoint, bitflow.StdTransportTarget, target)
	}
	sink := &ConsoleBoxSink{
		CliLogBoxTask: gotermBox.CliLogBoxTask{
			CliLogBox:         ConsoleBoxSettings,
			UpdateInterval:    ConsoleBoxUpdateInterval,
			MinUpdateInterval: ConsoleBoxMinUpdateInterval,
		},
		ImmediateScreenUpdate: !factory.ConsoleBoxNoImmediateScreenUpdate,
	}
	if !ConsoleBoxOutputTestMode {
		sink.Init()
	}
	return sink, nil
}

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
	bitflow.AbstractSampleOutput
	gotermBox.CliLogBoxTask

	// ImmediateScreenUpdate causes the console box to be updated immediately
	// whenever a sample is received by this ConsoleBoxSink. Otherwise, the screen
	// will be updated in regular intervals based on the settings in CliLogBoxTask.
	ImmediateScreenUpdate bool

	lock       sync.Mutex
	lastSample *bitflow.Sample
	lastHeader *bitflow.Header
}

// Implement the bitflow.ConsoleSampleSink interface
func (sink *ConsoleBoxSink) WritesToConsole() bool {
	return true
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
	return bitflow.TextMarshaller{
		TextWidth: textWidth,
	}.WriteSample(sample, header, sample.NumTags() > 0, out)
}

// Close implements the SampleSink interface. It stops the screen refresh
// goroutine.
func (sink *ConsoleBoxSink) Close() {
	sink.CliLogBoxTask.Stop()
	sink.CloseSink()
}

// Stop shadows the Stop() method from gotermBox.CliLogBoxTask to make sure
// that this SampleSink is actually closed in the Close() method.
func (sink *ConsoleBoxSink) Stop() {
}

// Sample implements the SampleSink interface. The latest sample is stored
// and displayed on the console on the next screen refresh. Intermediate
// samples might get lost without being displayed.
func (sink *ConsoleBoxSink) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	sink.lock.Lock()
	sink.lastSample = sample
	sink.lastHeader = header
	if sink.ImmediateScreenUpdate {
		sink.TriggerUpdate()
	}
	sink.lock.Unlock()
	return sink.AbstractSampleOutput.Sample(nil, sample, header)
}
