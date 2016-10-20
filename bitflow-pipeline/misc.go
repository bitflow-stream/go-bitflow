package main

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
)

func init() {
	RegisterAnalysisParams("print_tags", print_tags, "tag to print")
	RegisterAnalysisParams("count_tags", count_tags, "tag to count")
	RegisterAnalysis("print_timerange", print_timerange)
	RegisterAnalysisParams("print_timeline", print_timeline, "number of buckets for the timeline-histogram") // Print a timeline showing a rudimentary histogram of the number of samples
	RegisterAnalysis("count_invalid", count_invalid)
}

type UniqueTagPrinter struct {
	pipeline.AbstractProcessor
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

func (printer *UniqueTagPrinter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
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
			if label == "" {
				label = "(missing value)"
			}
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
	pipeline.AbstractProcessor
	from  time.Time
	to    time.Time
	count int
}

func (printer *TimerangePrinter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
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
	duration := printer.to.Sub(printer.from) / time.Millisecond * time.Millisecond // Round
	log.Printf("Time range of %v samples: %v - %v (%v)", printer.count,
		printer.from.Format(TimrangePrinterFormat), printer.to.Format(TimrangePrinterFormat), duration)
	printer.AbstractProcessor.Close()
}

func (printer *TimerangePrinter) String() string {
	return "Print time range"
}

func print_timerange(p *SamplePipeline) {
	p.Add(new(TimerangePrinter))
}

type TimelinePrinter struct {
	NumBuckets uint64
}

func (p *TimelinePrinter) String() string {
	return fmt.Sprintf("Print timeline (len %v)", p.NumBuckets)
}

func (p *TimelinePrinter) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	var from time.Time
	var to time.Time
	for _, sample := range samples {
		t := sample.Time
		if from.IsZero() || to.IsZero() {
			from = t
			to = t
		} else if t.Before(from) {
			from = t
		} else if t.After(to) {
			to = t
		}
	}
	duration := to.Sub(from)
	bucketDuration := duration / time.Duration(p.NumBuckets)
	buckets := make([]int, p.NumBuckets)
	bucketEnds := make([]time.Time, p.NumBuckets)
	for i := uint64(0); i < p.NumBuckets-1; i++ {
		bucketEnds[i] = from.Add(time.Duration(i+1) * bucketDuration)
	}
	bucketEnds[p.NumBuckets-1] = to // No rounding error
	for _, sample := range samples {
		index := sort.Search(len(buckets), func(n int) bool {
			return !sample.Time.After(bucketEnds[n])
		})
		if index == len(buckets) {
			log.Fatalln("WRONG:", sample.Time, "FIRST:", bucketEnds[0], "LAST:", bucketEnds[len(bucketEnds)-1], "START:", from, "END:", to)
		}
		buckets[index]++
	}
	largestBuffer := 0
	for _, num := range buckets {
		if num > largestBuffer {
			largestBuffer = num
		}
	}
	var timeline bytes.Buffer
	for _, bucketSize := range buckets {
		if bucketSize == 0 {
			timeline.WriteRune('-')
		} else {
			num := int(math.Ceil(float64(bucketSize)/float64(largestBuffer)*10)) - 1 // [0..9]
			timeline.WriteString(strconv.Itoa(num))
		}
	}

	log.Println("[Timeline]: Start:", from)
	log.Println("[Timeline]: End:", to)
	log.Println("[Timeline]: Duration:", duration)
	log.Println("[Timeline]: One bucket:", bucketDuration)
	log.Println("[Timeline]:", timeline.String())
	return header, samples, nil
}

func print_timeline(p *SamplePipeline, param string) {
	buckets, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		log.Warnln("Failed to parse parameter for -e print_timeline:", err)
		buckets = 10
	}
	if buckets == 0 {
		buckets = 1
	}
	p.Batch(&TimelinePrinter{NumBuckets: buckets})
}

type InvalidCounter struct {
	pipeline.AbstractProcessor
	invalidSamples int
	totalSamples   int
	invalidValues  int
	totalValues    int
}

func (counter *InvalidCounter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	sampleValid := true
	for _, val := range sample.Values {
		counter.totalValues += 1
		if !pipeline.IsValidNumber(float64(val)) {
			counter.invalidValues += 1
			sampleValid = false
		}
	}
	counter.totalSamples += 1
	if !sampleValid {
		counter.invalidSamples += 1
	}
	return nil
}

func (counter *InvalidCounter) Close() {
	log.Printf("Invalid numbers: %v of %v, in %v of %v samples",
		counter.invalidValues, counter.totalValues, counter.invalidSamples, counter.totalSamples)
	counter.AbstractProcessor.Close()
}

func count_invalid(p *SamplePipeline) {
	p.Add(new(InvalidCounter))
}