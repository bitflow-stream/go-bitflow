package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
)

func init() {
	RegisterAnalysis("tag_injection_info", tag_injection_info, "Convert tags (cls, target) into (injected, measured, anomaly)")
	RegisterAnalysis("injection_directory_structure", injection_directory_structure, "Split samples into a directory structure based on the tags provided by tag_injection_info. Must be used as last step before a file output")
	RegisterAnalysisParamsErr("split_experiments", split_experiments, "Split samples into separate files based on their timestamps. When the difference between two (sorted) timestamps is too large, start a new file. Must be used as the last step before a file output", []string{"min_duration"})
}

const (
	remoteInjectionTag       = "target"
	remoteInjectionSeparator = "|"
	injectedTag              = "injected"
	anomalyTag               = "anomaly"
	measuredTag              = "measured"
)

func tag_injection_info(p *Pipeline) {
	p.Add(&SimpleProcessor{
		Description: fmt.Sprintf("Injection info tagger (transform tags %s and %s into %s, %s and %s)", ClassTag, remoteInjectionTag, injectedTag, measuredTag, anomalyTag),
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			measured := ""
			injected := ""
			anomaly := ""

			if sample.HasTag("host") {
				measured = sample.Tag("host")
			} else {
				measured = sample.Tag(SourceTag)
			}
			if sample.HasTag(ClassTag) {
				// A local injection
				injected = measured
				anomaly = sample.Tag(ClassTag)
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
	})
}

func injection_directory_structure(p *Pipeline) {
	distributor := &TagsDistributor{
		Tags:        []string{injectedTag, anomalyTag, measuredTag},
		Separator:   string(filepath.Separator),
		Replacement: "_unknown_",
	}
	builder := MultiFileDirectoryBuilder(false, nil)
	p.Add(&MetricFork{
		MultiPipeline: MultiPipeline{
			ParallelClose: true,
		},
		Distributor: distributor,
		Builder:     builder},
	)
}

type TimeDistributor struct {
	MinimumPause time.Duration

	counter  int
	lastTime time.Time
}

func (d *TimeDistributor) String() string {
	return "split after pauses of " + d.MinimumPause.String()
}

func (d *TimeDistributor) Distribute(sample *bitflow.Sample, header *bitflow.Header) []interface{} {
	last := d.lastTime
	d.lastTime = sample.Time
	if !last.IsZero() && sample.Time.Sub(last) >= d.MinimumPause {
		d.counter++
	}
	return []interface{}{d.counter}
}

func split_experiments(p *Pipeline, params map[string]string) error {
	duration, err := time.ParseDuration(params["min_duration"])
	if err != nil {
		err = parameterError("min_duration", err)
	} else {
		p.Add(&MetricFork{
			MultiPipeline: MultiPipeline{
				ParallelClose: true,
			},
			Distributor: &TimeDistributor{MinimumPause: duration},
			Builder:     MultiFileSuffixBuilder(nil),
		})
	}
	return err
}
