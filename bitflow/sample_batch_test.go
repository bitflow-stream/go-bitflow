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

func TestSampleBatchTimeMixed(t *testing.T) {
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
	bStep := &TestBatchStep{
		assert:             assert.New(t),
		expectedBatchSizes: []int{6, 6, 1, 1},
		batchCounter:       0,
	}
	processor.Add(bStep)

	_test_mixed_batch(samples, processor, bStep, t)
}

func TestSampleBatchNumberMixed(t *testing.T) {
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
	bStep := &TestBatchStep{
		assert:             assert.New(t),
		expectedBatchSizes: []int{5, 5, 1, 1},
		batchCounter:       0,
	}
	processor.Add(bStep)

	_test_mixed_batch(samples, processor, bStep, t)
}
