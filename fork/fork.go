package fork

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
)

type ForkDistributor interface {
	Distribute(sample *bitflow.Sample, header *bitflow.Header) []interface{}
	String() string
}

type MetricFork struct {
	AbstractMetricFork

	Distributor ForkDistributor
	Builder     PipelineBuilder
}

func (f *MetricFork) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := f.Check(sample, header); err != nil {
		return err
	}
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
