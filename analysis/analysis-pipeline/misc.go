package main

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"
	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
)

var (
	filter_label string
)

func init() {
	RegisterAnalysis("print_labels", nil, print_labels)
	RegisterAnalysis("filter_label", nil, filter_labels)
	flag.StringVar(&filter_label, "filter_label", "", "Used for -e filter_label")
}

type UniqueTagPrinter struct {
	Tag string
}

func (printer *UniqueTagPrinter) ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample, error) {
	log.Println("Printing", printer.Tag, "tags of", len(samples), "samples")
	labels := make(map[string]bool)
	for _, sample := range samples {
		labels[sample.Tag(printer.Tag)] = true
	}
	for label := range labels {
		fmt.Println(label)
	}
	return header, samples, nil
}

func (printer *UniqueTagPrinter) String() string {
	return "Print " + printer.Tag + " tags"
}

func print_labels(p *sample.CmdSamplePipeline) {
	p.Add(new(BatchProcessor).Add(&UniqueTagPrinter{"cls"}))
}

func filter_labels(p *sample.CmdSamplePipeline) {
	p.Add(&SampleFilter{
		IncludeFilter: func(inSample *sample.Sample) bool {
			return inSample.Tag("cls") == filter_label
		},
	})
}
