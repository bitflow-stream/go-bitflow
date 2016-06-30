package main

import (
	"flag"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
)

const host_tag = "host"

var (
	host_sorter = &SampleSorter{[]string{host_tag}}
	host_tagger = &SampleTagger{[]string{host_tag}}
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

	RegisterAnalysis("merge_hosts", host_tagger, merge_hosts)
	RegisterAnalysis("pick_10percent", nil, pick_10percent)
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
			log.Println("Warning: --featureStats not given, not storing feature statistics")
		} else {
			p.Add(NewStoreStats(*targetFile))
		}
		p.Add(new(BatchProcessor).Add(new(StandardizationScaling)).Add(new(SampleShuffler)))
	}
}

func filter_basic(p *sample.CmdSamplePipeline) {
	p.Add(NewMetricFilter().IncludeRegex("^cpu$|^mem/percent$|^net-io/bytes$|^disk-io/[s|v]da/ioTime$"))
}

func merge_hosts(p *sample.CmdSamplePipeline) {
	p.Add(NewMultiHeaderMerger())

	suffix_len := len("-1.xxx")
	if filesource, ok := p.Source.(*sample.FileSource); ok {
		filesource.ConvertFilename = func(filename string) string {
			name := filepath.Base(filename)
			if len(name) >= suffix_len {
				return name[:len(name)-suffix_len]
			} else {
				return name
			}
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
