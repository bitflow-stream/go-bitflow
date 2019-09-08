package math

import (
	"fmt"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

type Aggregator interface {
	Aggregate(header *bitflow.Header, samples []*bitflow.Sample) []bitflow.Value
	Type() string
}

type SumAggregator struct {
	strType string
}

func (a *SumAggregator) Aggregate(header *bitflow.Header, samples []*bitflow.Sample) []bitflow.Value {
	values := make([]bitflow.Value, len(header.Fields))
	for _, sample := range samples {
		for i, value := range sample.Values {
			values[i] += value
		}
	}
	return values
}

func (a *SumAggregator) Type() string {
	return "sum"
}

type MultiplyAggregator struct {
	strType string
}

func (a *MultiplyAggregator) Aggregate(header *bitflow.Header, samples []*bitflow.Sample) []bitflow.Value {
	values := make([]bitflow.Value, len(header.Fields))
	for i, sample := range samples {
		for j, value := range sample.Values {
			if i == 0 {
				values[j] = value
			} else {
				values[j] *= value
			}
		}
	}
	return values
}

func (a *MultiplyAggregator) Type() string {
	return "multiply"
}

type AverageAggregator struct {
	strType string
}

func (a *AverageAggregator) Aggregate(header *bitflow.Header, samples []*bitflow.Sample) []bitflow.Value {
	values := make([]bitflow.Value, len(header.Fields))
	for _, sample := range samples {
		for i, value := range sample.Values {
			values[i] += value
		}
	}
	for i := range values {
		values[i] /= bitflow.Value(len(samples))
	}
	return values
}

func (a *AverageAggregator) Type() string {
	return "avg"
}

type SlopeAggregator struct {
	strType string
}

func (a *SlopeAggregator) Aggregate(header *bitflow.Header, samples []*bitflow.Sample) []bitflow.Value {
	values := make([]bitflow.Value, len(header.Fields))
	firstSample := samples[0]
	lastSample := samples[len(samples)-1]
	for i := range header.Fields {
		values[i] = lastSample.Values[i] - firstSample.Values[i]
	}
	return values
}

func (a *SlopeAggregator) Type() string {
	return "slope"
}

type BatchAggregator struct {
	aggregator   Aggregator
	appendFields bool
}

func (ba *BatchAggregator) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	resultSamples := make([]*bitflow.Sample, 1)
	resultHeader := header
	values := ba.aggregator.Aggregate(header, samples)

	referenceSample := samples[len(samples)-1]
	if ba.appendFields {
		sampleAndHeader := bitflow.SampleAndHeader{
			Sample: referenceSample,
			Header: header,
		}
		fields := make([]string, len(header.Fields))
		for i, field := range header.Fields {
			fields[i] = field + "_" + ba.aggregator.Type()
		}
		sampleAndHeader.AddFields(fields, values)
		resultSamples[0] = sampleAndHeader.Sample
		resultHeader = sampleAndHeader.Header
	} else {
		resultSamples[0] = referenceSample.DeepClone()
		resultSamples[0].Values = values
	}
	return resultHeader, resultSamples, nil
}

func (ba *BatchAggregator) String() string {
	return "Batch aggregation. Using " + ba.aggregator.Type() + " aggregator."
}

func RegisterBatchAggregator(b reg.ProcessorRegistry) {
	validTypes := []string{"sum", "multiply", "avg", "slope"}
	b.RegisterBatchStep("aggregate",
		func(params map[string]interface{}) (bitflow.BatchProcessingStep, error) {
			aggregatorType := params["type"].(string)
			appendField := params["append-field"].(bool)
			switch aggregatorType {
			case "sum":
				return &BatchAggregator{
					aggregator:   &SumAggregator{},
					appendFields: appendField,
				}, nil
			case "multiply":
				return &BatchAggregator{
					aggregator:   &MultiplyAggregator{},
					appendFields: appendField,
				}, nil
			case "avg":
				return &BatchAggregator{
					aggregator:   &AverageAggregator{},
					appendFields: appendField,
				}, nil
			case "slope":
				return &BatchAggregator{
					aggregator:   &SlopeAggregator{},
					appendFields: appendField,
				}, nil
			default:
				return nil, fmt.Errorf("Invalid aggregator type %v. Valid types: %v", aggregatorType, validTypes)
			}
		},
		"Aggregates sample batch to a single sample by applying the operation defined by 'type' parameter respectively on each value").
		Required("type", reg.String(), fmt.Sprintf("Aggregate operation. Valid operation types: %v", validTypes)).
		Optional("append-field", reg.Bool(), false, "Control whether the aggregation results should be appended to the "+
			"last sample of the batch (true) or replace the existing fields by the aggregation results (false).")
}
