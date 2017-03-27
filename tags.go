package pipeline

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
)

func NewTaggingProcessor(tags map[string]string) bitflow.SampleProcessor {
	return &SimpleProcessor{
		Description: fmt.Sprintf("Set tags %v", tags),
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			for key, value := range tags {
				sample.SetTag(key, value)
			}
			return sample, header, nil
		},
	}
}
