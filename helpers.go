package pipeline

import (
	"fmt"
	"math"
	"strings"

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
	s.Resize(len(values))
	for i, val := range values {
		s.Values[i] = bitflow.Value(val)
	}
}

func IsValidNumber(val float64) bool {
	return !math.IsNaN(val) && !math.IsInf(val, 0)
}

// String is a trivial implementation of the fmt.Stringer interface
type String string

func (s String) String() string {
	return string(s)
}

// ====================== Printing ======================

type IndentPrinter struct {
	OuterIndent  string
	InnerIndent  string
	FillerIndent string
	CornerIndent string
}

type StringerContainer interface {
	ContainedStringers() []fmt.Stringer
}

func (p IndentPrinter) Print(obj fmt.Stringer) string {
	return strings.Join(p.PrintLines(obj), "\n")
}

func (p IndentPrinter) PrintLines(obj fmt.Stringer) []string {
	return p.printLines(obj, "", "")
}

func (p IndentPrinter) printLines(obj fmt.Stringer, headerIndent, childIndent string) []string {
	str := headerIndent
	if obj == nil {
		str += "(nil)"
	} else {
		str += obj.String()
	}
	res := []string{str}
	if container, ok := obj.(StringerContainer); ok {
		parts := container.ContainedStringers()
		if len(parts) == 0 {
			parts = append(parts, String("empty"))
		}
		for i, part := range parts {
			var nextHeader, nextChild string
			if i == len(parts)-1 {
				nextHeader = p.CornerIndent
				nextChild = p.FillerIndent
			} else {
				nextHeader = p.InnerIndent
				nextChild = p.OuterIndent
			}
			partLines := p.printLines(part, childIndent+nextHeader, childIndent+nextChild)
			res = append(res, partLines...)
		}
	}
	return res
}
