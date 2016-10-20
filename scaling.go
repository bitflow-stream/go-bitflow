package pipeline

import (
	"math"

	"github.com/antongulenko/data2go"
)

type MinMaxScaling struct {
}

func GetMinMax(header *data2go.Header, samples []*data2go.Sample) ([]float64, []float64) {
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

func (s *MinMaxScaling) ProcessBatch(header *data2go.Header, samples []*data2go.Sample) (*data2go.Header, []*data2go.Sample, error) {
	min, max := GetMinMax(header, samples)
	out := make([]*data2go.Sample, len(samples))
	for num, inSample := range samples {
		values := make([]data2go.Value, len(inSample.Values))
		for i, val := range inSample.Values {
			res := scaleMinMax(float64(val), min[i], max[i])
			values[i] = data2go.Value(res)
		}
		var outSample data2go.Sample
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

func GetStats(header *data2go.Header, samples []*data2go.Sample) []FeatureStats {
	res := make([]FeatureStats, len(header.Fields))
	for _, sample := range samples {
		for i, val := range sample.Values {
			res[i].Push(float64(val))
		}
	}
	return res
}

func (s *StandardizationScaling) ProcessBatch(header *data2go.Header, samples []*data2go.Sample) (*data2go.Header, []*data2go.Sample, error) {
	stats := GetStats(header, samples)
	out := make([]*data2go.Sample, len(samples))
	for num, inSample := range samples {
		values := make([]data2go.Value, len(inSample.Values))
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
			values[i] = data2go.Value(res)
		}
		var outSample data2go.Sample
		outSample.Values = values
		outSample.CopyMetadataFrom(inSample)
		out[num] = &outSample
	}
	return header, out, nil
}

func (s *StandardizationScaling) String() string {
	return "Standardization scaling"
}
