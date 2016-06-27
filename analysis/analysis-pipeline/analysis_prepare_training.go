package main

import (
	"flag"
	"log"
	"time"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
)

func init() {
	RegisterAnalysis("prepare", setSampleSource, prepare_training_data)
	RegisterAnalysis("prepare_sorted", setSampleSource, prepare_training_data_sorted)
	RegisterAnalysis("prepare_shuffled", setSampleSource, prepare_training_data_shuffled)

	targetFile := flag.String("featureStats", "", "Target file for printing feature statistics (for -e aggregate_scale)")
	RegisterAnalysis("aggregate_scale", nil, aggregate_and_scale(targetFile))
	RegisterAnalysis("shuffle", nil, shuffle_data)
}

func prepare_training_data_shuffled(p *sample.CmdSamplePipeline) {
	prepare_training_data(p)
	p.Add(new(BatchProcessor).Add(new(SampleShuffler)))
}

func prepare_training_data_sorted(p *sample.CmdSamplePipeline) {
	prepare_training_data(p)
	p.Add(new(BatchProcessor).Add(new(TimestampSort)))
}

func prepare_training_data(p *sample.CmdSamplePipeline) {
	convertFilenames(&p.SamplePipeline)
	p.Add(NewMetricFilter().ExcludeRegex("libvirt|ovsdb"))
	p.Add(NewMultiHeaderMerger())
}

func shuffle_data(p *sample.CmdSamplePipeline) {
	p.Add(new(BatchProcessor).Add(new(SampleShuffler)))
}

func aggregate_and_scale(targetFile *string) func(p *sample.CmdSamplePipeline) {
	return func(p *sample.CmdSamplePipeline) {
		p.Add(new(BatchProcessor).Add(new(TimestampSort)))
		p.Add((&FeatureAggregator{WindowDuration: 10 * time.Second}).AddAvg("_avg").AddSlope("_slope"))
		if targetFile == nil || *targetFile == "" {
			log.Println("Warning: --featureStats not given, not storing feature statistics")
		} else {
			p.Add(NewStoreStats(*targetFile))
		}
		p.Add(new(BatchProcessor).Add(new(StandardizationScaling)).Add(new(SampleShuffler)))
	}
}
