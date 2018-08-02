package steps

import (
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
)

func RegisterDrop(b builder.PipelineBuilder) {
	b.RegisterAnalysis("drop",
		func(p *pipeline.SamplePipeline) {
			p.Add(&pipeline.SimpleProcessor{
				Description: "Drop all samples",
				Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
					return nil, nil, nil
				},
			})
		},
		"Drop all samples")
}
