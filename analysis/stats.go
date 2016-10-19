package analysis

import (
	"math"
	"sort"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go"
	"github.com/antongulenko/go-onlinestats"
	"github.com/go-ini/ini"
)

type StoreStats struct {
	AbstractProcessor
	TargetFile string

	stats map[string]*FeatureStats
}

func NewStoreStats(targetFile string) *StoreStats {
	return &StoreStats{
		TargetFile: targetFile,
		stats:      make(map[string]*FeatureStats),
	}
}

type FeatureStats struct {
	onlinestats.Running
	Min float64
	Max float64
}

func NewFeatureStats() *FeatureStats {
	return &FeatureStats{
		Min: math.MaxFloat64,
		Max: -math.MaxFloat64,
	}
}

func (stats *FeatureStats) Push(values ...float64) {
	for _, value := range values {
		stats.Running.Push(value)
		stats.Min = math.Min(stats.Min, value)
		stats.Max = math.Max(stats.Max, value)
	}
}

func (stats *StoreStats) Sample(inSample *data2go.Sample, header *data2go.Header) error {
	if err := stats.Check(inSample, header); err != nil {
		return err
	}
	for index, field := range header.Fields {
		val := inSample.Values[index]
		feature, ok := stats.stats[field]
		if !ok {
			feature = NewFeatureStats()
			stats.stats[field] = feature
		}
		feature.Push(float64(val))
	}
	return stats.OutgoingSink.Sample(inSample, header)
}

func (stats *StoreStats) Close() {
	defer stats.CloseSink(nil)
	if err := stats.StoreStatistics(); err != nil {
		log.Println("Error storing feature statistics:", err)
		stats.Error(err)
	}
}

func (stats *StoreStats) StoreStatistics() error {
	printFloat := func(val float64) string {
		return strconv.FormatFloat(val, 'g', -1, 64)
	}

	cfg := ini.Empty()
	for _, name := range stats.sortedFeatures() {
		feature := stats.stats[name]
		section := cfg.Section(name)
		section.NewKey("avg", printFloat(feature.Mean()))
		section.NewKey("stddev", printFloat(feature.Stddev()))
		section.NewKey("count", strconv.FormatUint(uint64(feature.Len()), 10))
		section.NewKey("min", printFloat(feature.Min))
		section.NewKey("max", printFloat(feature.Max))
	}
	return cfg.SaveTo(stats.TargetFile)
}

func (stats *StoreStats) sortedFeatures() []string {
	features := make([]string, 0, len(stats.stats))
	for name := range stats.stats {
		features = append(features, name)
	}
	sort.Strings(features)
	return features
}

func (stats *StoreStats) String() string {
	return "Store Statistics (to " + stats.TargetFile + ")"
}
