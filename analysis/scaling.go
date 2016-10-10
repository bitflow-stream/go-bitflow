package analysis

import (
	"math"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/go-onlinestats"
)

type MinMaxScaling struct {
}

func GetMinMax(header *sample.Header, samples []*sample.Sample) ([]float64, []float64) {
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

func (s *MinMaxScaling) ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample, error) {
	min, max := GetMinMax(header, samples)
	out := make([]*sample.Sample, len(samples))
	for num, inSample := range samples {
		values := make([]sample.Value, len(inSample.Values))
		for i, val := range inSample.Values {
			total := max[i] - min[i]
			res := (float64(val) - min[i]) / total
			if math.IsNaN(res) || math.IsInf(res, 0) {
				res = float64(val)
			}
			values[i] = sample.Value(res)
		}
		var outSample sample.Sample
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

func GetStats(header *sample.Header, samples []*sample.Sample) []onlinestats.Running {
	res := make([]onlinestats.Running, len(header.Fields))
	for _, sample := range samples {
		for i, val := range sample.Values {
			res[i].Push(float64(val))
		}
	}
	return res
}

func (s *StandardizationScaling) ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample, error) {
	stats := GetStats(header, samples)
	out := make([]*sample.Sample, len(samples))
	for num, inSample := range samples {
		values := make([]sample.Value, len(inSample.Values))
		for i, val := range inSample.Values {
			m := stats[i].Mean()
			s := stats[i].Stddev()
			res := (float64(val) - m) / s
			values[i] = sample.Value(res)
		}
		var outSample sample.Sample
		outSample.Values = values
		outSample.CopyMetadataFrom(inSample)
		out[num] = &outSample
	}
	return header, out, nil
}

func (s *StandardizationScaling) String() string {
	return "Standardization scaling"
}
