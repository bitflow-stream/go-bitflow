package main

import (
	"fmt"

	"github.com/antongulenko/data2go/sample"
)

func init() {
	RegisterAnalysis("print_tags", print_tags) // param: tag to print
}

type UniqueTagPrinter struct {
	Tag string
}

func (printer *UniqueTagPrinter) ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample, error) {
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
	return "Print unique " + printer.Tag + " tags"
}

func print_tags(p *SamplePipeline, params string) {
	p.Batch(&UniqueTagPrinter{params})
}
