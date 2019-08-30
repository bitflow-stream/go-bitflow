package math

import (
	"fmt"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

type Aggregator interface {
	Aggregate(header *bitflow.Header, samples []*bitflow.Sample, refSample *bitflow.Sample) (*bitflow.Sample, error)
	String() string
}

type SumAggregator struct{}

func (a *SumAggregator) Aggregate(header *bitflow.Header, samples []*bitflow.Sample, refSample *bitflow.Sample) (*bitflow.Sample, error){
	values := make([]bitflow.Value, len(header.Fields))
	for _, sample := range samples{
		for i, value := range sample.Values {
			values[i] += value
		}
	}
	sample := refSample.Clone()
	sample.Values = values
	return sample, nil
}

func (a *SumAggregator) String() string {
	return "Sum Aggregator"
}

type MultiplyAggregator struct{}

func (a *MultiplyAggregator) Aggregate(header *bitflow.Header, samples []*bitflow.Sample, refSample *bitflow.Sample) (*bitflow.Sample, error){
	values := make([]bitflow.Value, len(header.Fields))
	for _, sample := range samples{
		for i, value := range sample.Values {
			values[i] *= value
		}
	}
	sample := refSample.Clone()
	sample.Values = values
	return sample, nil
}

func (a *MultiplyAggregator) String() string {
	return "Multiply Aggregator"
}

type AverageAggregator struct{}

func (a *AverageAggregator) Aggregate(header *bitflow.Header, samples []*bitflow.Sample, refSample *bitflow.Sample) (*bitflow.Sample, error){
	values := make([]bitflow.Value, len(header.Fields))
	for _, sample := range samples{
		for i, value := range sample.Values {
			values[i] += value * (1.0 / bitflow.Value(len(samples)))
		}
	}
	sample := refSample.Clone()
	sample.Values = values
	return sample, nil
}

func (a *AverageAggregator) String() string {
	return "Average Aggregator"
}

type BatchAggregator struct {
	bitflow.NoopProcessor

	aggregator Aggregator
}

func (ba *BatchAggregator) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	samples = make([]*bitflow.Sample, 1)
	samples[0], _= ba.aggregator.Aggregate(header, samples, samples[len(samples) -1])
	return header, samples, nil
}

func (ba *BatchAggregator) String() string {
	return "Batch aggregation. Using " + ba.aggregator.String() + "."
}

func RegisterBatchAggregator(b reg.ProcessorRegistry) {
	b.RegisterBatchStep("aggregate",
		func(params map[string]interface{}) (bitflow.BatchProcessingStep, error) {
			aggregatorType := params["type"].(string)
			switch aggregatorType {
			case "sum":
				return &BatchAggregator{aggregator: &SumAggregator{}}, nil
			case "multiply":
				return &BatchAggregator{aggregator: &MultiplyAggregator{}}, nil
			case "avg":
				return &BatchAggregator{aggregator: &AverageAggregator{}}, nil
			default:
				validTypes := []string{"sum", "multiply","avg"}
				return nil, fmt.Errorf("Invalid aggregator type %v. Valid types: %v", aggregatorType, validTypes)
			}
		},
		"Aggregates sample batch to single sample by applying the operation defined by 'type' parameter respectively on each value").
		Required("type", reg.String())
}
