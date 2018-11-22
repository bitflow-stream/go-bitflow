package steps

import (
	"bufio"
	"bytes"
	"math"

	"github.com/antongulenko/go-onlinestats"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"gonum.org/v1/gonum/mat"
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
	s.Resize(len(values))
	for i, val := range values {
		s.Values[i] = bitflow.Value(val)
	}
}

func AppendToSample(s *bitflow.Sample, values []float64) {
	oldValues := s.Values
	l := len(s.Values)
	if !s.Resize(l + len(values)) {
		copy(s.Values, oldValues)
	}
	for i, val := range values {
		s.Values[l+i] = bitflow.Value(val)
	}
}

func IsValidNumber(val float64) bool {
	return !math.IsNaN(val) && !math.IsInf(val, 0)
}

func FillSampleFromMatrix(s *bitflow.Sample, row int, mat *mat.Dense) {
	FillSample(s, mat.RawRowView(row))
}

func FillSamplesFromMatrix(s []*bitflow.Sample, mat *mat.Dense) {
	for i, sample := range s {
		FillSampleFromMatrix(sample, i, mat)
	}
}

const _nul = rune(0)

func SplitShellCommand(s string) []string {
	scanner := bufio.NewScanner(bytes.NewBuffer([]byte(s)))
	scanner.Split(bufio.ScanRunes)
	var res []string
	var buf bytes.Buffer
	quote := _nul
	for scanner.Scan() {
		r := rune(scanner.Text()[0])
		flush := false
		switch quote {
		case _nul:
			switch r {
			case ' ', '\t', '\r', '\n':
				flush = true
			case '"', '\'':
				quote = r
				flush = true
			}
		case '"', '\'':
			if r == quote {
				flush = true
				quote = _nul
			}
		}

		if flush {
			if buf.Len() > 0 {
				res = append(res, buf.String())
				buf.Reset()
			}
		} else {
			buf.WriteRune(r)
		}
	}

	// Un-closed quotes are ignored
	if buf.Len() > 0 {
		res = append(res, buf.String())
	}
	return res
}

type FeatureStats struct {
	onlinestats.Running
	Min float64
	Max float64
}

func NewFeatureStats() *FeatureStats {
	return &FeatureStats{
		Min: math.MaxFloat64,
		Max: -math.MaxFloat64,
	}
}

func (stats *FeatureStats) Push(values ...float64) {
	for _, value := range values {
		stats.Running.Push(value)
		stats.Min = math.Min(stats.Min, value)
		stats.Max = math.Max(stats.Max, value)
	}
}

func (stats *FeatureStats) ScaleMinMax(val float64, outputMin, outputMax float64) float64 {
	return ScaleMinMax(val, stats.Min, stats.Max, outputMin, outputMax)
}

func (stats *FeatureStats) ScaleStddev(val float64) float64 {
	return ScaleStddev(val, stats.Mean(), stats.Stddev(), stats.Min, stats.Max)
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
	if !IsValidNumber(res) {
		// Special case for zero standard deviation: fallback to min-max scaling
		res = ScaleMinMax(float64(val), min, max, -1, 1)
	}
	return res
}

func ScaleMinMax(val, min, max, outputMin, outputMax float64) float64 {
	res := (val - min) / (max - min)
	if IsValidNumber(res) {
		// res is now in 0..1, transpose it within the range outputMin..outputMax
		res = res*(outputMax-outputMin) + outputMin
	} else {
		res = (outputMax + outputMin) / 2 // min == max -> pick the middle
	}
	return res
}
