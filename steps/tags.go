package steps

import (
	"fmt"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterTaggingProcessor(b reg.ProcessorRegistry) {
	b.RegisterStep("tags",
		func(p *bitflow.SamplePipeline, params map[string]string) error {
			p.Add(NewTaggingProcessor(params))
			return nil
		},
		"Set the given tags on every sample",
		reg.VariableParams())
}

func NewTaggingProcessor(tags map[string]string) bitflow.SampleProcessor {
	templates := make(map[string]bitflow.TagTemplate, len(tags))
	for key, value := range tags {
		templates[key] = bitflow.TagTemplate{
			Template:     value,
			MissingValue: "",
		}
	}

	return &bitflow.SimpleProcessor{
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
