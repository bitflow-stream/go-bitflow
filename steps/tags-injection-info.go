package steps

import (
	"fmt"
	"strings"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	log "github.com/sirupsen/logrus"
)

const (
	classTag                 = "cls"
	remoteInjectionTag       = "target"
	remoteInjectionSeparator = "|"
	injectedTag              = "injected"
	anomalyTag               = "anomaly"
	measuredTag              = "measured"
)

// TODO move this to a plugin

func RegisterInjectionInfoTagger(b *query.PipelineBuilder) {
	create := func(pipe *pipeline.SamplePipeline) {
		proc := &pipeline.SimpleProcessor{
			Description: fmt.Sprintf("Injection info tagger (transform tags %s and %s into %s, %s and %s)", classTag, remoteInjectionTag, injectedTag, measuredTag, anomalyTag),
			Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
				measured := ""
				injected := ""
				anomaly := ""

				if sample.HasTag("host") {
					measured = sample.Tag("host")
				} else {
					measured = sample.Tag("src")
				}
				if sample.HasTag(classTag) {
					// A local injection
					injected = measured
					anomaly = sample.Tag(classTag)
				} else if sample.HasTag(remoteInjectionTag) {
					// A remote injection
					remote := sample.Tag(remoteInjectionTag) // Ignore target0, target1 etc.
					parts := strings.Split(remote, remoteInjectionSeparator)
					if len(parts) != 2 {
						log.Warnln("Tag", remoteInjectionTag, "has invalid value:", remote)
						return sample, header, nil
					}
					injected = parts[0]
					anomaly = parts[1]
				}

				sample.SetTag(injectedTag, injected)
				sample.SetTag(measuredTag, measured)
				sample.SetTag(anomalyTag, anomaly)
				return sample, header, nil
			},
		}
		pipe.Add(proc)
	}

	b.RegisterAnalysis("tag_injection_info", create, "Convert tags (cls, target) into (injected, measured, anomaly)")
}
