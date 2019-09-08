package bitflow

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type TestBatchStep struct {
	assert             *assert.Assertions
	expectedBatchSizes []int
	batchCounter       int
}

func (tb *TestBatchStep) ProcessBatch(header *Header, samples []*Sample) (*Header, []*Sample, error) {
	tb.assert.Equal(tb.expectedBatchSizes[tb.batchCounter], len(samples))
	tb.batchCounter += 1
	return nil, nil, nil
}

func (tb *TestBatchStep) String() string {
	return "TestBatchStep"
}

func _do_test_batch(processor *BatchProcessor, samples []Sample, bStep *TestBatchStep, t *testing.T) {
	assert := assert.New(t)
	processor.Start(&sync.WaitGroup{})
	for i := range samples {
		_ = processor.Sample(&samples[i], &Header{Fields: []string{"dummy"}})
	}
	processor.Close()
	assert.Equal(len(bStep.expectedBatchSizes), bStep.batchCounter)
}

func _test_mixed_batch(samples []Sample, processor *BatchProcessor, bStep *TestBatchStep, t *testing.T) {
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
	_do_test_batch(processor, samples, bStep, t)
}

func _testSampleBatchTimeMixed(t *testing.T, handler WindowHandler, expectedBatchSizes []int) {
	processor := &BatchProcessor{
		FlushTags:          []string{"test", "flush"},
		FlushSampleLag:     time.Second * time.Duration(30),
		Handler:            handler,
		DontFlushOnClose:   false,
		ForwardImmediately: false,
	}

	samples := make([]Sample, 13)
	timeNow := time.Now()
	for i := 0; i < 13; i++ {
		timeFuture := timeNow.Add(time.Second * time.Duration(i))
		samples[i] = Sample{
			Values: []Value{1.0},
			Time:   timeFuture,
			tags:   map[string]string{"test": "test", "flush": "flush"},
		}
	}
	bStep := &TestBatchStep{
		assert:             assert.New(t),
		expectedBatchSizes: expectedBatchSizes,
		batchCounter:       0,
	}
	processor.Add(bStep)

	_test_mixed_batch(samples, processor, bStep, t)
}

func _testSampleBatchNumberMixed(t *testing.T, handler WindowHandler, expectedBatchSizes []int) {
	processor := &BatchProcessor{
		FlushTags:          []string{"test", "flush"},
		FlushSampleLag:     time.Second * time.Duration(30),
		Handler:            handler,
		DontFlushOnClose:   false,
		ForwardImmediately: false,
	}

	samples := make([]Sample, 11)
	timeNow := time.Now()
	for i := 0; i < 11; i++ {
		timeFuture := timeNow.Add(time.Second * time.Duration(i))
		samples[i] = Sample{
			Values: []Value{1.0},
			Time:   timeFuture,
			tags:   map[string]string{"test": "test", "flush": "flush"},
		}
	}
	bStep := &TestBatchStep{
		assert:             assert.New(t),
		expectedBatchSizes: expectedBatchSizes,
		batchCounter:       0,
	}
	processor.Add(bStep)

	_test_mixed_batch(samples, processor, bStep, t)
}

func TestSampleBatchNumberMixed(t *testing.T) {
	_testSampleBatchNumberMixed(t, &SampleWindowHandler{Size: 5}, []int{5, 5, 1, 1, 1})
	_testSampleBatchNumberMixed(t, &SampleWindowHandler{Size: 5, StepSize: 1}, []int{5, 5, 5, 5, 5, 5, 5, 1, 1})
	_testSampleBatchNumberMixed(t, &SampleWindowHandler{Size: 5, StepSize: 4}, []int{5, 5, 3, 1, 1})
	_testSampleBatchNumberMixed(t, &SampleWindowHandler{Size: 5, StepSize: 5}, []int{5, 5, 1, 1, 1})
	_testSampleBatchNumberMixed(t, &SampleWindowHandler{Size: 5, StepSize: 10}, []int{5, 1, 1, 1})
}

func TestSampleBatchTimeMixed(t *testing.T) {
	_testSampleBatchTimeMixed(t, &TimeWindowHandler{Size: 5 * time.Second}, []int{6, 6, 1, 1, 1})
	_testSampleBatchTimeMixed(t, &TimeWindowHandler{Size: 5 * time.Second, StepSize: 1 * time.Second}, []int{6, 6, 6, 6, 6, 6, 6, 6, 1, 1})
	_testSampleBatchTimeMixed(t, &TimeWindowHandler{Size: 5 * time.Second, StepSize: 4 * time.Second}, []int{6, 6, 5, 1, 1})
	_testSampleBatchTimeMixed(t, &TimeWindowHandler{Size: 5 * time.Second, StepSize: 5 * time.Second}, []int{6, 6, 1, 1, 1})
	_testSampleBatchTimeMixed(t, &TimeWindowHandler{Size: 5 * time.Second, StepSize: 10 * time.Second}, []int{6, 1, 1})
}

func TestBatchTagsAndTimeout(t *testing.T) {
	_testSampleBatchTimeMixed(t, &SimpleWindowHandler{}, []int{13, 1, 1})
	_testSampleBatchNumberMixed(t, &SimpleWindowHandler{}, []int{11, 1, 1})
}

// Without both

// Sample:
// 	full flush
//	Step size
//	step size > window size
