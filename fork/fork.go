package fork

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
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
	var errors golib.MultiError
	for _, key := range keys {
		pipeline := f.getPipeline(f.Builder, key, f)
		if pipeline != nil {
			errors.Add(pipeline.Sample(sample, header))
		}
	}
	return errors.NilOrError()
}

func (f *MetricFork) String() string {
	return "Fork: " + f.Distributor.String()
}

func (f *MetricFork) ContainedStringers() []fmt.Stringer {
	return f.containedStringers(f.Builder)
}
