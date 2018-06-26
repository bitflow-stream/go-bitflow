package fork

import (
	"fmt"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type ForkDistributor interface {
	Distribute(sample *bitflow.Sample, header *bitflow.Header) []interface{}
	String() string
}

type MetricFork struct {
	AbstractMetricFork

	ParallelClose bool
	Distributor   ForkDistributor
	Builder       PipelineBuilder
}

func (f *MetricFork) Start(wg *sync.WaitGroup) golib.StopChan {
	f.parallelClose = f.ParallelClose
	return f.AbstractMetricFork.Start(wg)
}

func (f *MetricFork) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	keys := f.Distributor.Distribute(sample, header)
	sink := f.getPipelines(f.Builder, keys, f)
	return sink.Sample(sample, header)
}

func (f *MetricFork) String() string {
	return "Fork: " + f.Distributor.String()
}

func (f *MetricFork) ContainedStringers() []fmt.Stringer {
	return f.containedStringers(f.Builder)
}
