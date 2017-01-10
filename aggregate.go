package pipeline

import (
	"container/list"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
)

type FeatureAggregator struct {
	bitflow.AbstractProcessor
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

func (agg *FeatureAggregator) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := agg.Check(sample, header); err != nil {
		return err
	}
	if agg.checker.HeaderChanged(header) {
		agg.newHeader(header)
	}

	outValues := make([]bitflow.Value, 0, len(agg.outHeader.Fields))
	for i := range header.Fields {
		stats := agg.currentHeaderStats[i]
		inValue := sample.Values[i]
		outValues = append(outValues, inValue)
		stats.Push(inValue, sample.Time)
		agg.flushWindow(stats)
		for _, operation := range agg.aggregators {
			outValues = append(outValues, operation(stats))
		}
	}
	sample = sample.Clone()
	sample.Values = outValues
	return agg.OutgoingSink.Sample(sample, agg.outHeader)
}

func (agg *FeatureAggregator) newHeader(header *bitflow.Header) {
	outFields := make([]string, 0, len(header.Fields)*(1+len(agg.suffixes)))
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
	return "Feature Aggregator (" + strings.Join(agg.suffixes, ", ") + ")"
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
	if stats.num == 0 {
		return 0
	}
	front := stats.values.Front().Value.(bitflow.Value)
	if stats.num == 1 {
		return front
	}
	back := stats.values.Back().Value.(bitflow.Value)
	return back - front
}
