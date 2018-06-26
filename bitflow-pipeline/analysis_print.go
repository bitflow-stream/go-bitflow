package main

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	log "github.com/sirupsen/logrus"
)

func RegisterPrintAnalyses(b *query.PipelineBuilder) {
	b.RegisterAnalysis("print_header", print_header, "Print every changing header to the log")
	b.RegisterAnalysisParams("print_tags", print_tags, "When done processing, print every encountered value of the given tag", []string{"tag"})
	b.RegisterAnalysisParams("count_tags", count_tags, "When done processing, print the number of times every value of the given tag was encountered", []string{"tag"})
	b.RegisterAnalysis("print_timerange", print_time_range, "When done processing, print the first and last encountered timestamp")
	b.RegisterAnalysisParamsErr("histogram", print_timeline, "When done processing, print a timeline showing a rudimentary histogram of the number of samples", []string{}, "buckets")
	b.RegisterAnalysis("count_invalid", count_invalid, "When done processing, print the number of invalid metric values and samples containing such values (NaN, -/+ infinity, ...)")
	b.RegisterAnalysis("print_common_metrics", print_common_metrics, "When done processing, print the metrics that occurred in all processed headers")
}

func print_header(p *SamplePipeline) {
	var checker bitflow.HeaderChecker
	numSamples := 0
	p.Add(&SimpleProcessor{
		Description: "header printer",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if checker.HeaderChanged(header) {
				if checker.LastHeader != nil {
					log.Println("Samples after last header:", numSamples)
				}
				log.Println(header)
				numSamples = 0
			}
			numSamples++
			return sample, header, nil
		},
		OnClose: func() {
			log.Println("Samples after last header:", numSamples)
		},
	})
}

type UniqueTagPrinter struct {
	bitflow.NoopProcessor
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
	val := sample.Tag(printer.Tag)
	if printer.Count {
		printer.values[val] = printer.values[val] + 1
	} else {
		printer.values[val] = 1
	}
	return printer.NoopProcessor.Sample(sample, header)
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
	printer.NoopProcessor.Close()
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

func print_tags(p *SamplePipeline, params map[string]string) {
	p.Add(NewUniqueTagPrinter(params["tag"]))
}

func count_tags(p *SamplePipeline, params map[string]string) {
	p.Add(NewUniqueTagCounter(params["tag"]))
}

func print_time_range(p *SamplePipeline) {
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

func print_timeline(p *SamplePipeline, params map[string]string) error {
	numBuckets := uint64(10)
	if bucketsStr, hasBuckets := params["buckets"]; hasBuckets {
		var err error
		numBuckets, err = strconv.ParseUint(bucketsStr, 10, 64)
		if err != nil {
			return query.ParameterError("buckets", err)
		}
	}
	if numBuckets <= 0 {
		numBuckets = 1
	}

	p.Batch(&SimpleBatchProcessingStep{
		Description: fmt.Sprintf("Print timeline (len %v)", numBuckets),
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			var from, to time.Time
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
			bucketDuration := duration / time.Duration(numBuckets)
			buckets := make([]int, numBuckets)
			bucketEnds := make([]time.Time, numBuckets)

			for i := uint64(0); i < numBuckets-1; i++ {
				bucketEnds[i] = from.Add(time.Duration(i+1) * bucketDuration)
			}
			bucketEnds[numBuckets-1] = to // No rounding error
			for _, sample := range samples {
				index := sort.Search(len(buckets), func(n int) bool {
					return !sample.Time.After(bucketEnds[n])
				})
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
		},
	})
	return nil
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
