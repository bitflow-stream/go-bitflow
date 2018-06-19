package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
	. "github.com/antongulenko/go-bitflow-pipeline/fork"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	log "github.com/sirupsen/logrus"
)

func RegisterPreprocessingSteps(b *query.PipelineBuilder) {
	b.RegisterAnalysis("tag_injection_info", tag_injection_info, "Convert tags (cls, target) into (injected, measured, anomaly)")
	b.RegisterAnalysisParamsErr("injection_directory_structure", injection_directory_structure, "Split samples into a directory structure based on the tags provided by tag_injection_info. Must be used as last step before a file output", nil)
	b.RegisterAnalysisParamsErr("split_experiments", split_experiments, "Split samples into separate files based on their timestamps. When the difference between two (sorted) timestamps is too large, start a new file. Must be used as the last step before a file output", []string{"min_duration", "file"}, "batch_tag")
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

func injection_directory_structure(p *SamplePipeline, params map[string]string) error {
	distributor := &TagTemplateDistributor{
		Template: fmt.Sprintf("%s/%s/%s", injectedTag, anomalyTag, measuredTag),
	}
	builder, err := make_multi_file_pipeline_builder(params)
	if err == nil {
		p.Add(&MetricFork{
			ParallelClose: true,
			Distributor:   distributor,
			Builder:       builder},
		)
	}
	return err
}

type PauseTagger struct {
	bitflow.NoopProcessor
	MinimumPause time.Duration
	Tag          string

	counter  int
	lastTime time.Time
}

func (d *PauseTagger) String() string {
	return fmt.Sprintf("increment tag '%v' after pauses of %v", d.Tag, d.MinimumPause.String())
}

func (d *PauseTagger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	last := d.lastTime
	d.lastTime = sample.Time
	if !last.IsZero() && sample.Time.Sub(last) >= d.MinimumPause {
		d.counter++
	}
	sample.SetTag(d.Tag, strconv.Itoa(d.counter))
	return d.GetSink().Sample(sample, header)
}

func split_experiments(p *SamplePipeline, params map[string]string) error {
	tag := params["batch_tag"]
	if tag == "" {
		tag = "experiment_batch"
	}
	duration, err := time.ParseDuration(params["min_duration"])
	if err != nil {
		return query.ParameterError("min_duration", err)
	}
	group := bitflow.NewFileGroup(params["file"])
	fileTemplate := group.BuildFilenameStr("${" + tag + "}")

	delete(params, "batch_tag")
	delete(params, "min_duration")
	delete(params, "file")
	builder, err := make_multi_file_pipeline_builder(params)
	if err == nil {
		p.Add(&PauseTagger{MinimumPause: duration, Tag: tag})
		p.Add(&MetricFork{
			ParallelClose: true,
			Distributor:   &TagTemplateDistributor{Template: fileTemplate},
			Builder:       builder,
		})
	}
	return err
}
