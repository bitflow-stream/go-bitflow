package bitflow

import (
	"io"
	"sync"
	"time"

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

	// CliLogBox provides configuration options for the command-line box
	gotermBox.CliLogBox

	// UpdateInterval gives the wait-period between screen-refresh cycles.
	UpdateInterval time.Duration

	updateTask *golib.LoopTask
	lock       sync.Mutex
	lastSample *Sample
	lastHeader *Header
}

// Should be called as early as possible to intercept all log messages
func (sink *ConsoleBoxSink) Init() {
	sink.CliLogBox.Init()
	sink.RegisterMessageHook()
}

// String implements the MetricSink interface.
func (sink *ConsoleBoxSink) String() string {
	return "ConsoleBoxSink"
}

// Start implements the MetricSink interface. It starts a goroutine
// that regularly refreshes the screen to display the current sample values
// and latest log output lines.
func (sink *ConsoleBoxSink) Start(wg *sync.WaitGroup) golib.StopChan {
	sink.InterceptLogger()
	log.Println("Printing samples to table")
	sink.updateTask = golib.NewErrLoopTask("", func(stop golib.StopChan) error {
		if err := sink.updateBox(); err != nil {
			return err
		}
		select {
		case <-time.After(sink.UpdateInterval):
		case <-stop:
		}
		return nil
	})
	sink.updateTask.StopHook = func() {
		sink.updateBox()
		sink.RestoreLogger()
	}
	return sink.updateTask.Start(wg)
}

func (sink *ConsoleBoxSink) updateBox() (err error) {
	sink.Update(func(out io.Writer, textWidth int) {
		sink.lock.Lock()
		sample := sink.lastSample
		header := sink.lastHeader
		sink.lock.Unlock()
		if sample != nil && header != nil {
			marshaller := &TextMarshaller{
				TextWidth: textWidth,
			}
			err = marshaller.WriteSample(sample, header, out)
		}
	})
	return
}

// Close implements the MetricSink interface. It stops the screen refresh
// goroutine.
func (sink *ConsoleBoxSink) Close() {
	sink.updateTask.Stop()
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
