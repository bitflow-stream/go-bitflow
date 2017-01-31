package bitflow

import (
	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/golib"
)

// SamplePipeline reads data from a source, pipes it through zero or more SampleProcessor
// instances and outputs the final Samples into a MetricSink.
// The job of the SamplePipeline is to connect all the processing steps
// in the Construct method. After calling Construct, the SamplePipeline should not
// used any further.
type SamplePipeline struct {
	Source     MetricSource
	Processors []SampleProcessor
	Sink       MetricSink
}

// Construct connects the MetricSource, all SampleProcessors and the MetricSink
// inside the receiving SamplePipeline. It adds all these golib.Task instances
// to the given golib.TaskGroup. Afterwards, tasks.WaitAndStop() cann be called
// to start the entire pipeline. At least the Source and the Sink field
// must be set in the pipeline. The Construct method will not fail without them,
// but starting the resulting pipeline will usually not work.
func (p *SamplePipeline) Construct(tasks *golib.TaskGroup) {
	// First connect all sources with their sinks
	source := p.Source
	for _, processor := range p.Processors {
		if source != nil {
			source.SetSink(processor)
		}
		source = processor
	}
	if source != nil {
		source.SetSink(p.Sink)
	}

	// Then add all tasks in reverse: start the final sink first.
	// Each sink must be started before the source can push data into it.
	tasks.Add(p.Sink)
	for i := len(p.Processors) - 1; i >= 0; i-- {
		tasks.Add(p.Processors[i])
	}
	tasks.Add(p.Source)
}

// Add adds the SampleProcessor parameter to the list of SampleProcessors in the
// receiving SamplePipeline. The Source and Sink fields must be accessed directly.
// The Processors field can also be accessed directly, but the Add method allows
// chaining multiple Add invokations like so:
//   pipeline.Add(processor1).Add(processor2)
func (p *SamplePipeline) Add(processor SampleProcessor) *SamplePipeline {
	if processor != nil {
		p.Processors = append(p.Processors, processor)
	}
	return p
}

// ConfigureStandalone prints a warning if the sink or source of the pipeline
// are not set, and sets them to non-nil values. This can optionally be called after
// p.Configure().
func (p *SamplePipeline) ConfigureStandalone() {
	if p.Sink == nil {
		log.Warnln("No data sinks selected, data will not be output anywhere.")
		p.Sink = new(EmptyMetricSink)
	}
	if p.Source == nil {
		log.Warnln("No data source provided, no data will be received or generated.")
		p.Source = new(EmptyMetricSource)
	}
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
//   // ... Set p.Processors (Optional)
//   // ... Set p.Source and p.Sink using f.CreateSource() and f.CreateSink()
//   // p.ConfigureStandalone // (Optional)
//   os.Exit(p.StartAndWait()) // os.Exit() should be called in an outer method if 'defer' is used here
//
// An additional golib.Task is started along with the pipeline, which listens
// for the Ctrl-C user external interrupt and makes the pipeline stoppable cleanly
// by the user.
//
// StartAndWait returns the number of errors that occured in the pipeline.
func (p *SamplePipeline) StartAndWait() int {
	tasks := golib.NewTaskGroup()
	p.Construct(tasks)
	log.Debugln("Press Ctrl-C to interrupt")
	tasks.Add(&golib.NoopTask{
		Chan:        golib.ExternalInterrupt(),
		Description: "external interrupt",
	})
	return tasks.PrintWaitAndStop()
}