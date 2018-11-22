package bitflow

import (
	"bytes"
	"fmt"
	"strings"
)

// String is a trivial implementation of the fmt.Stringer interface
type String string

func (s String) String() string {
	return string(s)
}

// ====================== Sorting ======================

type SortedStringers []fmt.Stringer

func (t SortedStringers) Len() int {
	return len(t)
}

func (t SortedStringers) Less(a, b int) bool {
	return t[a].String() < t[b].String()
}

func (t SortedStringers) Swap(a, b int) {
	t[a], t[b] = t[b], t[a]
}

type SortedStringPairs struct {
	Keys   []string
	Values []string
}

func (s *SortedStringPairs) FillFromMap(values map[string]string) {
	s.Keys = make([]string, 0, len(values))
	s.Values = make([]string, 0, len(values))
	for key, value := range values {
		s.Keys = append(s.Keys, key)
		s.Values = append(s.Values, value)
	}
}

func (s *SortedStringPairs) Len() int {
	return len(s.Keys)
}

func (s *SortedStringPairs) Less(i, j int) bool {
	return s.Keys[i] < s.Keys[j]
}

func (s *SortedStringPairs) Swap(i, j int) {
	s.Keys[i], s.Keys[j] = s.Keys[j], s.Keys[i]
	s.Values[i], s.Values[j] = s.Values[j], s.Values[i]
}

func (s *SortedStringPairs) String() string {
	var buf bytes.Buffer
	for i, key := range s.Keys {
		if i > 0 {
			buf.WriteString(" ")
		}
		fmt.Fprintf(&buf, "%v=%v", key, s.Values[i])
	}
	return buf.String()
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

type TitledSamplePipeline struct {
	*SamplePipeline
	Title string
}

func (t *TitledSamplePipeline) String() string {
	return t.Title
}
