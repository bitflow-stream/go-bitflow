package steps

import (
	"fmt"
	"regexp"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

type MetricSplitter struct {
	NoopProcessor
	header bitflow.HeaderChecker
	splits []*_splitSample

	Splitters []*regexp.Regexp
}

type _splitSample struct {
	tags      map[string]string
	indices   []int
	outHeader *bitflow.Header
}

var MetricSplitterDescription = "Metrics that are matched by the regex will be converted to separate samples. When the regex contains named groups, their names and values will be added as tags, and an individual samples will be created for each unique value combination."

func RegisterMetricSplitter(b reg.ProcessorRegistry) {
	b.RegisterStep("split", func(p *bitflow.SamplePipeline, params map[string]string) error {
		splitter, err := NewMetricSplitter([]string{params["regex"]})
		if err == nil {
			p.Add(splitter)
		}
		return err
	}, MetricSplitterDescription, reg.RequiredParams("regex"))
}

func NewMetricSplitter(regexes []string) (*MetricSplitter, error) {
	result := &MetricSplitter{
		Splitters: make([]*regexp.Regexp, len(regexes)),
	}
	for i, regex := range regexes {
		compiled, err := regexp.Compile(regex)
		if err != nil {
			return nil, fmt.Errorf("Failed to compile regex %v, %v", i+1, regex)
		}
		result.Splitters[i] = compiled
	}
	return result, nil
}

func (m *MetricSplitter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	for _, out := range m.Split(sample, header) {
		if err := m.NoopProcessor.Sample(out.Sample, out.Header); err != nil {
			return err
		}
	}
	return nil
}

func (m *MetricSplitter) String() string {
	return fmt.Sprintf("Split metrics into separate samples: %v", m.Splitters)
}

func (m *MetricSplitter) Split(sample *bitflow.Sample, header *bitflow.Header) []bitflow.SampleAndHeader {
	if m.header.HeaderChanged(header) {
		m.prepare(header)
	}
	return m.split(sample, header)
}

func (m *MetricSplitter) prepare(header *bitflow.Header) {
	splits := make(map[string]*_splitSample)
	var orderedSplits []string
	defaultSplit := &_splitSample{
		tags:      make(map[string]string),
		outHeader: new(bitflow.Header),
	}

	for i, field := range header.Fields {
		matchedSplit := defaultSplit

		for _, regex := range m.Splitters {
			matches := regex.FindAllStringSubmatch(field, 1)
			if len(matches) > 0 {
				// This field will go into a separate sample. Do not check against further regexes.
				// Use the first match as reference to populate the tags of the split sample

				tags := make(map[string]string, len(regex.SubexpNames()))
				values := matches[0][1:]
				for i, name := range regex.SubexpNames()[1:] {
					tags[name] = values[i]
				}
				encoded := bitflow.EncodeTags(tags)

				split, ok := splits[encoded]
				if !ok {
					split = &_splitSample{
						tags:      tags,
						outHeader: new(bitflow.Header),
					}
					splits[encoded] = split
					orderedSplits = append(orderedSplits, encoded)
				}

				matchedSplit = split
				break
			}
		}

		matchedSplit.outHeader.Fields = append(matchedSplit.outHeader.Fields, field)
		matchedSplit.indices = append(matchedSplit.indices, i)
	}

	m.splits = m.splits[0:0]
	m.splits = append(m.splits, defaultSplit)
	for _, encoded := range orderedSplits {
		m.splits = append(m.splits, splits[encoded])
	}
}

func (m *MetricSplitter) split(sample *bitflow.Sample, header *bitflow.Header) []bitflow.SampleAndHeader {
	res := make([]bitflow.SampleAndHeader, len(m.splits))
	for i, split := range m.splits {
		outSample := sample.Clone()
		for key, value := range split.tags {
			outSample.SetTag(key, value)
		}
		values := make([]bitflow.Value, len(split.indices))
		for i, index := range split.indices {
			values[i] = sample.Values[index]
		}
		outSample.Values = values
		res[i] = bitflow.SampleAndHeader{outSample, split.outHeader}
	}
	return res
}
