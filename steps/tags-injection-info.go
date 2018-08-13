package steps

import (
	"fmt"
	"strings"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	log "github.com/sirupsen/logrus"
)

// TODO move this to a plugin

func RegisterTargetTagSplitter(b *query.PipelineBuilder) {
	const (
		targetTag          = "target"
		targetTagSeparator = "|"
		injectedTag        = "injected"
		anomalyTag         = "anomaly"
	)

	description := fmt.Sprintf("Split '%v' tag into '%v' and '%v'", targetTag, injectedTag, anomalyTag)
	b.RegisterAnalysis("split_target_tag", func(pipe *pipeline.SamplePipeline) {
		proc := &pipeline.SimpleProcessor{
			Description: description,
			Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
				if sample.HasTag(targetTag) {
					remote := sample.Tag(targetTag)
					parts := strings.Split(remote, targetTagSeparator)
					if len(parts) != 2 {
						log.Warnln("Tag", targetTag, "has invalid value:", remote)
						return sample, header, nil
					}
					sample.SetTag(injectedTag, parts[0])
					sample.SetTag(anomalyTag, parts[1])
				}
				return sample, header, nil
			},
		}
		pipe.Add(proc)
	}, description)
}
