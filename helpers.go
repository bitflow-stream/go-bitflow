package pipeline

import (
	"bufio"
	"bytes"
	"math"
	"regexp"

	"github.com/antongulenko/go-bitflow"
	"github.com/gonum/matrix/mat64"
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

func FillSampleFromMatrix(s *bitflow.Sample, row int, mat *mat64.Dense) {
	FillSample(s, mat.RawRowView(row))
}

func FillSamplesFromMatrix(s []*bitflow.Sample, mat *mat64.Dense) {
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

func ResolveTagTemplate(template string, missingValues string, sample *bitflow.Sample) string {
	return TagTemplate{Template: template, MissingValue: missingValues}.BuildKey(sample)
}

type TagTemplate struct {
	Template     string // Placeholders like ${xxx} will be replaced by tag values (left empty if tag missing)
	MissingValue string // Replacement for missing values
}

var templateRegex = regexp.MustCompile("\\$\\{[^\\{]*\\}") // Example: ${hello}

func (t TagTemplate) BuildKey(sample *bitflow.Sample) string {
	return templateRegex.ReplaceAllStringFunc(t.Template, func(placeholder string) string {
		placeholder = placeholder[2 : len(placeholder)-1] // Strip the ${} prefix/suffix
		if sample.HasTag(placeholder) {
			return sample.Tag(placeholder)
		} else {
			return t.MissingValue
		}
	})
}
