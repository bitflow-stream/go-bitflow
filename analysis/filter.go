package analysis

import (
	"fmt"
	"regexp"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go/sample"
)

type AbstractMetricMapper struct {
	AbstractProcessor
	Description      fmt.Stringer
	ConstructIndices func(inHeader sample.Header) ([]int, []string)

	inHeader   sample.Header
	outHeader  sample.Header
	outIndices []int
}

func (self *AbstractMetricMapper) Header(header sample.Header) error {
	if err := self.CheckSink(); err != nil {
		return err
	}
	if !header.Equals(&self.inHeader) {
		self.inHeader = header
		var outFields []string
		if construct := self.ConstructIndices; construct == nil {
			return fmt.Errorf("AbstractMetricChanger.ConstructIndices must not be nil")
		} else {
			self.outIndices, outFields = construct(header)
		}
		if len(self.outIndices) != len(outFields) {
			return fmt.Errorf("AbstractMetricChanger.ConstructIndices returned non equal sized results")
		}
		if len(outFields) == 0 {
			log.Warnf("%v removed all metrics", self)
		}
		self.outHeader = header.Clone(outFields)
	}
	return self.OutgoingSink.Header(self.outHeader)
}

func (self *AbstractMetricMapper) Sample(inSample sample.Sample, _ sample.Header) error {
	if err := self.Check(inSample, self.inHeader); err != nil {
		return err
	}
	outValues := make([]sample.Value, len(self.outIndices))
	for i, index := range self.outIndices {
		outValues[i] = inSample.Values[index]
	}
	outSample := inSample.Clone()
	outSample.Values = outValues
	return self.OutgoingSink.Sample(outSample, self.outHeader)
}

func (self *AbstractMetricMapper) String() string {
	if desc := self.Description; desc == nil {
		return "Abstract Metric Mapper"
	} else {
		return desc.String()
	}
}

type AbstractMetricFilter struct {
	AbstractMetricMapper
	IncludeFilter func(name string) bool // Return true if metric should be INcluded
}

func (self *AbstractMetricFilter) constructIndices(inHeader sample.Header) ([]int, []string) {
	outFields := make([]string, 0, len(inHeader.Fields))
	outIndices := make([]int, 0, len(inHeader.Fields))
	filter := self.IncludeFilter
	if filter == nil {
		return nil, nil
	}
	for index, field := range inHeader.Fields {
		if filter(field) {
			outFields = append(outFields, field)
			outIndices = append(outIndices, index)
		}
	}
	return outIndices, outFields
}

type MetricFilter struct {
	AbstractMetricFilter
	exclude []*regexp.Regexp
	include []*regexp.Regexp
}

func NewMetricFilter() *MetricFilter {
	res := new(MetricFilter)
	res.Description = res
	res.ConstructIndices = res.constructIndices
	res.IncludeFilter = res.filter
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

type MetricMapper struct {
	AbstractMetricMapper
	Metrics []string
}

func NewMetricMapper(metrics []string) *MetricMapper {
	mapper := &MetricMapper{
		Metrics: metrics,
	}
	mapper.ConstructIndices = mapper.constructIndices
	mapper.Description = mapper
	return mapper
}

func (mapper *MetricMapper) constructIndices(inHeader sample.Header) ([]int, []string) {
	fields := make([]int, 0, len(mapper.Metrics))
	metrics := make([]string, 0, len(mapper.Metrics))
	for _, metric := range mapper.Metrics {
		found := false
		for field, inMetric := range inHeader.Fields {
			if metric == inMetric {
				fields = append(fields, field)
				metrics = append(metrics, metric)
				found = true
				break
			}
		}
		if !found {
			log.Warnf("%v: metric %v not found", mapper, metric)
		}
	}
	return fields, metrics
}

func (mapper *MetricMapper) String() string {
	maxlen := 3
	if len(mapper.Metrics) > maxlen {
		return fmt.Sprintf("Metric Mapper: %v ...", mapper.Metrics[:maxlen])
	} else {
		return fmt.Sprintf("Metric Mapper: %v", mapper.Metrics)
	}
}

type SampleFilter struct {
	AbstractProcessor
	Description   string
	IncludeFilter func(inSample *sample.Sample) bool // Return true if sample should be INcluded
}

func (p *SampleFilter) Sample(inSample sample.Sample, header sample.Header) error {
	if err := p.Check(inSample, header); err != nil {
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
