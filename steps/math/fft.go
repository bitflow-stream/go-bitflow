package math

import (
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/ktye/fft"
	log "github.com/sirupsen/logrus"
)

// TODO result modes: real part OR imaginary part OR both OR magnitude. Also, keep original values or not.
// TODO parametrize if the absolute value of the result should be used? What else is optional?
// TODO renaming of computed output metrics (prefix/suffix)
// TODO allow filtering the metrics that get computed
// TODO inverse FFT (for different input data modes)

const (
	FftFreqMetricName = "freq"
)

var (
	fftCache     = make(map[int]*fft.FFT)
	fftCacheLock sync.Mutex

	warnedResizedRadixes = make(map[int]bool)
	warningLock          sync.Mutex
)

func RegisterFFT(b reg.ProcessorRegistry) {
	b.RegisterBatchStep("fft",
		func(params map[string]interface{}) (bitflow.BatchProcessingStep, error) {
			return new(BatchFft), nil
		},
		"Compute a radix-2 FFT on every metric of the batch. Output the real and imaginary parts of the result")
}

type BatchFft struct {
	// By default, the upper half of the FFT results is cut off, because the FFT results are symmetric.
	// Set LeaveSymmetricPart to true to include preserve all result values.
	LeaveSymmetricPart bool

	// By default, the FFT results are normalized. Set SkipNormalization to true to avoid this.
	// Normalization helps comparing the results of FFTs with different numbers of input values.
	// Normalization is done by dividing every FFT result value by the total number of input values to the FFT.
	// If FillZeros=true, this is the closest power of two above the number of actual input samples. If FillZeros=false, this is the closest power of two beneath that.
	SkipNormalization bool

	// If the number of input samples is not exactly a power of two, the input batch must either be cut short or filled up with zeros.
	// The default is to cut the input to a power of two. set FillZeros to true, to instead fill the input with zeros to reach the next higher power of two.
	FillZeros bool

	// If FrequencyMetricIndex >= 0, the frequency for every FFT result value will be added as a new metric at the given index.
	// If the index is higher than the total number of input metrics, it will be appended as the last metric.
	// The name of that new metric will be FftFreqMetricName ("freq").
	FrequencyMetricIndex int

	// If SamplingFrequency > 0, it will be used as the total sampling frequency when computing the frequency metric.
	// If this is <= 0, the sampling frequency will be computed automatically from the timestamps of the first and last input sample, and the total number of samples.
	SamplingFrequency float64

	// FatalErrors can be set to true to return errors that occur when processing a batch.
	// Otherwise, the error will be printed as a warning and an empty result batch will be produced.
	FatalErrors bool
}

func getFft(num int) (*fft.FFT, error) {
	fftCacheLock.Lock()
	defer fftCacheLock.Unlock()
	res, ok := fftCache[num]
	if !ok {
		newFft, err := fft.New(num)
		if err != nil {
			return nil, err
		}
		res = &newFft
		fftCache[num] = res
	}
	return res, nil
}

func warnFftRadixResize(oldSize, newSize int) {
	warningLock.Lock()
	defer warningLock.Unlock()
	if !warnedResizedRadixes[oldSize] {
		log.Warnf("Radix-2 FFT computation changes number of processed samples from %v to %v", oldSize, newSize)
		warnedResizedRadixes[oldSize] = true
	}
}

func (s *BatchFft) samplingFrequency(samples []*bitflow.Sample) (float64, error) {
	samplingFreq := s.SamplingFrequency
	if samplingFreq <= 0 {
		if len(samples) < 2 {
			return 0, fmt.Errorf("The number of input samples is %v, not enough to compute the sampling frequency", len(samples))
		}
		start := samples[0].Time
		end := samples[len(samples)-1].Time
		timeDiff := end.Sub(start)
		if timeDiff < 0 {
			return 0, fmt.Errorf("First sample has later timestamp (%v) than last sample (%v)", start, end)
		}
		samplingFreq = float64(len(samples)) / timeDiff.Seconds()
	}
	return samplingFreq, nil
}

func (s *BatchFft) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	outHeader, outSamples, err := s.processBatch(header, samples)
	if err != nil && !s.FatalErrors {
		tagStr := ""
		if len(samples) > 0 {
			tagStr = ", tags of first received sample:" + samples[0].TagString()
		}
		log.Warnln("Error processing FFT:", err, tagStr)
		err = nil
		outHeader = header
		outSamples = nil
	}
	return outHeader, outSamples, err
}

func (s *BatchFft) processBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	// Get an FFT instance and change the number of samples to a power of 2, if necessary
	f, err := getFft(len(samples))
	if err != nil {
		return nil, nil, err
	}
	if f.N < len(samples) && s.FillZeros {
		// Get the next larger FFT size (power of two). The missing input values will be filled with zeros.
		f, err = getFft(f.N * 2)
		if err != nil {
			return nil, nil, err
		}
	}
	if f.N != len(samples) {
		warnFftRadixResize(len(samples), f.N)
		if f.N < len(samples) {
			samples = samples[:f.N]
		}
	}

	// Convert the data batch to a column-representation to compute the FFT separately for each metric
	numFields := len(header.Fields)
	outputFields := s.OutputSampleSize(numFields)
	cols := make([][]complex128, numFields)
	for i := range header.Fields {
		cols[i] = make([]complex128, f.N)
	}
	for j, sample := range samples {
		for i := range header.Fields {
			// TODO is the imaginary part always zero?
			cols[i][j] = complex(float64(sample.Values[i]), 0)
		}
		// Resize the sample to take the real and imaginary part of the result
		sample.Resize(outputFields)
	}

	// Create the metrics for the output header
	freqIndex := s.FrequencyMetricIndex
	if freqIndex > numFields {
		freqIndex = numFields
	}
	outHeader := header
	if outputFields != len(outHeader.Fields) {
		outHeader = &bitflow.Header{Fields: make([]string, outputFields)}
		copy(outHeader.Fields, header.Fields[:freqIndex])
		if freqIndex < len(header.Fields) {
			copy(outHeader.Fields[freqIndex+1:], header.Fields[freqIndex:])
		}
		outHeader.Fields[freqIndex] = FftFreqMetricName
	}

	// Optionally compute the sampling frequency from timestamps and number of samples
	samplingFrequency, err := s.samplingFrequency(samples)
	if err != nil {
		return nil, nil, err
	}

	// Optionally remove the second half of the results, because the FFT result is symmetric
	if !s.LeaveSymmetricPart {
		samples = samples[:(len(samples)/2)+1]
		samplingFrequency /= 2
	}

	// Compute the FFT for every metric, insert the results back into the input samples
	for i, col := range cols {
		res := f.Transform(col)
		for j, sample := range samples {
			val := math.Abs(real(res[j]))
			if !s.SkipNormalization {
				val /= float64(f.N)
			}
			valIndex := i
			if valIndex >= freqIndex {
				valIndex++
			}
			sample.Values[valIndex] = bitflow.Value(val)
		}
	}

	// Compute the frequencies
	fftFreqStep := samplingFrequency / float64(len(samples))
	for i, sample := range samples {
		sample.Values[freqIndex] = bitflow.Value(float64(i) * fftFreqStep)
	}

	return outHeader, samples, nil
}

func (s *BatchFft) OutputSampleSize(sampleSize int) int {
	res := sampleSize
	if s.FrequencyMetricIndex >= 0 {
		res++
	}
	return res
}

func (s *BatchFft) String() string {
	var sampleFreq, freqMetric string
	if s.SamplingFrequency > 0 {
		sampleFreq = "sample frequency: " + strconv.FormatFloat(s.SamplingFrequency, 'f', -1, 64)
	} else {
		sampleFreq = "auto sample frequency"
	}
	if s.FrequencyMetricIndex < 0 {
		freqMetric = "frequency metric not appended"
	} else {
		freqMetric = "frequency appended at index " + strconv.Itoa(s.FrequencyMetricIndex)
	}
	return fmt.Sprintf("FFT (cut symmetric part: %v, normalize: %v, fill zeros: %v, %v, %v)",
		!s.LeaveSymmetricPart, !s.SkipNormalization, s.FillZeros, sampleFreq, freqMetric)
}
