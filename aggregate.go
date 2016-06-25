package analysis

import (
	"container/list"
	"strings"
	"time"

	"github.com/antongulenko/data2go/sample"
)

type FeatureAggregator struct {
	AbstractProcessor
	WindowSize     int           // Applied if >0
	WindowDuration time.Duration // Applied if >0
	UseCurrentTime bool          // If true, use time.Now() as reference for WindowTime. Otherwise, use the timestamp of the latest Sample.

	aggregators        []FeatureAggregatorOperation
	suffixes           []string
	FeatureWindowStats map[string]*FeatureWindowStats

	outHeader sample.Header
}

func NewFeatureAggregator() *FeatureAggregator {
	return &FeatureAggregator{
		FeatureWindowStats: make(map[string]*FeatureWindowStats),
	}
}

type FeatureAggregatorOperation func(stats *FeatureWindowStats) sample.Value

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

func (agg *FeatureAggregator) Sample(inSample sample.Sample, header sample.Header) error {
	if err := agg.CheckSink(); err != nil {
		return err
	}
	if err := inSample.Check(header); err != nil {
		return err
	}
	if !agg.outHeader.Equals(&header) {
		agg.buildOutHeader(header)
	}

	outValues := make([]sample.Value, len(agg.outHeader.Fields))
	for i, field := range header.Fields {
		stats := agg.getFeatureWindowStats(field)
		inValue := inSample.Values[i]
		outValues[i] = inValue
		stats.Push(inValue, inSample.Time)
		agg.flushWindow(stats)
		for j, operation := range agg.aggregators {
			outValues[i+1+j] = operation(stats)
		}
	}
	outSample := inSample.Clone()
	outSample.Values = outValues
	return agg.OutgoingSink.Sample(outSample, agg.outHeader)
}

func (agg *FeatureAggregator) buildOutHeader(header sample.Header) {
	outFields := make([]string, len(header.Fields)*(1+len(agg.suffixes)))
	for i, field := range header.Fields {
		outFields[i] = field
		for j, suffix := range agg.suffixes {
			outFields[i+1+j] = field + suffix
		}
	}
}

func (agg *FeatureAggregator) getFeatureWindowStats(field string) *FeatureWindowStats {
	// TODO optimization: arrange *FeatureWindowStats in slice in buildOutHeader, avoid frequent map lookups
	stats, ok := agg.FeatureWindowStats[field]
	if !ok {
		stats = &FeatureWindowStats{}
		agg.FeatureWindowStats[field] = stats
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
	return "Feature Aggregator: " + strings.Join(agg.suffixes, ", ")
}

type FeatureWindowStats struct {
	sum        sample.Value
	num        int
	values     list.List
	timestamps list.List
}

func (stats *FeatureWindowStats) Push(val sample.Value, timestamp time.Time) {
	stats.sum += val
	stats.num++
	stats.values.PushBack(val)
	stats.timestamps.PushBack(timestamp)
}

func (stats *FeatureWindowStats) Flush(num int) {
	flushedSum := sample.Value(0)
	i := 0
	link := stats.values.Front()
	for link != nil && i < num {
		stats.timestamps.Remove(stats.timestamps.Front())
		val := stats.values.Remove(link).(sample.Value)
		flushedSum += val
		i++
		link = stats.values.Front()
	}
	stats.sum = stats.sum - flushedSum
	stats.num = stats.num - i
}

func FeatureWindowAverage(stats *FeatureWindowStats) sample.Value {
	if stats.num == 0 {
		return 0
	}
	return stats.sum / sample.Value(stats.num)
}

func FeatureWindowSlope(stats *FeatureWindowStats) sample.Value {
	if stats.num == 0 {
		return 0
	}
	front := stats.values.Front().Value.(sample.Value)
	if stats.num == 1 {
		return front
	}
	back := stats.values.Back().Value.(sample.Value)
	return back - front
}
