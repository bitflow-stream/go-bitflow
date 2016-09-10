package analysis

import (
	"fmt"
	"regexp"

	"github.com/antongulenko/data2go/sample"
)

type AbstractMetricFilter struct {
	AbstractProcessor
	IncludeFilter func(name string) bool // Return true if metric should be INcluded
	Description   fmt.Stringer

	inHeader   sample.Header
	outHeader  sample.Header
	outIndices []int
}

func (p *AbstractMetricFilter) Header(header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}

	p.inHeader = header
	outFields := make([]string, 0, len(header.Fields))
	p.outIndices = make([]int, 0, len(header.Fields))
	for index, field := range header.Fields {
		if p.IncludeFilter(field) {
			outFields = append(outFields, field)
			p.outIndices = append(p.outIndices, index)
		}
	}
	if len(outFields) == 0 {
		return fmt.Errorf("%v filtered out all metrics!", p)
	}
	p.outHeader = header.Clone(outFields)
	return p.OutgoingSink.Header(p.outHeader)
}

func (p *AbstractMetricFilter) Sample(inSample sample.Sample, _ sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := inSample.Check(p.inHeader); err != nil {
		return err
	}
	outValues := make([]sample.Value, len(p.outIndices))
	for i, index := range p.outIndices {
		outValues[i] = inSample.Values[index]
	}
	outSample := inSample.Clone()
	outSample.Values = outValues
	return p.OutgoingSink.Sample(outSample, p.outHeader)
}

func (p *AbstractMetricFilter) String() string {
	if desc := p.Description; desc == nil {
		return "Generic MetricFilter"
	} else {
		return desc.String()
	}
}

type MetricFilter struct {
	AbstractMetricFilter
	exclude []*regexp.Regexp
	include []*regexp.Regexp
}

func NewMetricFilter() *MetricFilter {
	res := new(MetricFilter)
	res.Description = res
	res.AbstractMetricFilter.IncludeFilter = res.filter
	return res
}

func (filter *MetricFilter) Exclude(regex *regexp.Regexp) *MetricFilter {
	filter.exclude = append(filter.exclude, regex)
	return filter
}

func (filter *MetricFilter) ExcludeStr(substr string) *MetricFilter {
	return filter.ExcludeRegex(regexp.QuoteMeta(substr))
}

func (filter *MetricFilter) ExcludeRegex(regexStr string) *MetricFilter {
	return filter.Exclude(regexp.MustCompile(regexStr))
}

func (filter *MetricFilter) Include(regex *regexp.Regexp) *MetricFilter {
	filter.include = append(filter.include, regex)
	return filter
}

func (filter *MetricFilter) IncludeStr(substr string) *MetricFilter {
	return filter.IncludeRegex(regexp.QuoteMeta(substr))
}

func (filter *MetricFilter) IncludeRegex(regexStr string) *MetricFilter {
	return filter.Include(regexp.MustCompile(regexStr))
}

func (filter *MetricFilter) filter(name string) bool {
	excluded := false
	for _, regex := range filter.exclude {
		if excluded = regex.MatchString(name); excluded {
			break
		}
	}
	if !excluded && len(filter.include) > 0 {
		excluded = true
		for _, regex := range filter.include {
			if excluded = !regex.MatchString(name); !excluded {
				break
			}
		}
	}
	return !excluded
}

func (p *MetricFilter) String() string {
	return fmt.Sprintf("MetricFilter(%v exclude filters, %v include filters)", len(p.exclude), len(p.include))
}

type SampleFilter struct {
	AbstractProcessor
	Description   string
	IncludeFilter func(inSample *sample.Sample) bool // Return true if sample should be INcluded
}

func (p *SampleFilter) Sample(inSample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := inSample.Check(header); err != nil {
		return err
	}
	if filter := p.IncludeFilter; filter != nil && filter(&inSample) {
		return p.OutgoingSink.Sample(inSample, header)
	} else {
		return nil
	}
}

func (p *SampleFilter) String() string {
	if p.Description == "" {
		return "Sample Filter"
	} else {
		return "Sample Filter: " + p.Description
	}
}
