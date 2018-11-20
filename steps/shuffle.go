package steps

import (
	"math/rand"

	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
	log "github.com/sirupsen/logrus"
)

func NewSampleShuffler() *pipeline.SimpleBatchProcessingStep {
	return &pipeline.SimpleBatchProcessingStep{
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
		func(p *pipeline.SamplePipeline) {
			p.Batch(NewSampleShuffler())
		},
		"Shuffle a batch of samples to a random ordering",
		reg.SupportBatch())
}
