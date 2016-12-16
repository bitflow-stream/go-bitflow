package bitflow

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
)

const (
	tcp_download_retry_interval = 1000 * time.Millisecond
)

// CmdSamplePipeline is an extension for SamplePipeline that defines many command line flags
// and additional methods for controlling the pipeline. It is basically a helper type to be
// reused in different main packages. All fields named Flag* are set by the according command
// line flags and evaluated in Init(). FlagInputs is not set by command line flags automatically.
// After flag.Parse(), those fields can be modified to override the command line flags defined by the user.
//
// The sequence of operations on a CmdSamplePipeline should follow the following example:
//   // ... Define additional flags using the "flag" package (Optional)
//   var p sample.CmdSamplePipeline
//   p.ParseFlags()
//   flag.Parse()
//   p.FlagInputs = flag.Args() // Or other value, optional
//   defer golib.ProfileCpu()() // Optional
//   p.SetSource(customSource) // Optional
//   p.ReadSampleHandler = customHandler // Optional
//   // ... Modify p.Flag* values // Optional
//   p.Init()
//   p.Tasks.Add(customTask) // Optional
//   os.Exit(p.StartAndWait())
//
type CmdSamplePipeline struct {
	// SamplePipeline in the CmdSamplePipeline should not be accessed directly,
	// except for the Add method. The Sink and Source fields should only be read,
	// because they are set in the Init() method.
	SamplePipeline

	// Tasks will be initialized in Init() and should be accessed before
	// calling StartAndWait(). It can be used to add additional golib.Task instances
	// that should be started along with the pipeline in StartAndWait().
	// The pipeline tasks will be added in StartAndWait(), so the additional tasks
	// will be started before the pipeline.
	Tasks *golib.TaskGroup

	// ReadSampleHandler will be set to the Handler field in the SampleReader that is created in Init().
	// It can be used to modify Samples and Headers directly after they are read
	// by the SampleReader, e.g. directly after reading a Header from file, or directly
	// after receiving a Sample over a TCP connection. The main purpose of this
	// is that the ReadSampleHandler receives a string representation of the data
	// source, which can be used as a tag in the Samples. By default, the data source is not
	// stored in the Samples and this information will be lost one the Sample enters the pipeline.
	//
	// The data source string differs depending on the MetricSource used. The FileSource will
	// use the file name, while the TCPSource will use the remote TCP endpoint.
	//
	// Must be set before calling Init().
	ReadSampleHandler ReadSampleHandler

	EndpointFactory
}

// Init initialized the receiving CmdSamplePipeline by creating the Tasks field,
// evaluating all command line flags and configurations, and setting the Source
// and Sink fields in the contained SamplePipeline.
//
// The things to do before calling Init() is to set the ReadSampleHandler and FlagInputs fields
// and use the SetSource method, as those are evaluated in Init(). Both are optional.
// RegisterFlags() should also be called.
//
// Init must be called before accessing p.Tasks and before calling StartAndWait().
//
// See the documentation of the CmdSamplePipeline type.
func (p *CmdSamplePipeline) Init() {
	p.Tasks = golib.NewTaskGroup()

	// ====== Initialize sink(s)
	sinks, err := p.CreateOutput()
	if err != nil {
		log.Fatalln(err)
	}
	if len(sinks) == 0 {
		log.Warnln("No data sinks selected, data will not be output anywhere.")
	}
	p.Sink = sinks

	// ====== Initialize source(s)
	source, err := p.CreateInput(p.ReadSampleHandler)
	if err != nil {
		log.Fatalln(err)
	}
	if source == nil {
		log.Warnln("No data source provided, no data will be received or generated.")
		if len(sinks) == 0 && len(p.Processors) == 0 {
			log.Fatalln("No tasks defined")
		}
		source = new(EmptyMetricSource)
	}
	p.Source = source
}

// StartAndWait constructs the pipeline and starts it. It blocks until the pipeline
// is finished.
//
// An additional golib.Task is started along with the pipeline, which listens
// for the Ctrl-C user external interrupt and makes the pipeline stoppable cleanly
// when the user wants to.
//
// p.Tasks can be filled with additional Tasks before calling this.
//
// StartAndWait returns the number of errors that occured in the pipeline or in
// additional tasks added to the TaskGroup.
func (p *CmdSamplePipeline) StartAndWait() int {
	p.Construct(p.Tasks)
	log.Debugln("Press Ctrl-C to interrupt")
	p.Tasks.Add(&golib.NoopTask{
		Chan:        golib.ExternalInterrupt(),
		Description: "external interrupt",
	})
	return p.Tasks.PrintWaitAndStop()
}
