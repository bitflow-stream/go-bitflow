package fork

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
)

type RemapDistributor interface {
	Distribute(forkPath [][]interface{}) []interface{}
	String() string
}

type ForkRemapper struct {
	AbstractMetricFork

	Distributor RemapDistributor
	Builder     PipelineBuilder
}

func (f *ForkRemapper) GetPipeline(forkPath [][]interface{}) bitflow.MetricSink {
	key := f.Distributor.Distribute(forkPath)
	return f.getPipeline(f.Builder, key, f)
}

// This is just the 'default' channel if this ForkRemapper is used like a regular SampleProcessor
func (f *ForkRemapper) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := f.Check(sample, header); err != nil {
		return err
	}
	return f.GetPipeline(nil).Sample(sample, header)
}

func (f *ForkRemapper) String() string {
	return "Remapping Fork: " + f.Distributor.String()
}

func (f *ForkRemapper) ContainedStringers() []fmt.Stringer {
	return f.containedStringers(f.Builder)
}
