package main

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
)

func init() {
	RegisterSampleHandler("host", &SampleTagger{SourceTags: []string{"host"}, DontOverwrite: true})

	RegisterAnalysis("aggregate_10s", aggregate_data_10s)
	RegisterAnalysis("filter_basic", filter_basic)
	RegisterAnalysis("filter_hypervisor", filter_hypervisor)
	RegisterAnalysis("merge_hosts", merge_hosts)
	RegisterAnalysis("convert_filenames", convert_filenames)
	RegisterAnalysis("convert_filenames2", convert_filenames2)
	RegisterAnalysisParams("tag_split_files", split_tag_in_files, "tag to use as splitter")

	RegisterAnalysis("tag_injection_info", tag_injection_info)
	RegisterAnalysis("split_distributed_experiments", split_distributed_experiments)
	RegisterAnalysis("split_distributed_experiments_only", split_distributed_experiments_only)
	RegisterAnalysisParams("split_experiments", split_experiments, "number of seconds without sample before starting a new file")
}

func aggregate_data_10s(p *SamplePipeline) {
	// TODO properly parameterize the aggregator, move to analysis_basic.go
	p.Add((&FeatureAggregator{WindowDuration: 10 * time.Second}).AddAvg("_avg").AddSlope("_slope"))
}

func filter_basic(p *SamplePipeline) {
	p.Add(NewMetricFilter().IncludeRegex("^cpu$|^mem/percent$|^net-io/bytes$|^disk-io/[s|v]da/ioTime$"))
}

func filter_hypervisor(p *SamplePipeline) {
	p.Add(NewMetricFilter().ExcludeRegex("^ovsdb/|^libvirt/"))
}

func merge_hosts(p *SamplePipeline) {
	merge_headers(p)

	suffix_regex := regexp.MustCompile("\\....$")  // Strip file ending
	num_regex := regexp.MustCompile("(-[0-9]+)?$") // Strip optional appended numbering
	if filesource, ok := p.Source.(*bitflow.FileSource); ok {
		filesource.ConvertFilename = func(filename string) string {
			name := filepath.Base(filename)
			name = suffix_regex.ReplaceAllString(name, "")
			name = num_regex.ReplaceAllString(name, "")
			return name
		}
	}
}

func convert_filenames(p *SamplePipeline) {
	// Replace the src tag with the name of the parent folder
	if filesource, ok := p.Source.(*bitflow.FileSource); ok {
		filesource.ConvertFilename = func(filename string) string {
			return filepath.Base(filepath.Dir(filename))
		}
	}
}

func convert_filenames2(p *SamplePipeline) {
	// Replace the src tag with the name of the parent-parent folder
	if filesource, ok := p.Source.(*bitflow.FileSource); ok {
		filesource.ConvertFilename = func(filename string) string {
			return filepath.Base(filepath.Dir(filepath.Dir(filename)))
		}
	}
}

func split_tag_in_files(p *SamplePipeline, params string) {
	p.Add(NewMetricFork(
		&TagsDistributor{
			Tags:        []string{params},
			Separator:   "_",
			Replacement: "empty",
		},
		MultiFileSuffixBuilder(nil)))
}

const (
	remoteInjectionTag       = "remote-target"
	remoteInjectionSeparator = "|"
	injectedTag              = "injected"
	anomalyTag               = "anomaly"
	measuredTag              = "measured"
)

type InjectionInfoTagger struct {
	bitflow.AbstractProcessor
}

func (*InjectionInfoTagger) String() string {
	return "injection info tagger (tags " + ClassTag + " and " + remoteInjectionTag + ")"
}

func (p *InjectionInfoTagger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
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
		remote := sample.Tag(remoteInjectionTag) // Ignore remote-target0 etc.
		parts := strings.Split(remote, remoteInjectionSeparator)
		if len(parts) != 2 {
			log.Warnln("Tag", remoteInjectionTag, "has invalid value:", remote)
		}
		injected = parts[0]
		anomaly = parts[1]
	}

	sample.SetTag(injectedTag, injected)
	sample.SetTag(measuredTag, measured)
	sample.SetTag(anomalyTag, anomaly)
	return p.OutgoingSink.Sample(sample, header)
}

func tag_injection_info(p *SamplePipeline) {
	p.Add(new(InjectionInfoTagger))
}

func split_distributed_experiments(p *SamplePipeline) {
	distributor := &TagsDistributor{
		Tags:        []string{injectedTag, anomalyTag, measuredTag},
		Separator:   string(filepath.Separator),
		Replacement: "_unknown_",
	}
	builder := MultiFileDirectoryBuilder(false, func() []bitflow.SampleProcessor {
		return []bitflow.SampleProcessor{
			NewMultiHeaderMerger(),
			new(BatchProcessor).Add(new(SampleSorter)),
		}
	})
	p.Add(NewMetricFork(distributor, builder))
}

func split_distributed_experiments_only(p *SamplePipeline) {
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
