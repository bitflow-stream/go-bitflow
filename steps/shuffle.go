package steps

import (
	"math/rand"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

func NewSampleShuffler() *bitflow.SimpleBatchProcessingStep {
	return &bitflow.SimpleBatchProcessingStep{
		Description: "sample shuffler",
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			log.Println("Shuffling", len(samples), "samples")
			for i := range samples {
				j := rand.Intn(i + 1)
				samples[i], samples[j] = samples[j], samples[i]
			}
			return header, samples, nil
		},
	}
}

func RegisterSampleShuffler(b reg.ProcessorRegistry) {
	b.RegisterAnalysis("shuffle",
		func(p *bitflow.SamplePipeline) {
			p.Batch(NewSampleShuffler())
		},
		"Shuffle a batch of samples to a random ordering",
		reg.SupportBatch())
}
