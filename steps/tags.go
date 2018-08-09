package steps

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

func RegisterTaggingProcessor(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) {
		p.Add(NewTaggingProcessor(params))
	}
	b.RegisterAnalysisParams("tags", create, "Set the given tags on every sample", nil)
}

func NewTaggingProcessor(tags map[string]string) bitflow.SampleProcessor {
	return &pipeline.SimpleProcessor{
		Description: fmt.Sprintf("Set tags %v", tags),
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			for key, value := range tags {
				sample.SetTag(key, value)
			}
			return sample, header, nil
		},
	}
}
