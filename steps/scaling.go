package steps

import (
	"math"

	"github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

type MinMaxScaling struct {
}

func RegisterMinMaxScaling(b *query.PipelineBuilder) {
	b.RegisterAnalysis("scale_min_max",
		func(p *pipeline.SamplePipeline) {
			p.Batch(new(MinMaxScaling))
		},
		"Normalize a batch of samples using a min-max scale")
}

func RegisterStandardizationScaling(b *query.PipelineBuilder) {
	b.RegisterAnalysis("standardize",
		func(p *pipeline.SamplePipeline) {
			p.Batch(new(StandardizationScaling))
		},
		"Normalize a batch of samples based on the mean and std-deviation")
}

func GetMinMax(header *bitflow.Header, samples []*bitflow.Sample) ([]float64, []float64) {
	min := make([]float64, len(header.Fields))
	max := make([]float64, len(header.Fields))
	for num := range header.Fields {
		min[num] = math.MaxFloat64
		max[num] = -math.MaxFloat64
	}
	for _, sample := range samples {
		for i, val := range sample.Values {
			min[i] = math.Min(min[i], float64(val))
			max[i] = math.Max(max[i], float64(val))
		}
	}
	return min, max
}

func scaleMinMax(val, min, max float64) float64 {
	res := (val - min) / (max - min)
	if !pipeline.IsValidNumber(res) {
		res = 0.5 // Zero standard-deviation -> pick the middle between 0 and 1
	}
	return res
}

func (s *MinMaxScaling) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	min, max := GetMinMax(header, samples)
	for _, sample := range samples {
		for i, val := range sample.Values {
			res := scaleMinMax(float64(val), min[i], max[i])
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

func GetStats(header *bitflow.Header, samples []*bitflow.Sample) []FeatureStats {
	res := make([]FeatureStats, len(header.Fields))
	for _, sample := range samples {
		for i, val := range sample.Values {
			res[i].Push(float64(val))
		}
	}
	return res
}

func scaleStddev(val float64, stats FeatureStats) float64 {
	m, s := stats.Mean(), stats.Stddev()
	res := (val - m) / s
	if !pipeline.IsValidNumber(res) {
		// Special case for zero standard deviation: fallback to min-max scaling
		min, max := stats.Min, stats.Max
		res = scaleMinMax(float64(val), min, max)
		res = (res - 0.5) * 2 // Value range: -1..1
	}
	return res
}

func (s *StandardizationScaling) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	stats := GetStats(header, samples)
	for _, sample := range samples {
		for i, val := range sample.Values {
			res := scaleStddev(float64(val), stats[i])
			sample.Values[i] = bitflow.Value(res)
		}
	}
	return header, samples, nil
}

func (s *StandardizationScaling) String() string {
	return "Standardization scaling"
}
