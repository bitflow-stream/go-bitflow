package math

import (
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/steps"
)

type BatchAggregateFunc func(aggregated bitflow.Value, newValue bitflow.Value) bitflow.Value

type BatchAggregator struct {
	Aggregator  BatchAggregateFunc
	Description string
}

func (a *BatchAggregator) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	resultSample := samples[len(samples)-1].Clone()
	resultSample.Values = a.computeValues(header, samples)
	return header, []*bitflow.Sample{resultSample}, nil
}

func (a *BatchAggregator) computeValues(header *bitflow.Header, samples []*bitflow.Sample) []bitflow.Value {
	if len(samples) == 0 {
		return nil
	}

	// Start with the first sample
	values := make([]bitflow.Value, len(header.Fields))
	for i, value := range samples[0].Values {
		values[i] = value
	}

	for _, sample := range samples[1:] {
		for i, value := range sample.Values {
			values[i] = a.Aggregator(values[i], value)
		}
	}
	return values
}

func (a *BatchAggregator) String() string {
	return "Batch Aggregation: " + a.Description
}

func registerAggregator(b reg.ProcessorRegistry, operation string, factoryFunc func() bitflow.BatchProcessingStep) {
	b.RegisterBatchStep(operation,
		func(params map[string]interface{}) (bitflow.BatchProcessingStep, error) {
			return factoryFunc(), nil
		}, "Compute for all values in the batch (per metric): "+operation)
}

func RegisterBatchAggregators(b reg.ProcessorRegistry) {
	registerAggregator(b, "multiply", NewBatchMultiplyAggregator)
	registerAggregator(b, "sum", NewBatchSumAggregator)
}

func NewBatchMultiplyAggregator() bitflow.BatchProcessingStep {
	return &BatchAggregator{
		Aggregator: func(aggregated bitflow.Value, newValue bitflow.Value) bitflow.Value {
			return aggregated * newValue
		},
		Description: "multiply",
	}
}

func NewBatchSumAggregator() bitflow.BatchProcessingStep {
	return &BatchAggregator{
		Aggregator: func(aggregated bitflow.Value, newValue bitflow.Value) bitflow.Value {
			return aggregated + newValue
		},
		Description: "sum",
	}
}

type GetFeatureFunc func(stats steps.FeatureStats) float64

type BatchFeatureStatsAggregator struct {
	Aggregate   GetFeatureFunc
	Description string
}

func (ba *BatchFeatureStatsAggregator) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	resultSample := samples[len(samples)-1].Clone()
	resultSample.Values = ba.computeValues(header, samples)
	return header, []*bitflow.Sample{resultSample}, nil
}

func (ba *BatchFeatureStatsAggregator) computeValues(header *bitflow.Header, samples []*bitflow.Sample) []bitflow.Value {
	stats := steps.GetStats(header, samples)
	res := make([]bitflow.Value, len(header.Fields))
	for i, metricStats := range stats {
		res[i] = bitflow.Value(ba.Aggregate(metricStats))
	}
	return res
}

func (ba *BatchFeatureStatsAggregator) String() string {
	return "Batch aggregation: " + ba.Description
}

func RegisterBatchFeatureStatsAggregators(b reg.ProcessorRegistry) {
	registerAggregator(b, "avg", NewBatchAvgAggregator)
	registerAggregator(b, "stddev", NewBatchStddevAggregator)
	registerAggregator(b, "kurtosis", NewBatchKurtosisAggregator)
	registerAggregator(b, "variance", NewBatchVarianceAggregator)
	registerAggregator(b, "min", NewBatchMinAggregator)
	registerAggregator(b, "max", NewBatchMaxAggregator)
}

func NewBatchAvgAggregator() bitflow.BatchProcessingStep {
	return &BatchFeatureStatsAggregator{
		Aggregate:   func(stats steps.FeatureStats) float64 { return stats.Mean() },
		Description: "avg",
	}
}

func NewBatchStddevAggregator() bitflow.BatchProcessingStep {
	return &BatchFeatureStatsAggregator{
		Aggregate:   func(stats steps.FeatureStats) float64 { return stats.Stddev() },
		Description: "stddev",
	}
}

func NewBatchKurtosisAggregator() bitflow.BatchProcessingStep {
	return &BatchFeatureStatsAggregator{
		Aggregate:   func(stats steps.FeatureStats) float64 { return stats.Kurtosis() },
		Description: "kurtosis",
	}
}

func NewBatchVarianceAggregator() bitflow.BatchProcessingStep {
	return &BatchFeatureStatsAggregator{
		Aggregate:   func(stats steps.FeatureStats) float64 { return stats.Var() },
		Description: "variance",
	}
}

func NewBatchMinAggregator() bitflow.BatchProcessingStep {
	return &BatchFeatureStatsAggregator{
		Aggregate:   func(stats steps.FeatureStats) float64 { return stats.Min },
		Description: "min",
	}
}

func NewBatchMaxAggregator() bitflow.BatchProcessingStep {
	return &BatchFeatureStatsAggregator{
		Aggregate:   func(stats steps.FeatureStats) float64 { return stats.Max },
		Description: "max",
	}
}
