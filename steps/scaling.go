package steps

import (
	"math"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"

	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
)

type MinMaxScaling struct {
	Min float64
	Max float64
}

func RegisterMinMaxScaling(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("scale_min_max",
		func(p *pipeline.SamplePipeline, params map[string]string) (err error) {
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
		func(p *pipeline.SamplePipeline) {
			p.Batch(new(StandardizationScaling))
		},
		"Normalize a batch of samples based on the mean and std-deviation",
		reg.SupportBatch())
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

func ScaleMinMax(val, min, max, outputMin, outputMax float64) float64 {
	res := (val - min) / (max - min)
	if pipeline.IsValidNumber(res) {
		// res is now in 0..1, transpose it within the range outputMin..outputMax
		res = res*(outputMax-outputMin) + outputMin
	} else {
		res = (outputMax + outputMin) / 2 // min == max -> pick the middle
	}
	return res
}

func (s *MinMaxScaling) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	min, max := GetMinMax(header, samples)
	for _, sample := range samples {
		for i, val := range sample.Values {
			res := ScaleMinMax(float64(val), min[i], max[i], s.Min, s.Max)
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

func ScaleStddev(val float64, mean, stddev, min, max float64) float64 {
	res := (val - mean) / stddev
	if !pipeline.IsValidNumber(res) {
		// Special case for zero standard deviation: fallback to min-max scaling
		res = ScaleMinMax(float64(val), min, max, -1, 1)
	}
	return res
}

func (s *StandardizationScaling) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	stats := GetStats(header, samples)
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
