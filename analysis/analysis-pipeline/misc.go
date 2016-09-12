package main

import (
	"fmt"
	"log"
	"time"

	"github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
)

func init() {
	RegisterAnalysisParams("print_tags", print_tags, "tag to print")
	RegisterAnalysis("print_timerange", print_timerange)
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
		// Print to stdout instead of logger
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

const TimrangePrinterFormat = "02.01.2006 15:04:05"

type TimerangePrinter struct {
	analysis.AbstractProcessor
	from  time.Time
	to    time.Time
	count int
}

func (printer *TimerangePrinter) Sample(sample sample.Sample, header sample.Header) error {
	if err := printer.Check(sample, header); err != nil {
		return err
	}
	printer.count++
	t := sample.Time
	if printer.from.IsZero() || printer.to.IsZero() {
		printer.from = t
		printer.to = t
	} else if t.Before(printer.from) {
		printer.from = t
	} else if t.After(printer.to) {
		printer.to = t
	}
	return printer.OutgoingSink.Sample(sample, header)
}

func (printer *TimerangePrinter) Close() {
	log.Printf("Time range of %v samples: %v - %v", printer.count, printer.from.Format(TimrangePrinterFormat), printer.to.Format(TimrangePrinterFormat))
	printer.AbstractProcessor.Close()
}

func (printer *TimerangePrinter) String() string {
	return "Print time range"
}

func print_timerange(p *SamplePipeline) {
	p.Add(new(TimerangePrinter))
}
