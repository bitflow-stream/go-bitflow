package math

import (
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/steps"
)

type MinMaxScaling struct {
	Min float64
	Max float64
}

func RegisterMinMaxScaling(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("scale_min_max",
		func(p *bitflow.SamplePipeline, params map[string]string) (err error) {
			p.Batch(&MinMaxScaling{
				Min: reg.FloatParam(params, "min", 0, true, &err),
				Max: reg.FloatParam(params, "max", 1, true, &err),
			})
			return
		},
		"Normalize a batch of samples using a min-max scale. The output value range is 0..1 by default, but can be customized.",
		reg.SupportBatch())
}

func RegisterStandardizationScaling(b reg.ProcessorRegistry) {
	b.RegisterAnalysis("standardize",
		func(p *bitflow.SamplePipeline) {
			p.Batch(new(StandardizationScaling))
		},
		"Normalize a batch of samples based on the mean and std-deviation",
		reg.SupportBatch())
}

func (s *MinMaxScaling) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	min, max := steps.GetMinMax(header, samples)
	for _, sample := range samples {
		for i, val := range sample.Values {
			res := steps.ScaleMinMax(float64(val), min[i], max[i], s.Min, s.Max)
			sample.Values[i] = bitflow.Value(res)
		}
	}
	return header, samples, nil
}

func (s *MinMaxScaling) String() string {
	return "Min-Max scaling"
}

type StandardizationScaling struct {
}

func (s *StandardizationScaling) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	stats := steps.GetStats(header, samples)
	for _, sample := range samples {
		for i, val := range sample.Values {
			res := stats[i].ScaleStddev(float64(val))
			sample.Values[i] = bitflow.Value(res)
		}
	}
	return header, samples, nil
}

func (s *StandardizationScaling) String() string {
	return "Standardization scaling"
}
