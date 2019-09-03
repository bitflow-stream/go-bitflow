package math

import (
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func _testAggregator(aggregator Aggregator, header *bitflow.Header, samples []*bitflow.Sample, expectedValues []bitflow.Value, t *testing.T) {
	assert := assert.New(t)
	values := aggregator.Aggregate(header, samples)
	assert.Equal(expectedValues, values)
}

func getTestSamples() (*bitflow.Header, []*bitflow.Sample) {
	samples := make([]*bitflow.Sample, 3)
	samples[0] = &bitflow.Sample{
		Values: []bitflow.Value{2,5,3},
		Time:   time.Time{},
	}
	samples[1] = &bitflow.Sample{
		Values: []bitflow.Value{2,2,3},
		Time:   time.Time{},
	}
	samples[2] = &bitflow.Sample{
		Values: []bitflow.Value{5,2,3},
		Time:   time.Time{},
	}
	return &bitflow.Header{Fields: []string{"test1", "test2", "test3"}}, samples
}

func TestSumAggregator(t *testing.T) {
	expectedValues := []bitflow.Value{9,9,9}
	header, samples := getTestSamples()
	_testAggregator(&SumAggregator{}, header, samples, expectedValues, t)
}

func TestMultiplyAggregator(t *testing.T) {
	expectedValues := []bitflow.Value{20,20,27}
	header, samples := getTestSamples()
	_testAggregator(&MultiplyAggregator{}, header, samples, expectedValues, t)
}

func TestAvgAggregator(t *testing.T) {
	expectedValues := []bitflow.Value{3,3,3}
	header, samples := getTestSamples()
	_testAggregator(&AverageAggregator{}, header, samples, expectedValues, t)
}
