package steps

import (
	"container/list"
	"fmt"
	"strconv"
	"time"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
)

type FeatureAggregator struct {
	bitflow.NoopProcessor
	WindowSize     int           // Applied if >0
	WindowDuration time.Duration // Applied if >0
	UseCurrentTime bool          // If true, use time.Now() as reference for WindowTime. Otherwise, use the timestamp of the latest Sample.

	aggregators []FeatureAggregatorOperation
	suffixes    []string

	checker            bitflow.HeaderChecker
	outHeader          *bitflow.Header
	allStats           map[string]*FeatureWindowStats
	currentHeaderStats []*FeatureWindowStats
}

type FeatureAggregatorOperation func(stats *FeatureWindowStats) bitflow.Value

func (agg *FeatureAggregator) Add(suffix string, operation FeatureAggregatorOperation) *FeatureAggregator {
	agg.aggregators = append(agg.aggregators, operation)
	agg.suffixes = append(agg.suffixes, suffix)
	return agg
}

func (agg *FeatureAggregator) AddAvg(suffix string) *FeatureAggregator {
	return agg.Add(suffix, FeatureWindowAverage)
}

func (agg *FeatureAggregator) AddSlope(suffix string) *FeatureAggregator {
	return agg.Add(suffix, FeatureWindowSlope)
}

func (agg *FeatureAggregator) MergeProcessor(other bitflow.SampleProcessor) bool {
	if agg2, ok := other.(*FeatureAggregator); !ok {
		return false
	} else {
		if agg.WindowSize != agg2.WindowSize || agg.WindowDuration != agg2.WindowDuration || agg.UseCurrentTime != agg2.UseCurrentTime {
			// This is likely not intended, since the follow-up aggregator will aggregate the already aggregated metrics
			log.Warnf("%v: Cannot merge the follow-up aggregator due to different parameters: %v", agg, agg2)
			return false
		}
		agg.aggregators = append(agg.aggregators, agg2.aggregators...)
		agg.suffixes = append(agg.suffixes, agg2.suffixes...)
		return true
	}
}

func (agg *FeatureAggregator) OutputSampleSize(sampleSize int) int {
	return sampleSize * (1 + len(agg.aggregators))
}

func (agg *FeatureAggregator) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if agg.checker.HeaderChanged(header) {
		agg.newHeader(header)
	}

	inValues := sample.Values
	sample.Resize(len(agg.outHeader.Fields))
	outValues := sample.Values[0:0] // Reuse the capacity
	for i := range header.Fields {
		stats := agg.currentHeaderStats[i]
		inValue := inValues[i]
		outValues = append(outValues, inValue)
		stats.Push(inValue, sample.Time)
		agg.flushWindow(stats)
		for _, operation := range agg.aggregators {
			outValues = append(outValues, operation(stats))
		}
	}
	sample.Values = outValues
	return agg.NoopProcessor.Sample(sample, agg.outHeader)
}

func (agg *FeatureAggregator) newHeader(header *bitflow.Header) {
	outFields := make([]string, 0, agg.OutputSampleSize(len(header.Fields)))
	agg.currentHeaderStats = agg.currentHeaderStats[0:0]
	for _, field := range header.Fields {
		outFields = append(outFields, field)
		for _, suffix := range agg.suffixes {
			outFields = append(outFields, field+suffix)
		}
		agg.currentHeaderStats = append(agg.currentHeaderStats, agg.getWindow(field))
	}
	agg.outHeader = header.Clone(outFields)
	log.Println(agg, "increasing header from", len(header.Fields), "to", len(outFields))
}

func (agg *FeatureAggregator) getWindow(field string) *FeatureWindowStats {
	if agg.allStats == nil {
		agg.allStats = make(map[string]*FeatureWindowStats)
	}
	stats, ok := agg.allStats[field]
	if !ok {
		stats = &FeatureWindowStats{}
		agg.allStats[field] = stats
	}
	return stats
}

func (agg *FeatureAggregator) flushWindow(stats *FeatureWindowStats) {
	if agg.WindowSize > 0 && stats.num > agg.WindowSize {
		stats.Flush(stats.num - agg.WindowSize)
	}
	if agg.WindowDuration > 0 && stats.num > 0 {
		var now time.Time
		if agg.UseCurrentTime {
			now = time.Now()
		} else {
			now = stats.timestamps.Back().Value.(time.Time)
		}

		i := 0
		for link := stats.timestamps.Front(); link != nil; link = link.Next() {
			timestamp := link.Value.(time.Time)
			diff := now.Sub(timestamp)
			if diff >= agg.WindowDuration || diff < 0 { // Also flush illegal "future" timestamps
				i++
			} else {
				break
			}
		}
		stats.Flush(i)
	}
}

func (agg *FeatureAggregator) String() string {
	desc := "Feature Aggregator"
	if agg.WindowSize > 0 {
		desc += " [" + strconv.Itoa(agg.WindowSize) + " samples]"
	}
	if agg.WindowDuration > 0 {
		desc += " [" + agg.WindowDuration.String()
		if agg.UseCurrentTime {
			desc += " from current time"
		}
		desc += "]"
	}
	return fmt.Sprintf("%s %v", desc, agg.suffixes)
}

type FeatureWindowStats struct {
	sum        bitflow.Value
	num        int
	values     list.List
	timestamps list.List
}

func (stats *FeatureWindowStats) Push(val bitflow.Value, timestamp time.Time) {
	stats.sum += val
	stats.num++
	stats.values.PushBack(val)
	stats.timestamps.PushBack(timestamp)
}

func (stats *FeatureWindowStats) Flush(num int) {
	flushedSum := bitflow.Value(0)
	i := 0
	link := stats.values.Front()
	for link != nil && i < num {
		stats.timestamps.Remove(stats.timestamps.Front())
		val := stats.values.Remove(link).(bitflow.Value)
		flushedSum += val
		i++
		link = stats.values.Front()
	}
	stats.sum = stats.sum - flushedSum
	stats.num = stats.num - i
}

func FeatureWindowAverage(stats *FeatureWindowStats) bitflow.Value {
	if stats.num == 0 {
		return 0
	}
	return stats.sum / bitflow.Value(stats.num)
}

func FeatureWindowSlope(stats *FeatureWindowStats) bitflow.Value {
	if stats.num <= 1 {
		return 0
	}
	front := stats.values.Front().Value.(bitflow.Value)
	back := stats.values.Back().Value.(bitflow.Value)
	return back - front
}

func RegisterAggregateAvg(b builder.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("avg",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			agg, err := create_aggregator(params)
			if err != nil {
				return err
			}
			p.Add(agg.AddAvg("_avg"))
			return nil
		},
		"Add an average metric for every incoming metric. Optional parameter: duration or number of samples", builder.OptionalParams("window"))
}

func RegisterAggregateSlope(b builder.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("slope",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			agg, err := create_aggregator(params)
			if err != nil {
				return err
			}
			p.Add(agg.AddSlope("_slope"))
			return nil
		},
		"Add a slope metric for every incoming metric. Optional parameter: duration or number of samples", builder.OptionalParams( "window"))
}

func create_aggregator(params map[string]string) (*FeatureAggregator, error) {
	window, haveWindow := params["window"]
	if !haveWindow {
		return &FeatureAggregator{}, nil
	}
	dur, err1 := time.ParseDuration(window)
	if err1 == nil {
		return &FeatureAggregator{WindowDuration: dur}, nil
	}
	num, err2 := strconv.Atoi(window)
	if err2 == nil {
		return &FeatureAggregator{WindowSize: num}, nil
	}
	return nil, builder.ParameterError("window", golib.MultiError{err1, err2})
}
