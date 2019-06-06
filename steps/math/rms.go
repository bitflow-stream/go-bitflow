package math

import (
	"math"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterRMS(b reg.ProcessorRegistry) {
	b.RegisterBatchStep("rms",
		func(_ map[string]string) (bitflow.BatchProcessingStep, error) {
			return new(BatchRms), nil
		},
		"Compute the Root Mean Square value for every metric in a data batch. Output a single sample with all values.")
}

type BatchRms struct {
}

func (r *BatchRms) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	if len(samples) == 0 {
		return header, samples, nil
	}
	res := make([]bitflow.Value, len(header.Fields))
	num := float64(len(samples))
	for i := range header.Fields {
		rms := float64(0)
		for _, sample := range samples {
			val := float64(sample.Values[i])
			rms += val * val / num
		}
		rms = math.Sqrt(rms)
		res[i] = bitflow.Value(rms)
	}
	outSample := samples[0].Clone() // Use the first sample as the reference for metadata (timestamp and tags)
	outSample.Values = res
	return header, []*bitflow.Sample{outSample}, nil
}

func (r *BatchRms) String() string {
	return "Root Mean Square"
}
