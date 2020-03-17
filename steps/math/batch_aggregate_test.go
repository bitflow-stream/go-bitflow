package math

import (
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

func _testAggregator(aggregator bitflow.BatchProcessingStep, header *bitflow.Header, samples []*bitflow.Sample, expectedValues []bitflow.Value, t *testing.T) {
	assert := require.New(t)
	outHeader, samples, err := aggregator.ProcessBatch(header, samples)
	assert.NoError(err)
	assert.Equal(header.Fields, outHeader.Fields)
	assert.Len(samples, 1)
	assert.Equal(expectedValues, samples[0].Values)
}

func getTestSamples() (*bitflow.Header, []*bitflow.Sample) {
	samples := make([]*bitflow.Sample, 3)
	samples[0] = &bitflow.Sample{
		Values: []bitflow.Value{2, 5, 3},
		Time:   time.Time{},
	}
	samples[1] = &bitflow.Sample{
		Values: []bitflow.Value{2, 2, 3},
		Time:   time.Time{},
	}
	samples[2] = &bitflow.Sample{
		Values: []bitflow.Value{5, 2, 3},
		Time:   time.Time{},
	}
	return &bitflow.Header{Fields: []string{"test1", "test2", "test3"}}, samples
}

func TestSumAggregator(t *testing.T) {
	expectedValues := []bitflow.Value{9, 9, 9}
	header, samples := getTestSamples()
	_testAggregator(NewBatchSumAggregator(), header, samples, expectedValues, t)
}

func TestMultiplyAggregator(t *testing.T) {
	expectedValues := []bitflow.Value{20, 20, 27}
	header, samples := getTestSamples()
	_testAggregator(NewBatchMultiplyAggregator(), header, samples, expectedValues, t)
}

func TestAvgAggregator(t *testing.T) {
	expectedValues := []bitflow.Value{3, 3, 3}
	header, samples := getTestSamples()
	_testAggregator(NewBatchAvgAggregator(), header, samples, expectedValues, t)
}

func TestMinAggregator(t *testing.T) {
	expectedValues := []bitflow.Value{2, 2, 3}
	header, samples := getTestSamples()
	_testAggregator(NewBatchMinAggregator(), header, samples, expectedValues, t)
}

func TestMaxAggregator(t *testing.T) {
	expectedValues := []bitflow.Value{5, 5, 3}
	header, samples := getTestSamples()
	_testAggregator(NewBatchMaxAggregator(), header, samples, expectedValues, t)
}
