package main

import (
	"path/filepath"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
)

func init() {
	RegisterSampleHandler("host", &SampleTagger{SourceTags: []string{"host"}, DontOverwrite: true})

	RegisterAnalysis("aggregate_scale", aggregate_and_scale) // Param: file for storing feature stats
	RegisterAnalysis("filter_basic", filter_basic)
	RegisterAnalysis("filter_hypervisor", filter_hypervisor)
	RegisterAnalysis("merge_hosts", merge_hosts)
	RegisterAnalysis("convert_filenames", convert_filenames)
}

func aggregate_and_scale(p *SamplePipeline, params string) {
	p.Batch(&SampleSorter{[]string{"host"}})
	p.Add((&FeatureAggregator{WindowDuration: 10 * time.Second}).AddAvg("_avg").AddSlope("_slope"))
	if params == "" {
		log.Warnln("No parameter given to -e aggregate_scale, not storing feature statistics")
	} else {
		p.Add(NewStoreStats(params))
	}
	p.Batch(new(StandardizationScaling)).Batch(new(SampleShuffler))
}

func filter_basic(p *SamplePipeline, _ string) {
	p.Add(NewMetricFilter().IncludeRegex("^cpu$|^mem/percent$|^net-io/bytes$|^disk-io/[s|v]da/ioTime$"))
}

func filter_hypervisor(p *SamplePipeline, _ string) {
	p.Add(NewMetricFilter().ExcludeRegex("^ovsdb/|^libvirt/"))
}

func merge_hosts(p *SamplePipeline, _ string) {
	p.Add(NewMultiHeaderMerger())

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

func convert_filenames(p *SamplePipeline, _ string) {
	// Replace the src tag with the name of the parent-parent folder
	if filesource, ok := p.Source.(*sample.FileSource); ok {
		filesource.ConvertFilename = func(filename string) string {
			return filepath.Base(filepath.Dir(filepath.Dir(filename)))
		}
	}
}
