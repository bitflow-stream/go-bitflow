package pipeline

import (
	"fmt"
	"regexp"
	"sort"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-onlinestats"
)

type MetricMapperHelper struct {
	Description string

	inHeader   *bitflow.Header
	outHeader  *bitflow.Header
	outIndices []int
}

func (helper *MetricMapperHelper) incomingHeader(header *bitflow.Header, constructIndices func(inHeader *bitflow.Header) ([]int, []string)) error {
	helper.inHeader = header
	var outFields []string
	helper.outIndices, outFields = constructIndices(header)
	if len(helper.outIndices) != len(outFields) {
		return fmt.Errorf("AbstractMetricChanger.ConstructIndices returned non equal sized results")
	}
	if len(outFields) == 0 {
		log.Warnln(helper.Description, "removed all metrics")
	} else {
		log.Println(helper.Description, "changes metrics", len(header.Fields), "->", len(outFields))
	}
	helper.outHeader = header.Clone(outFields)
	return nil
}

func (helper *MetricMapperHelper) convertSample(inSample *bitflow.Sample) *bitflow.Sample {
	outValues := make([]bitflow.Value, len(helper.outIndices))
	for i, index := range helper.outIndices {
		outValues[i] = inSample.Values[index]
	}
	outSample := inSample.Clone()
	outSample.Values = outValues
	return outSample
}

type AbstractMetricMapper struct {
	bitflow.AbstractProcessor
	Description      fmt.Stringer
	ConstructIndices func(inHeader *bitflow.Header) ([]int, []string)

	helper MetricMapperHelper
}

func (self *AbstractMetricMapper) Header(header *bitflow.Header) error {
	if err := self.CheckSink(); err != nil {
		return err
	}
	if !header.Equals(self.helper.inHeader) {
		self.helper = MetricMapperHelper{
			Description: self.String(),
		}
		if err := self.helper.incomingHeader(header, self.ConstructIndices); err != nil {
			return err
		}
	}
	return self.OutgoingSink.Header(self.helper.outHeader)
}

func (self *AbstractMetricMapper) Sample(inSample *bitflow.Sample, _ *bitflow.Header) error {
	if err := self.Check(inSample, self.helper.inHeader); err != nil {
		return err
	}
	outSample := self.helper.convertSample(inSample)
	return self.OutgoingSink.Sample(outSample, self.helper.outHeader)
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

func (self *AbstractMetricFilter) constructIndices(inHeader *bitflow.Header) ([]int, []string) {
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

func (filter *MetricFilter) MergeProcessor(other bitflow.SampleProcessor) bool {
	if otherFilter, ok := other.(*MetricFilter); !ok {
		return false
	} else {
		filter.exclude = append(filter.exclude, otherFilter.exclude...)
		filter.include = append(filter.include, otherFilter.include...)
		return true
	}
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

func (mapper *MetricMapper) constructIndices(inHeader *bitflow.Header) ([]int, []string) {
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

type BatchMetricMapper struct {
	Description      string
	ConstructIndices func(header *bitflow.Header, samples []*bitflow.Sample) ([]int, []string)
}

func (mapper *BatchMetricMapper) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	helper := &MetricMapperHelper{
		Description: mapper.String(),
	}
	constructIndices := func(_ *bitflow.Header) ([]int, []string) {
		return mapper.ConstructIndices(header, samples)
	}
	if err := helper.incomingHeader(header, constructIndices); err != nil {
		return nil, nil, err
	}
	outSamples := make([]*bitflow.Sample, len(samples))
	for i, inSample := range samples {
		outSample := helper.convertSample(inSample)
		outSamples[i] = outSample
	}
	return helper.outHeader, outSamples, nil
}

func (mapper *BatchMetricMapper) String() string {
	if mapper.Description == "" {
		return "Batch Metric Mapper"
	} else {
		return mapper.Description
	}
}

type MetricVarianceFilter struct {
	BatchMetricMapper
	MinimumWeightedStddev float64 // Stddev, relative to the mean. Could also use absolute stddev/variance.
}

func NewMetricVarianceFilter(minimumWeightedStddev float64) *MetricVarianceFilter {
	filter := &MetricVarianceFilter{
		MinimumWeightedStddev: minimumWeightedStddev,
	}
	filter.ConstructIndices = filter.constructIndices
	filter.Description = fmt.Sprintf("Metric Variance Filter (%.2f%%)", minimumWeightedStddev*100)
	return filter
}

func (filter *MetricVarianceFilter) constructIndices(header *bitflow.Header, samples []*bitflow.Sample) ([]int, []string) {
	numFields := len(header.Fields)
	variances := make([]onlinestats.Running, numFields)
	for _, sample := range samples {
		for i := range header.Fields {
			variances[i].Push(float64(sample.Values[i]))
		}
	}
	indices := make([]int, 0, numFields)
	fields := make([]string, 0, numFields)
	for i, field := range header.Fields {
		weighted_stddev := variances[i].Stddev()
		if mean := variances[i].Mean(); mean != 0 {
			weighted_stddev /= mean
		}
		if weighted_stddev >= filter.MinimumWeightedStddev {
			indices = append(indices, i)
			fields = append(fields, field)
		}
	}
	return indices, fields
}

type MetricRenamer struct {
	AbstractMetricMapper
	regexes      []*regexp.Regexp
	replacements []string
}

func NewMetricRenamer(regexes []*regexp.Regexp, replacements []string) *MetricRenamer {
	if len(regexes) != len(replacements) {
		panic(fmt.Sprintf("MetricRenamer: number of regexes does not match number of replacements (%v != %v)", len(regexes), len(replacements)))
	}
	renamer := &MetricRenamer{
		regexes:      regexes,
		replacements: replacements,
	}
	renamer.Description = renamer
	renamer.ConstructIndices = renamer.constructIndices
	return renamer
}

func (r *MetricRenamer) String() string {
	return fmt.Sprintf("Metric renamer (%v regexes)", len(r.regexes))
}

func (r *MetricRenamer) constructIndices(inHeader *bitflow.Header) ([]int, []string) {
	fields := make(indexedFields, len(inHeader.Fields))
	for i, field := range inHeader.Fields {
		for i, regex := range r.regexes {
			replace := r.replacements[i]
			field = regex.ReplaceAllString(field, replace)
		}
		fields[i].index = i
		fields[i].field = field
	}
	sort.Sort(fields)
	indices := make([]int, len(fields))
	outFields := make([]string, len(fields))
	for i, field := range fields {
		indices[i] = field.index
		outFields[i] = field.field
	}
	return indices, outFields
}

func (r *MetricRenamer) MergeProcessor(other bitflow.SampleProcessor) bool {
	if otherFilter, ok := other.(*MetricRenamer); !ok {
		return false
	} else {
		r.regexes = append(r.regexes, otherFilter.regexes...)
		r.replacements = append(r.replacements, otherFilter.replacements...)
		return true
	}
}

type indexedFields []struct {
	index int
	field string
}

func (f indexedFields) Len() int {
	return len(f)
}

func (f indexedFields) Less(i, j int) bool {
	return f[i].field < f[j].field
}

func (f indexedFields) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
