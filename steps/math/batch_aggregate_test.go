package math

import (
	"testing"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
)

type BatchAggregateTestSuite struct {
	golib.AbstractTestSuite
}

func TestBatchAggregate(t *testing.T) {
	new(BatchAggregateTestSuite).Run(t)
}

func (s *BatchAggregateTestSuite) testAggregator(aggregator bitflow.BatchProcessingStep, header *bitflow.Header, samples []*bitflow.Sample, expectedValues []bitflow.Value) {
	outHeader, samples, err := aggregator.ProcessBatch(header, samples)
	s.NoError(err)
	s.Equal(header.Fields, outHeader.Fields)
	s.Len(samples, 1)
	s.Equal(expectedValues, samples[0].Values)
}

func (s *BatchAggregateTestSuite) getTestSamples() (*bitflow.Header, []*bitflow.Sample) {
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

func (s *BatchAggregateTestSuite) TestSumAggregator() {
	expectedValues := []bitflow.Value{9, 9, 9}
	header, samples := s.getTestSamples()
	s.testAggregator(NewBatchSumAggregator(), header, samples, expectedValues)
}

func (s *BatchAggregateTestSuite) TestMultiplyAggregator() {
	expectedValues := []bitflow.Value{20, 20, 27}
	header, samples := s.getTestSamples()
	s.testAggregator(NewBatchMultiplyAggregator(), header, samples, expectedValues)
}

func (s *BatchAggregateTestSuite) TestAvgAggregator() {
	expectedValues := []bitflow.Value{3, 3, 3}
	header, samples := s.getTestSamples()
	s.testAggregator(NewBatchAvgAggregator(), header, samples, expectedValues)
}

func (s *BatchAggregateTestSuite) TestMinAggregator() {
	expectedValues := []bitflow.Value{2, 2, 3}
	header, samples := s.getTestSamples()
	s.testAggregator(NewBatchMinAggregator(), header, samples, expectedValues)
}

func (s *BatchAggregateTestSuite) TestMaxAggregator() {
	expectedValues := []bitflow.Value{5, 5, 3}
	header, samples := s.getTestSamples()
	s.testAggregator(NewBatchMaxAggregator(), header, samples, expectedValues)
}
