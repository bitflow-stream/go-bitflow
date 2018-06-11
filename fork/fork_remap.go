package fork

import (
	"fmt"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type RemapDistributor interface {
	Distribute(forkPath []interface{}) []interface{}
	String() string
}

type ForkRemapper struct {
	AbstractMetricFork

	ParallelClose bool
	Distributor   RemapDistributor
	Builder       PipelineBuilder
}

func (f *ForkRemapper) Start(wg *sync.WaitGroup) golib.StopChan {
	f.parallelClose = f.ParallelClose
	f.newPipelineHandler = func(sink bitflow.SampleProcessor) bitflow.SampleProcessor {
		// Synchronize writing, because multiple incoming pipelines can write to one pipeline
		return &ForkMerger{outgoing: sink}
	}
	return f.AbstractMetricFork.Start(wg)
}

func (f *ForkRemapper) GetMappedSink(forkPath []interface{}) bitflow.SampleProcessor {
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

func (f *ForkRemapper) ContainedStringers() []fmt.Stringer {
	return f.containedStringers(f.Builder)
}
