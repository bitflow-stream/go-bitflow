package bitflow

import (
	"errors"
	"fmt"

	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

// SamplePipeline reads data from a source and pipes it through zero or more SampleProcessor instances.
// The job of the SamplePipeline is to connect all the processing steps
// in the Construct method. After calling Construct, the SamplePipeline should not
// used any further.
type SamplePipeline struct {
	Source     SampleSource
	Processors []SampleProcessor

	lastProcessor SampleProcessor
}

// Construct connects the SampleSource and all SampleProcessors.
// It adds small wrapping golib.StoppableTask instances
// to the given golib.TaskGroup. Afterwards, tasks.WaitAndStop() can be called
// to start the entire pipeline. If the Source field is missing, it will be
// replaced with a new EmptySampleSource instance. nil values in the Processors
// field will be ignored. A new instance of DroppingSampleProcessor is added to the
// list of Processors to ensure that every step has a valid subsequent step.
//
// Additionally, all SampleProcessor instances will be wrapped in small wrapper objects
// that ensure that the samples and headers forwarded between the processors are consistent.
func (p *SamplePipeline) Construct(tasks *golib.TaskGroup) {
	firstSource := p.Source
	if firstSource == nil {
		firstSource = new(EmptySampleSource)
	}

	// First connect all sources with their sinks
	source := firstSource
	for _, processor := range p.Processors {
		if processor != nil {
			if resizingProcessor, ok := processor.(ResizingSampleProcessor); ok {
				wrapper := &resizingProcessorWrapper{sinkWrapper{false}, resizingProcessor}
				processor = wrapper
			} else {
				wrapper := &processorWrapper{sinkWrapper{false}, processor}
				processor = wrapper
			}
			source.SetSink(processor)
			source = processor
		}
	}

	// Make sure every SampleProcessor has a non-nil sink
	lastSink := new(DroppingSampleProcessor)
	source.SetSink(&processorWrapper{sinkWrapper{true}, lastSink})

	// Then add all tasks in reverse: start the final processor first.
	// Each processor must be started before the source can push data into it.
	tasks.Add(&ProcessorTaskWrapper{lastSink})
	for i := len(p.Processors) - 1; i >= 0; i-- {
		proc := p.Processors[i]
		if proc != nil {
			tasks.Add(&ProcessorTaskWrapper{proc})
		}
	}
    tasks.Add(&SourceTaskWrapper{firstSource})
}

// Add adds the SampleProcessor parameter to the list of SampleProcessors in the
// receiving SamplePipeline. The Source field must be accessed directly.
// The Processors field can also be accessed directly, but the Add method allows
// chaining multiple Add invocations like so:
//   pipeline.Add(processor1).Add(processor2)
func (p *SamplePipeline) Add(processor SampleProcessor) *SamplePipeline {
	if p.lastProcessor != nil {
		if merger, ok := p.lastProcessor.(MergeableProcessor); ok {
			if merger.MergeProcessor(processor) {
				// Merge successful: drop the incoming step
				return p
			}
		}
	}
	p.lastProcessor = processor
	if processor != nil {
		p.Processors = append(p.Processors, processor)
	}
	return p
}

func (p *SamplePipeline) Batch(steps ...BatchProcessingStep) *SamplePipeline {
	batch := new(BatchProcessor)
	for _, step := range steps {
		batch.Add(step)
	}
	return p.Add(batch)
}

func (p *SamplePipeline) String() string {
	return "Pipeline"
}

func (p *SamplePipeline) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, 0, len(p.Processors)+2)
	if p.Source != nil {
		res = append(res, p.Source)
	}
	for _, proc := range p.Processors {
		res = append(res, proc)
	}
	return res
}

func (p *SamplePipeline) FormatLines() []string {
	printer := IndentPrinter{
		OuterIndent:  "│ ",
		InnerIndent:  "├─",
		CornerIndent: "└─",
		FillerIndent: "  ",
	}
	return printer.PrintLines(p)
}

// StartAndWait constructs the pipeline and starts it. It blocks until the pipeline
// is finished. The Sink and Source fields must be set to non-nil values, for example
// using Configure* methods or setting the fields directly.
//
// The sequence of operations to start a SamplePipeline should roughly follow the following example:
//   // ... Define additional flags using the "flag" package (Optional)
//   var p sample.SamplePipeline
//   var f EndpointFactory
//   f.RegisterFlags()
//   flag.Parse()
//   // ... Modify f.Flag* values (Optional)
//   defer golib.ProfileCpu()() // (Optional)
//   // ... Set p.Processors (Optional, e.g. using f.CreateSink())
//   // ... Set p.Source using f.CreateSource()
//   os.Exit(p.StartAndWait()) // os.Exit() should be called in an outer method if 'defer' is used here
//
// An additional golib.Task is started along with the pipeline, which listens
// for the Ctrl-C user external interrupt and makes the pipeline stoppable cleanly
// by the user.
//
// StartAndWait returns the number of errors that occurred in the pipeline.
func (p *SamplePipeline) StartAndWait(extraTasks ...golib.Task) int {
	var tasks golib.TaskGroup
	p.Construct(&tasks)
	log.Debugln("Press Ctrl-C to interrupt")
	tasks.Add(&golib.NoopTask{
		Chan:        golib.ExternalInterrupt(),
		Description: "external interrupt",
	})
	tasks.Add(extraTasks...)
	return tasks.PrintWaitAndStop()
}

// ProcessorTaskWrapper can be used to convert an instance of SampleProcessor to a golib.Task.
// The Stop() method of the resulting Task is ignored.
type ProcessorTaskWrapper struct {
	SampleProcessor
}

// Stop implements the golib.Task interface. Calls to this Stop() method are ignored,
// because SampleProcessor instances should be shutdown through the Close() method.
func (t *ProcessorTaskWrapper) Stop() {
	// Ignore Stop() method
}

// SourceTaskWrapper can be used to convert an instance of SampleSource to a golib.Task.
// Calls to the Stop() method are mapped to the Close() method of the underlying SampleSource.
type SourceTaskWrapper struct {
	SampleSource
}

func (t *SourceTaskWrapper) Stop() {
	t.Close()
}

type processorWrapper struct {
	sinkWrapper
	SampleProcessor
}

func (p *processorWrapper) Sample(sample *Sample, header *Header) error {
	return p.forwardSample(p.SampleProcessor, sample, header)
}

type resizingProcessorWrapper struct {
	sinkWrapper
	ResizingSampleProcessor
}

func (p *resizingProcessorWrapper) Sample(sample *Sample, header *Header) error {
	return p.forwardSample(p.ResizingSampleProcessor, sample, header)
}

type sinkWrapper struct {
	dropSamples bool
}

func (w *sinkWrapper) forwardSample(p SampleProcessor, sample *Sample, header *Header) error {
	if w.dropSamples {
		return nil
	}
	if p.GetSink() == nil {
		return fmt.Errorf("No data sink set for %v", p)
	}
	if sample == nil {
		return errors.New("The sample is nil")
	}
	if header == nil {
		return errors.New("The header is nil")
	}
	if len(sample.Values) != len(header.Fields) {
		return fmt.Errorf("Unexpected number of values in sample: %v, expected %v",
			len(sample.Values), len(header.Fields))
	}
	return p.Sample(sample, header)
}
