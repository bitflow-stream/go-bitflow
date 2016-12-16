package bitflow

import (
	"errors"

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
// but starting the resulting pipeline will not work.
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
	if agg, ok := p.Sink.(AggregateSink); ok {
		for _, sink := range agg {
			tasks.Add(sink)
		}
	} else {
		tasks.Add(p.Sink)
	}
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

// Configure fills the Sink and Source fields of the SamplePipeline using
// the given EndpointFactory and ReadSampleHandler. Configure calls ConfigureSource
// and ConfigureSink and returns the first error (if any).
func (p *SamplePipeline) Configure(f *EndpointFactory, handler ReadSampleHandler) error {
	if err := p.ConfigureSource(f, handler); err != nil {
		return err
	}
	if err := p.ConfigureSink(f); err != nil {
		return err
	}
	return nil
}

// ConfigureSink uses the given EndpointFactory to fill the Sink field of the
// SamplePipeline.
func (p *SamplePipeline) ConfigureSink(f *EndpointFactory) error {
	sinks, err := f.CreateOutput()
	if err != nil {
		return err
	}
	if len(sinks) == 0 {
		log.Warnln("No data sinks selected, data will not be output anywhere.")
	}
	p.Sink = sinks
	return nil
}

// ConfigureSource uses the given EndpointFactory to fill the Source field of the
// SamplePipeline.
//
// The ReadSampleHandler parmeter can be used to modify Samples and Headers directly after they are read
// by the SampleReader, e.g. directly after reading a Header from file, or directly
// after receiving a Sample over a TCP connection. The main purpose of this
// is that the ReadSampleHandler receives a string representation of the data
// source, which can be used as a tag in the Samples. By default, the data source is not
// stored in the Samples and this information will be lost one the Sample enters the pipeline.
// The data source string differs depending on the MetricSource used. The FileSource will
// use the file name, while the TCPSource will use the remote TCP endpoint.
//
// The ReadSampleHandler parameter can be nil.
func (p *SamplePipeline) ConfigureSource(f *EndpointFactory, handler ReadSampleHandler) error {
	source, err := f.CreateInput(handler)
	if err != nil {
		return err
	}
	if source == nil {
		log.Warnln("No data source provided, no data will be received or generated.")
		source = new(EmptyMetricSource)
	}
	p.Source = source
	return nil
}

// HasTasks returns a non-nil error if the receiving SamplePipeline has no data sink,
// no data source, and no processor defined. A non-nil error indicates that the
// pipeline will do nothing when started.
func (p *SamplePipeline) HasTasks() error {
	if len(p.Processors) > 0 {
		return nil
	}
	if p.Source != nil {
		return nil
	}
	if p.Sink != nil {
		if agg, ok := p.Sink.(AggregateSink); !ok && len(agg) > 0 {
			return nil
		}
	}
	return errors.New("No tasks defined")
}

// StartAndWait constructs the pipeline and starts it. It blocks until the pipeline
// is finished. The Sink and Source fields must be set to non-nil values, for example
// using Configure* methods or setting the fields directly.
//
// The sequence of operations to start a SamplePipeline should roughly follow the following example:
//   // ... Define additional flags using the "flag" package (Optional)
//   var p sample.SamplePipeline
//   var f EndpointFactory
//   f.RegisterAllFlags()
//   flag.Parse()
//   p.FlagInputs = flag.Args() // Or other value, optional
//   // ... Modify f.Flag* values // Optional
//   defer golib.ProfileCpu()() // Optional
//   err := p.Configure(&f, sampleHandler) // Or access p.Sink/p.Source directly. sampleHandler can be nil.
//   // ... Error handling
//   os.Exit(p.StartAndWait())
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
