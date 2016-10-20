package pipeline

import (
	"math"

	"github.com/antongulenko/go-bitflow"
)

type MinMaxScaling struct {
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

func IsValidNumber(val float64) bool {
	return !math.IsNaN(val) && !math.IsInf(val, 0)
}

func scaleMinMax(val, min, max float64) float64 {
	res := (val - min) / (max - min)
	if !IsValidNumber(res) {
		res = 0.5 // Zero standard-deviation -> pick the middle between 0 and 1
	}
	return res
}

func (s *MinMaxScaling) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	min, max := GetMinMax(header, samples)
	out := make([]*bitflow.Sample, len(samples))
	for num, inSample := range samples {
		values := make([]bitflow.Value, len(inSample.Values))
		for i, val := range inSample.Values {
			res := scaleMinMax(float64(val), min[i], max[i])
			values[i] = bitflow.Value(res)
		}
		var outSample bitflow.Sample
		outSample.Values = values
		outSample.CopyMetadataFrom(inSample)
		out[num] = &outSample
	}
	return header, out, nil
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

func (s *StandardizationScaling) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	stats := GetStats(header, samples)
	out := make([]*bitflow.Sample, len(samples))
	for num, inSample := range samples {
		values := make([]bitflow.Value, len(inSample.Values))
		for i, val := range inSample.Values {
			m := stats[i].Mean()
			s := stats[i].Stddev()
			res := (float64(val) - m) / s
			if !IsValidNumber(res) {
				// Special case for zero standard deviation: fallback to min-max scaling
				min, max := stats[i].Min, stats[i].Max
				res = scaleMinMax(float64(val), min, max)
				res = (res - 0.5) * 2 // Value range: -1..1
			}
			values[i] = bitflow.Value(res)
		}
		var outSample bitflow.Sample
		outSample.Values = values
		outSample.CopyMetadataFrom(inSample)
		out[num] = &outSample
	}
	return header, out, nil
}

func (s *StandardizationScaling) String() string {
	return "Standardization scaling"
}
