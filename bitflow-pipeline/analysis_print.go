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
	. "github.com/antongulenko/go-bitflow-pipeline"
)

func init() {
	RegisterAnalysis("print", print_samples)
	RegisterAnalysisParams("print_tags", print_tags, "tag to print")
	RegisterAnalysisParams("count_tags", count_tags, "tag to count")
	RegisterAnalysis("print_timerange", print_timerange)
	RegisterAnalysisParams("print_timeline", print_timeline, "number of buckets for the timeline-histogram") // Print a timeline showing a rudimentary histogram of the number of samples
	RegisterAnalysis("count_invalid", count_invalid)
	RegisterAnalysis("print_common_metrics", print_common_metrics)
}

func print_samples(p *SamplePipeline) {
	p.Add(NewSamplePrinter())
}

type UniqueTagPrinter struct {
	bitflow.AbstractProcessor
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
	keys := make([]string, 0, len(printer.values))
	for label := range printer.values {
		keys = append(keys, label)
	}
	sort.Strings(keys)

	log.Println("Now outputting results of", printer)
	for _, label := range keys {
		// Print to stdout instead of logger
		if printer.Count {
			count := printer.values[label]
			if label == "" {
				label = "(missing value)"
			}
			fmt.Println(label, count)
			total += count
		} else if label != "" {
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
	return res + " unique values of tag '" + printer.Tag + "'"
}

func print_tags(p *SamplePipeline, params string) {
	p.Add(NewUniqueTagPrinter(params))
}

func count_tags(p *SamplePipeline, params string) {
	p.Add(NewUniqueTagCounter(params))
}

func print_timerange(p *SamplePipeline) {
	var (
		from  time.Time
		to    time.Time
		count int
	)
	p.Add(&SimpleProcessor{
		Description: "Print time range",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			count++
			t := sample.Time
			if from.IsZero() || to.IsZero() {
				from = t
				to = t
			} else if t.Before(from) {
				from = t
			} else if t.After(to) {
				to = t
			}
			return sample, header, nil
		},
		OnClose: func() {
			format := "02.01.2006 15:04:05"
			duration := to.Sub(from) / time.Millisecond * time.Millisecond // Round
			log.Printf("Time range of %v samples: %v - %v (%v)", count,
				from.Format(format), to.Format(format), duration)
		},
	})
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

func count_invalid(p *SamplePipeline) {
	var (
		invalidSamples int
		totalSamples   int
		invalidValues  int
		totalValues    int
	)
	p.Add(&SimpleProcessor{
		Description: "Invalid values counter",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			sampleValid := true
			for _, val := range sample.Values {
				totalValues += 1
				if !IsValidNumber(float64(val)) {
					invalidValues += 1
					sampleValid = false
				}
			}
			totalSamples += 1
			if !sampleValid {
				invalidSamples += 1
			}
			return sample, header, nil
		},
		OnClose: func() {
			log.Printf("Invalid numbers: %v of %v, in %v of %v samples",
				invalidValues, totalValues, invalidSamples, totalSamples)
		},
	})
}

func print_common_metrics(p *SamplePipeline) {
	var (
		checker bitflow.HeaderChecker
		common  map[string]bool
		num     int
	)
	p.Add(&SimpleProcessor{
		Description: "Common metrics printer",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if checker.HeaderChanged(header) {
				num++
				if common == nil {
					common = make(map[string]bool)
					for _, field := range header.Fields {
						common[field] = true
					}
				} else {
					incoming := make(map[string]bool)
					for _, field := range header.Fields {
						incoming[field] = true
					}
					for field := range common {
						if !incoming[field] {
							delete(common, field)
						}
					}
				}
			}
			return sample, header, nil
		},
		OnClose: func() {
			fields := make([]string, 0, len(common))
			for field := range common {
				fields = append(fields, field)
			}
			sort.Strings(fields)
			log.Printf("%v common metrics in %v headers: %v", len(fields), num, fields)
		},
	})
}
