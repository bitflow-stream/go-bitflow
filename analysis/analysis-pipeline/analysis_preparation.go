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
		&TagDistributor{params},
		SimpleMultiFileBuilder(nil)))
}

const remoteInjectionTag = "remote-target"
const remoteInjectionSeparator = "|"

type InjectionInfo struct {
	injected string
	measured string
	anomaly  string
}

type InjectionInfoDistributor struct {
}

func (*InjectionInfoDistributor) String() string {
	return "injection info (tags " + ClassTag + " and " + remoteInjectionTag + ")"
}

func (*InjectionInfoDistributor) Distribute(sample *sample.Sample, header *sample.Header) []interface{} {
	var info InjectionInfo
	if sample.HasTag("host") {
		info.measured = sample.Tag("host")
	} else {
		info.measured = sample.Tag(SourceTag)
	}

	if sample.HasTag(ClassTag) {
		// A local injection
		info.injected = info.measured
		info.anomaly = sample.Tag(ClassTag)
	} else if sample.HasTag(remoteInjectionTag) {
		// A remote injection
		remote := sample.Tag(remoteInjectionTag) // Ignore remote-target0 etc.
		parts := strings.Split(remote, remoteInjectionSeparator)
		if len(parts) != 2 {
			log.Warnln("Tag", remoteInjectionTag, "has invalid value:", remote)
		}
		info.injected = parts[0]
		info.anomaly = parts[1]
	}
	return []interface{}{info}
}

func split_distributed_experiments(p *SamplePipeline) {
	distributor := new(InjectionInfoDistributor)
	builder := &MultiFilePipelineBuilder{
		Description: "directory structure <injected host>/<anomaly>/<measured host>.xxx",
		NewFile: func(oldFile string, key interface{}) string {
			info, ok := key.(InjectionInfo)
			if !ok {
				log.Fatalf("split_distributed_experiments: Unexpected subpipeline key %v (%T). Expecting InjectionInfo.", key, key)
			}
			folder := filepath.Dir(oldFile)
			extension := filepath.Ext(oldFile)

			injected := info.injected
			anomaly := info.anomaly
			measured := info.measured
			if injected == "" {
				injected = "_unknown_"
			}
			if anomaly == "" {
				anomaly = "_unknown_"
			}
			if measured == "" {
				measured = "_unknown_"
			}
			return filepath.Join(folder, injected, anomaly, measured+extension)
		},
	}
	builder.Build = func() []sample.SampleProcessor {
		return []sample.SampleProcessor{
			NewMultiHeaderMerger(),
			new(BatchProcessor).Add(new(SampleSorter)),
		}
	}

	p.Add(NewMetricFork(distributor, builder))
}
