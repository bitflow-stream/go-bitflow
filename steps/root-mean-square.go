package steps

import (
	"math"

	bitflow "github.com/antongulenko/go-bitflow"
)

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
