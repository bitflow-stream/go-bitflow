package fork

/*
import (
	"fmt"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type RemapDistributor interface {
	Distribute(forkPath []string) []string
	String() string
}

type ForkRemapper struct {
	AbstractMetricFork
	Distributor RemapDistributor
}

func (f *ForkRemapper) Start(wg *sync.WaitGroup) golib.StopChan {
	f.newPipelineHandler = func(sink bitflow.SampleProcessor) bitflow.SampleProcessor {
		// Synchronize writing, because multiple incoming pipelines can write to one pipeline
		return &Merger{outgoing: sink}
	}
	return f.AbstractMetricFork.Start(wg)
}

func (f *ForkRemapper) GetMappedSink(forkPath []string) bitflow.SampleProcessor {
	keys := f.Distributor.Distribute(forkPath)
	return f.getPipelines(f.Builder, keys, f)
}

// This is just the 'default' channel if this ForkRemapper is used like a regular SampleProcessor
func (f *ForkRemapper) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	return f.GetMappedSink(nil).Sample(sample, header)
}

func (f *ForkRemapper) String() string {
	return "Remapping Fork: " + f.Distributor.String()
}

func (f *AbstractMetricFork) getRemappedSink(pipeline *bitflow.SamplePipeline, forkPath []string) bitflow.SampleProcessor {
	if len(pipeline.Processors) > 0 {
		last := pipeline.Processors[len(pipeline.Processors)-1]
		if _, isFork := last.(abstractForkContainer); isFork {
			// If the last step is a fork, it will handle remapping on its own
			return nil
		}
	}
	return f.getRemappedSinkRecursive(f.GetSink(), forkPath)
}

func (f *AbstractMetricFork) getRemappedSinkRecursive(outgoing bitflow.SampleProcessor, forkPath []string) bitflow.SampleProcessor {
	switch outgoing := outgoing.(type) {
	case *ForkRemapper:
		// Ask follow-up ForkRemapper for the pipeline we should connect to
		return outgoing.GetMappedSink(forkPath)
	case *Merger:
		// If there are multiple layers of forks, we have to resolve the ForkMergers until we get the actual outgoing sink
		return f.getRemappedSinkRecursive(outgoing.GetOriginalSink(), forkPath)
	default:
		// No follow-up ForkRemapper could be found
		return nil
	}
}

func (f *AbstractMetricFork) initializePipeline() {
	// Special handling of ForkRemapper: automatically connect mapped pipelines
	pipe.Add(f.getRemappedSink(pipe, path))
	f.StartPipeline(pipe, func(isPassive bool, err error) {
		if f.NonfatalErrors {
			f.LogFinishedPipeline(isPassive, err, fmt.Sprintf("[%v]: Subpipeline %v", description, path))
		} else {
			f.Error(err)
		}
	})
}



type StringRemapDistributor struct {
	Mapping map[string]string
}

func (d *StringRemapDistributor) Distribute(forkPath []interface{}) []interface{} {
	input := ""
	for i, path := range forkPath {
		if i > 0 {
			input += " "
		}
		input += fmt.Sprintf("%v", path)
	}
	result, ok := d.Mapping[input]
	if !ok {
		result = ""
		d.Mapping[input] = result
		log.Warnf("[%v]: No mapping found for fork path '%v', mapping to default output", d, input)
	}
	return []interface{}{result}
}

func (d *StringRemapDistributor) String() string {
	return fmt.Sprintf("String remapper (len %v)", len(d.Mapping))
}

func  register() {
	b.RegisterFork("remap", fork_remap, "The remap-fork can be used after another fork to remap the incoming sub-pipelines to new outgoing sub-pipelines", nil)
}

func fork_remap(params map[string]string) (fmt.Stringer, error) {
	return &StringRemapDistributor{
		Mapping: params,
	}, nil
}

*/
