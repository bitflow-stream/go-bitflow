package analysis

import (
	"math"
	"sort"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go/sample"
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
	stats onlinestats.Running
	min   float64
	max   float64
}

func (stats *FeatureStats) Push(values ...float64) {
	for _, value := range values {
		stats.stats.Push(value)
		stats.min = math.Min(stats.min, value)
		stats.max = math.Max(stats.max, value)
	}
}

func (stats *StoreStats) Sample(inSample sample.Sample, header sample.Header) error {
	if err := stats.Check(inSample, header); err != nil {
		return err
	}
	for index, field := range header.Fields {
		val := inSample.Values[index]
		feature, ok := stats.stats[field]
		if !ok {
			feature = &FeatureStats{
				min: math.MaxFloat64,
				max: -math.MaxFloat64,
			}
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
		section.NewKey("avg", printFloat(feature.stats.Mean()))
		section.NewKey("stddev", printFloat(feature.stats.Stddev()))
		section.NewKey("count", strconv.FormatUint(uint64(feature.stats.Len()), 10))
		section.NewKey("min", printFloat(feature.min))
		section.NewKey("max", printFloat(feature.max))
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
