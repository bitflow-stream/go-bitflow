package main

import (
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	log "github.com/sirupsen/logrus"
)

type DataCollector struct {
	bitflow.AbstractSampleSource
	loopTask golib.LoopTask
	samples  chan bitflow.SampleAndHeader

	loopWaitTime         time.Duration
	bufferedSamples      int
	fatalForwardingError bool

	// TODO if necessary, add additional fields or command line parameters here
}

// Assert that the DataCollector type implements the necessary interface
var _ bitflow.SampleSource = new(DataCollector)

func NewDataCollector() *DataCollector {
	col := &DataCollector{
		// TODO Initialize default parameter values here

		loopWaitTime:         100 * time.Millisecond,
		bufferedSamples:      100,
		fatalForwardingError: false,
	}
	col.loopTask.Loop = col.loopIteration
	return col
}

func (col *DataCollector) RegisterFlags() {
	// TODO register additional command line flags here

	flag.DurationVar(&col.loopWaitTime, "loop", col.loopWaitTime, "Time to wait in the main sample generation loop")
	flag.IntVar(&col.bufferedSamples, "sample-buffer", col.bufferedSamples, "Number of samples to buffer between the sample generation and sample forwarding routine")
	flag.BoolVar(&col.fatalForwardingError, "fatal-forwarding-error", col.fatalForwardingError, "Shutdown when there is an error forwarding a generated sample")
}

func (col *DataCollector) Initialize(args []string) error {
	if len(args) > 0 {
		// TODO if required, handle positional non-flag arguments. Return a non-nil error for invalid arguments, or when no arguments are expected.
		return NoPositionalArgumentsExpected
	}

	// TODO perform initialization with parsed command line arguments

	col.samples = make(chan bitflow.SampleAndHeader, col.bufferedSamples)
	col.loopTask.Description = fmt.Sprintf("LoopTask of %v", col)
	return nil
}

func (col *DataCollector) String() string {

	// TODO return a descriptive string, including any relevant parameters

	return fmt.Sprintf("Example data collector skeleton (loop frequency: %v, buffered samples: %v, fatal forwarding errors: %v)",
		col.loopWaitTime, col.bufferedSamples, col.fatalForwardingError)
}

func (col *DataCollector) Start(wg *sync.WaitGroup) golib.StopChan {
	stopper := col.loopTask.Start(wg)
	wg.Add(1)
	// Start a parallel routine to output samples, to decouple sample generation from sample sending
	go col.parallelForwardSamples(wg)
	return stopper
}

func (col *DataCollector) loopIteration(stopper golib.StopChan) error {
	if err := col.generateSamples(); err != nil {
		return err
	}
	stopper.WaitTimeout(col.loopWaitTime)
	return nil
}

func (col *DataCollector) generateSamples() error {

	// TODO this is the place to actually generate an arbitrary number of *bitflow.Sample instances and send them via the emitSample() method
	// This method is called in a loop, so samples can be generated in regular intervals. Alternatively, this method can be long-running and generate samples over a long period of time.
	// If the header does not change, the *bitflow.Header instance should be reused (unlike in this example).
	// Returning a non-nil error will cause the program to terminate, so non-fatal errors should be logged instead. The number of header fields and sample values must match, while the number of tags is arbitrary.

	exampleHeader := &bitflow.Header{
		Fields: []string{"example-field1", "example-field2"},
	}
	exampleSample := &bitflow.Sample{
		Time:   time.Now(),
		Values: []bitflow.Value{1.0, 2.0},
	}
	exampleSample.SetTag("example-tag", "example-value")
	col.emitSample(exampleSample, exampleHeader)

	return nil
}

func (col *DataCollector) emitSample(sample *bitflow.Sample, header *bitflow.Header) {
	col.samples <- bitflow.SampleAndHeader{Sample: sample, Header: header}
}

func (col *DataCollector) Close() {
	// This method is called when the DataSource (including the parallel LoopTask) is asked to stop producing samples.
	// We stop our LoopTask, which will also command the fork parallelForwardSamples() to terminate.
	// Some cleanup tasks could be performed in this method, but it is recommended to do this in the cleanup() method.
	col.loopTask.Stop()
}

func (col *DataCollector) cleanup() {

	// TODO if necessary, perform additional cleanup routines and gracefully close resources
	// This method is called, after our last sample has been forwarded and all parallel routines have closed down.
	// Make sure to not panic here.

}

func (col *DataCollector) parallelForwardSamples(wg *sync.WaitGroup) {
	defer wg.Done()
	defer col.cleanup()
	defer col.CloseSinkParallel(wg)
	waitChan := col.loopTask.WaitChan()

	for {
		select {
		case <-waitChan:
			return
		case sample := <-col.samples:
			err := col.GetSink().Sample(sample.Sample, sample.Header)
			if err != nil {
				err = fmt.Errorf("Failed to forward data sample (number of values: %v): %v", len(sample.Sample.Values), err)
				if col.fatalForwardingError {
					col.loopTask.StopErr(err)
				} else {
					log.Errorln(err)
				}
			}
		}
	}
}
