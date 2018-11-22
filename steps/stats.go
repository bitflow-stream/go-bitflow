package steps

import (
	"sort"
	"strconv"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"
)

type StoreStats struct {
	bitflow.NoopProcessor
	TargetFile string

	stats map[string]*FeatureStats
}

func NewStoreStats(targetFile string) *StoreStats {
	return &StoreStats{
		TargetFile: targetFile,
		stats:      make(map[string]*FeatureStats),
	}
}

func RegisterStoreStats(b reg.ProcessorRegistry) {
	create := func(p *bitflow.SamplePipeline, params map[string]string) {
		p.Add(NewStoreStats(params["file"]))
	}
	b.RegisterAnalysisParams("stats", create, "Output statistics about processed samples to a given ini-file", reg.RequiredParams("file"))
}

func (stats *StoreStats) Sample(inSample *bitflow.Sample, header *bitflow.Header) error {
	for index, field := range header.Fields {
		val := inSample.Values[index]
		feature, ok := stats.stats[field]
		if !ok {
			feature = NewFeatureStats()
			stats.stats[field] = feature
		}
		feature.Push(float64(val))
	}
	return stats.NoopProcessor.Sample(inSample, header)
}

func (stats *StoreStats) Close() {
	defer stats.CloseSink()
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
		var multiErr golib.MultiError
		multiErr.AddMulti(section.NewKey("avg", printFloat(feature.Mean())))
		multiErr.AddMulti(section.NewKey("stddev", printFloat(feature.Stddev())))
		multiErr.AddMulti(section.NewKey("count", strconv.FormatUint(uint64(feature.Len()), 10)))
		multiErr.AddMulti(section.NewKey("min", printFloat(feature.Min)))
		multiErr.AddMulti(section.NewKey("max", printFloat(feature.Max)))
		if err := multiErr.NilOrError(); err != nil {
			return err
		}
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
