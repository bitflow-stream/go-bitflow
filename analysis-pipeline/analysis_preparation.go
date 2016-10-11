package main

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
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
	if filesource, ok := p.Source.(*sample.FileSource); ok {
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
	if filesource, ok := p.Source.(*sample.FileSource); ok {
		filesource.ConvertFilename = func(filename string) string {
			return filepath.Base(filepath.Dir(filename))
		}
	}
}

func convert_filenames2(p *SamplePipeline) {
	// Replace the src tag with the name of the parent-parent folder
	if filesource, ok := p.Source.(*sample.FileSource); ok {
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
	AbstractProcessor
}

func (*InjectionInfoTagger) String() string {
	return "injection info tagger (tags " + ClassTag + " and " + remoteInjectionTag + ")"
}

func (p *InjectionInfoTagger) Sample(sample *sample.Sample, header *sample.Header) error {
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
	builder := MultiFileDirectoryBuilder(func() []sample.SampleProcessor {
		return []sample.SampleProcessor{
			NewMultiHeaderMerger(),
			new(BatchProcessor).Add(new(SampleSorter)),
		}
	})
	p.Add(NewMetricFork(distributor, builder))
}
