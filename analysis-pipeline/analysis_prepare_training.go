package main

import (
	"flag"
	"math/rand"
	"path/filepath"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

const host_tag = "host"

var (
	host_sorter = &SampleSorter{[]string{host_tag}}
	host_tagger = &SampleTagger{SourceTags: []string{host_tag}, DontOverwrite: true}

	metric_filter_include golib.StringSlice
	metric_filter_exclude golib.StringSlice
)

func init() {
	RegisterAnalysis("prepare", setSampleSource, prepare_training_data)
	RegisterAnalysis("prepare_sorted", setSampleSource, prepare_training_data_sorted)
	RegisterAnalysis("prepare_shuffled", setSampleSource, prepare_training_data_shuffled)

	targetFile := flag.String("featureStats", "", "Target file for printing feature statistics (for -e aggregate_scale)")
	RegisterAnalysis("aggregate_scale", nil, aggregate_and_scale(targetFile))
	RegisterAnalysis("shuffle", nil, shuffle_data)
	RegisterAnalysis("sort", nil, sort_data)
	RegisterAnalysis("filter_basic", setSampleSource, filter_basic)
	RegisterAnalysis("filter_hypervisor", setSampleSource, filter_hypervisor)

	RegisterAnalysis("merge_hosts", host_tagger, merge_hosts)
	RegisterAnalysis("pick_10percent", nil, pick_10percent)

	RegisterAnalysis("filter_metrics", nil, filter_metrics)
	flag.Var(&metric_filter_include, "metrics_include", "Include regex used with '-e filter_metrics'")
	flag.Var(&metric_filter_exclude, "metrics_exclude", "Exclude regex used with '-e filter_metrics'")
}

func prepare_training_data_shuffled(p *sample.CmdSamplePipeline) {
	prepare_training_data(p)
	p.Add(new(BatchProcessor).Add(new(SampleShuffler)))
}

func prepare_training_data_sorted(p *sample.CmdSamplePipeline) {
	prepare_training_data(p)
	p.Add(new(BatchProcessor).Add(host_sorter))
}

func prepare_training_data(p *sample.CmdSamplePipeline) {
	convertFilenames(&p.SamplePipeline)
	p.Add(NewMetricFilter().ExcludeRegex("libvirt|ovsdb"))
	p.Add(NewMultiHeaderMerger())
}

func shuffle_data(p *sample.CmdSamplePipeline) {
	p.Add(new(BatchProcessor).Add(new(SampleShuffler)))
}

func sort_data(p *sample.CmdSamplePipeline) {
	p.Add(new(BatchProcessor).Add(host_sorter))
}

func aggregate_and_scale(targetFile *string) func(p *sample.CmdSamplePipeline) {
	return func(p *sample.CmdSamplePipeline) {
		p.Add(new(BatchProcessor).Add(host_sorter))
		p.Add((&FeatureAggregator{WindowDuration: 10 * time.Second}).AddAvg("_avg").AddSlope("_slope"))
		if targetFile == nil || *targetFile == "" {
			log.Warnln("--featureStats not given, not storing feature statistics")
		} else {
			p.Add(NewStoreStats(*targetFile))
		}
		p.Add(new(BatchProcessor).Add(new(StandardizationScaling)).Add(new(SampleShuffler)))
	}
}

func filter_basic(p *sample.CmdSamplePipeline) {
	p.Add(NewMetricFilter().IncludeRegex("^cpu$|^mem/percent$|^net-io/bytes$|^disk-io/[s|v]da/ioTime$"))
}

func filter_hypervisor(p *sample.CmdSamplePipeline) {
	p.Add(NewMetricFilter().ExcludeRegex("^ovsdb/|^libvirt/"))
}

func merge_hosts(p *sample.CmdSamplePipeline) {
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

func pick_10percent(p *sample.CmdSamplePipeline) {
	p.Add(&SampleFilter{
		IncludeFilter: func(inSample *sample.Sample) bool {
			return rand.Int63()%10 == 0
		},
	})
}

func filter_metrics(p *sample.CmdSamplePipeline) {
	filter := NewMetricFilter()
	for _, include := range metric_filter_include {
		filter.IncludeRegex(include)
	}
	for _, exclude := range metric_filter_exclude {
		filter.ExcludeRegex(exclude)
	}
	p.Add(filter)
}
