package bitflow

import (
	"sync"
	"testing"
	"time"

	"github.com/antongulenko/golib"
	"github.com/stretchr/testify/suite"
)

type BatchStepTestSuite struct {
	golib.AbstractTestSuite
}

func TestBatchStep(t *testing.T) {
	suite.Run(t, new(BatchStepTestSuite))
}

type testBatchStep struct {
	*BatchStepTestSuite
	expectedBatchSizes []int
	batchCounter       int
}

func (step *testBatchStep) ProcessBatch(_ *Header, samples []*Sample) (*Header, []*Sample, error) {
	step.Equal(step.expectedBatchSizes[step.batchCounter], len(samples))
	step.batchCounter += 1
	return nil, nil, nil
}

func (step *testBatchStep) String() string {
	return "testBatchStep"
}

func (s *BatchStepTestSuite) testBatch(processor *BatchProcessor, samples []Sample, step *testBatchStep) {
	processor.Start(&sync.WaitGroup{})
	for i := range samples {
		_ = processor.Sample(&samples[i], &Header{Fields: []string{"dummy"}})
	}
	processor.Close()
	s.Equal(len(step.expectedBatchSizes), step.batchCounter)
}

func (s *BatchStepTestSuite) testMixedBatch(samples []Sample, processor *BatchProcessor, step *testBatchStep) {
	refTime := samples[0].Time
	timeFuture := refTime.Add(time.Second * time.Duration(70))
	samples = append(samples, Sample{
		Time: timeFuture,
		tags: map[string]string{"test": "test", "flush": "flush"},
	})
	timeFuture = refTime.Add(time.Second * time.Duration(71))
	samples = append(samples, Sample{
		Time: timeFuture,
		tags: map[string]string{"test": "test", "flush": "flush1"},
	})
	s.testBatch(processor, samples, step)
}

func (s *BatchStepTestSuite) TestSampleBatchTimeMixed() {
	processor := &BatchProcessor{
		FlushTags:          []string{"test", "flush"},
		FlushSampleLag:     time.Second * time.Duration(30),
		FlushAfterTime:     time.Second * time.Duration(5),
		DontFlushOnClose:   false,
		ForwardImmediately: false,
	}

	samples := make([]Sample, 12)
	timeNow := time.Now()
	for i := 0; i < 12; i++ {
		timeFuture := timeNow.Add(time.Second * time.Duration(i))
		samples[i] = Sample{
			Values: []Value{1.0},
			Time:   timeFuture,
			tags:   map[string]string{"test": "test", "flush": "flush"},
		}
	}
	step := &testBatchStep{
		BatchStepTestSuite: s,
		expectedBatchSizes: []int{6, 6, 1, 1},
		batchCounter:       0,
	}
	processor.Add(step)

	s.testMixedBatch(samples, processor, step)
}

func (s *BatchStepTestSuite) TestSampleBatchNumberMixed() {
	processor := &BatchProcessor{
		FlushTags:            []string{"test", "flush"},
		FlushSampleLag:       time.Second * time.Duration(30),
		FlushAfterNumSamples: 5,
		DontFlushOnClose:     false,
		ForwardImmediately:   false,
	}

	samples := make([]Sample, 10)
	timeNow := time.Now()
	for i := 0; i < 10; i++ {
		timeFuture := timeNow.Add(time.Second * time.Duration(i))
		samples[i] = Sample{
			Values: []Value{1.0},
			Time:   timeFuture,
			tags:   map[string]string{"test": "test", "flush": "flush"},
		}
	}
	step := &testBatchStep{
		BatchStepTestSuite: s,
		expectedBatchSizes: []int{5, 5, 1, 1},
		batchCounter:       0,
	}
	processor.Add(step)

	s.testMixedBatch(samples, processor, step)
}
