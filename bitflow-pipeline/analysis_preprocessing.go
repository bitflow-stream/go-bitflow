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
	RegisterAnalysis("tag_injection_info", tag_injection_info)
	RegisterAnalysis("injection_directory_structure", injection_directory_structure)
	RegisterAnalysisParams("split_experiments", split_experiments, "number of seconds without sample before starting a new file")
}

const (
	remoteInjectionTag       = "target"
	remoteInjectionSeparator = "|"
	injectedTag              = "injected"
	anomalyTag               = "anomaly"
	measuredTag              = "measured"
)

func tag_injection_info(p *SamplePipeline) {
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

func injection_directory_structure(p *SamplePipeline) {
	distributor := &TagsDistributor{
		Tags:        []string{injectedTag, anomalyTag, measuredTag},
		Separator:   string(filepath.Separator),
		Replacement: "_unknown_",
	}
	builder := MultiFileDirectoryBuilder(false, nil)
	p.Add(NewMetricFork(distributor, builder))
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

func split_experiments(p *SamplePipeline, params string) {
	duration, err := time.ParseDuration(params)
	if err != nil {
		log.Fatalln("Error parsing duration parameter for -e split_experiments:", err)
	}
	p.Add(NewMetricFork(&TimeDistributor{MinimumPause: duration}, MultiFileSuffixBuilder(nil)))
}
