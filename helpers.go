package pipeline

import (
	"math"

	"github.com/antongulenko/go-bitflow"
)

func ValuesToVector(input []bitflow.Value) []float64 {
	values := make([]float64, len(input))
	for i, val := range input {
		values[i] = float64(val)
	}
	return values
}

func SampleToVector(sample *bitflow.Sample) []float64 {
	return ValuesToVector(sample.Values)
}

func FillSample(s *bitflow.Sample, values []float64) {
	if len(s.Values) >= len(values) {
		s.Values = s.Values[:len(values)]
	} else {
		s.Values = make([]bitflow.Value, len(values))
	}
	for i, val := range values {
		s.Values[i] = bitflow.Value(val)
	}
}

func IsValidNumber(val float64) bool {
	return !math.IsNaN(val) && !math.IsInf(val, 0)
}
