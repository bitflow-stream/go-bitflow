package bitflow

import (
	"io"
	"sync"

	"github.com/antongulenko/golib"
)

// A SampleSink receives samples and headers to do arbitrary operations on them.
// The usual interface for this is SampleProcessor, but sometimes this simpler interface
// is useful.
type SampleSink interface {
	Sample(sample *Sample, header *Header) error
}

// SynchronizingSampleSink is a SampleSink implementation that allows multiple
// goroutines to write data to the same sink and synchronizes these writes through a mutex.
type SynchronizingSampleSink struct {
	OutgoingSink SampleProcessor
	mutex        sync.Mutex
}

// Sample implements the SampleSink interface.
func (s *SynchronizingSampleSink) Sample(sample *Sample, header *Header) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.OutgoingSink.Sample(sample, header)
}

// ResizingSampleProcessor is a helper interface that can be implemented by SampleProcessors
// in order to make RequiredValues() more reliable. The result of
// the OutputSampleSize() method should give a worst-case estimation of the number of values
// that will be present in Samples after this SampleProcessor is done processing a sample.
// This allows the optimization of pre-allocating a value array large enough to hold the final
// amount of metrics.
// The optimization works best when all samples are processed in a one-to-one fashion,
// i.e. no samples are split into multiple samples.
type ResizingSampleProcessor interface {
	SampleProcessor
	OutputSampleSize(sampleSize int) int
}

// RequiredValues the number of Values that should be large enough to hold
// the end-result after processing a Sample by all intermediate SampleProcessors.
// The result is based on ResizingSampleProcessor.OutputSampleSize(). SampleProcessor instances
// that do not implement the ResizingSampleProcessor interface are assumed to not increase the
// number metrics.
func RequiredValues(numFields int, sink SampleSink) int {
	for {
		if sink == nil {
			break
		}
		if sink, ok := sink.(ResizingSampleProcessor); ok {
			newSize := sink.OutputSampleSize(numFields)
			if newSize > numFields {
				numFields = newSize
			}
		}
		if source, ok := sink.(SampleSource); ok {
			sink = source.GetSink()
		} else {
			break
		}
	}
	return numFields
}

// ==================== Configuration types ====================

// ParallelSampleHandler is a configuration type that is included in
// SampleReader and SampleWriter. Both the reader and writer can marshall
// and unmarshall Samples in parallel, and these routines are controlled
// through the two parameters in ParallelSampleHandler.
type ParallelSampleHandler struct {
	// BufferedSamples is the number of Samples that are buffered between the
	// marshall/unmarshall routines and the routine that writes/reads the input
	// or output streams.
	// The purpose of the buffer is, for example, to allow the routine reading a file
	// to read the data for multiple Samples in one read operation, which then
	// allows the parallel parsing routines to parse all the read Samples at the same time.
	// Setting BufferedSamples is a trade-off between memory consumption and
	// parallelism, but most of the time a value of around 1000 or so should be enough.
	// If this value is not set, no parallelism will be possible because
	// the channel between the cooperating routines will block on each operation.
	BufferedSamples int

	// ParallelParsers can be set to the number of goroutines that will be
	// used when marshalling or unmarshalling samples. These routines can
	// parallelize the parsing and marshalling operations. The most benefit
	// from the parallelism comes when reading samples from e.g. files, because
	// reading the file into memory can be decoupled from parsing Samples,
	// and multiple Samples can be parsed at the same time.
	//
	// This must be set to a value greater than zero, otherwise no goroutines
	// will be started.
	ParallelParsers int
}

// ==================== Internal types ====================

type parallelSampleStream struct {
	err     golib.MultiError
	errLock sync.Mutex

	wg     sync.WaitGroup
	closed golib.StopChan
}

func (state *parallelSampleStream) addError(err error) bool {
	if err != nil {
		state.errLock.Lock()
		defer state.errLock.Unlock()
		state.err.Add(err)
		return true
	}
	return false
}

func (state *parallelSampleStream) hasError() bool {
	state.errLock.Lock()
	defer state.errLock.Unlock()
	if len(state.err) > 0 {
		for _, err := range state.err {
			if err != io.EOF {
				return true
			}
		}
	}
	return false
}

func (state *parallelSampleStream) getErrorNoEOF() error {
	state.errLock.Lock()
	defer state.errLock.Unlock()
	var result golib.MultiError
	if len(state.err) > 0 {
		for _, err := range state.err {
			if err != io.EOF {
				result.Add(err)
			}
		}
	}
	return result.NilOrError()
}

type bufferedSample struct {
	stream   *parallelSampleStream
	data     []byte
	sample   *Sample
	done     bool
	doneCond *sync.Cond
}

func (sample *bufferedSample) waitDone() {
	sample.doneCond.L.Lock()
	defer sample.doneCond.L.Unlock()
	for !sample.done {
		sample.doneCond.Wait()
	}
}

func (sample *bufferedSample) notifyDone() {
	sample.doneCond.L.Lock()
	defer sample.doneCond.L.Unlock()
	sample.done = true
	sample.doneCond.Broadcast()
}
