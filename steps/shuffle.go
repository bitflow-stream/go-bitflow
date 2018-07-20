package steps

import (
	"log"
	"math/rand"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
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

func RegisterSampleShuffler(b *query.PipelineBuilder) {
	b.RegisterAnalysis("shuffle",
		func(p *pipeline.SamplePipeline) {
			p.Batch(NewSampleShuffler())
		},
		"Shuffle a batch of samples to a random ordering")
}
