package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
)

func init() {
	RegisterAnalysisParams("print_tags", print_tags, "tag to print")
	RegisterAnalysisParams("count_tags", count_tags, "tag to count")
	RegisterAnalysis("print_timerange", print_timerange)
}

type UniqueTagPrinter struct {
	analysis.AbstractProcessor
	Tag    string
	Count  bool
	values map[string]int
}

func NewUniqueTagPrinter(tag string) *UniqueTagPrinter {
	return &UniqueTagPrinter{
		Tag:    tag,
		Count:  false,
		values: make(map[string]int),
	}
}

func NewUniqueTagCounter(tag string) *UniqueTagPrinter {
	return &UniqueTagPrinter{
		Tag:    tag,
		Count:  true,
		values: make(map[string]int),
	}
}

func (printer *UniqueTagPrinter) Sample(sample sample.Sample, header sample.Header) error {
	if err := printer.Check(sample, header); err != nil {
		return err
	}
	val := sample.Tag(printer.Tag)
	if printer.Count {
		printer.values[val] = printer.values[val] + 1
	} else {
		printer.values[val] = 1
	}
	return printer.OutgoingSink.Sample(sample, header)
}

func (printer *UniqueTagPrinter) Close() {
	total := 0
	for label, count := range printer.values {
		// Print to stdout instead of logger
		if printer.Count {
			fmt.Println(label, count)
			total += count
		} else {
			fmt.Println(label)
		}
	}
	if printer.Count {
		fmt.Println("Total", total)
	}
	printer.AbstractProcessor.Close()
}

func (printer *UniqueTagPrinter) String() string {
	var res string
	if printer.Count {
		res = "Count"
	} else {
		res = "Print"
	}
	return res + " unique " + printer.Tag + " tags"
}

func print_tags(p *SamplePipeline, params string) {
	p.Add(NewUniqueTagPrinter(params))
}

func count_tags(p *SamplePipeline, params string) {
	p.Add(NewUniqueTagCounter(params))
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
