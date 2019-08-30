package math

import (
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func _testAggregator(aggregator Aggregator, header *bitflow.Header, samples []*bitflow.Sample, expectedSample *bitflow.Sample, t *testing.T) {
	assert := assert.New(t)
	sample, err := aggregator.Aggregate(header, samples, samples[len(samples) - 1])
	assert.NoError(err)
	assert.Equal(expectedSample.Values, sample.Values)
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
	expectedSample := &bitflow.Sample{
		Values: []bitflow.Value{9,9,9},
		Time:   time.Time{},
	}
	header, samples := getTestSamples()
	_testAggregator(&SumAggregator{}, header, samples, expectedSample, t)
}

func TestMultiplyAggregator(t *testing.T) {
	expectedSample := &bitflow.Sample{
		Values: []bitflow.Value{20,20,27},
		Time:   time.Time{},
	}
	header, samples := getTestSamples()
	_testAggregator(&MultiplyAggregator{}, header, samples, expectedSample, t)
}

func TestAvgAggregator(t *testing.T) {
	expectedSample := &bitflow.Sample{
		Values: []bitflow.Value{3,3,3},
		Time:   time.Time{},
	}
	header, samples := getTestSamples()
	_testAggregator(&AverageAggregator{}, header, samples, expectedSample, t)
}
