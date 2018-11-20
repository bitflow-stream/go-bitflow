package steps

import (
	"fmt"

	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
)

func RegisterTaggingProcessor(b reg.ProcessorRegistry) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) {
		p.Add(NewTaggingProcessor(params))
	}
	b.RegisterAnalysisParams("tags", create, "Set the given tags on every sample")
}

func NewTaggingProcessor(tags map[string]string) bitflow.SampleProcessor {
	templates := make(map[string]pipeline.TagTemplate, len(tags))
	for key, value := range tags {
		templates[key] = pipeline.TagTemplate{
			Template:     value,
			MissingValue: "",
		}
	}

	return &pipeline.SimpleProcessor{
		Description: fmt.Sprintf("Set tags %v", tags),
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			for key, template := range templates {
				value := template.Resolve(sample)
				sample.SetTag(key, value)
			}
			return sample, header, nil
		},
	}
}
