package recovery

import (
	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/reg"
	"github.com/bitflow-stream/go-bitflow-pipeline/clustering"
	"github.com/bitflow-stream/go-bitflow-pipeline/steps"
)

type SimilarityComputation struct {
	NormalizeSimilarities bool // Apply stddev-normalization to the resulting similarities

	ScaleStddev     bool // Apply normalization to the input data prior to computing similarities
	ScaleMinMax     bool
	ScaleMinMaxFrom float64
	ScaleMinMaxTo   float64
}

func (c *SimilarityComputation) parseParams(params map[string]string, err *error) {
	c.NormalizeSimilarities = reg.BoolParam(params, "similarity-normalize-similarities", false, true, err)
	c.ScaleStddev = reg.BoolParam(params, "similarity-scale-stddev", false, true, err)
	c.ScaleMinMax = reg.BoolParam(params, "similarity-scale-min-max", false, true, err)
	c.ScaleMinMaxFrom = reg.FloatParam(params, "similarity-scale-from", 0, true, err)
	c.ScaleMinMaxTo = reg.FloatParam(params, "similarity-scale-to", 1, true, err)
}

func (c *SimilarityComputation) rankHistoricEvents(eventData []*ExecutionEvent, targetEventData AnomalyData) golib.RankedSlice {
	targetFeatures := copyAnomalyFeatures(targetEventData)
	events := make([]AnomalyData, len(eventData))
	for i, event := range eventData {
		events[i] = copyAnomalyFeatures(event.AnomalyFeatures)
	}

	c.preprocess(append(events, targetFeatures))
	similarities := c.computeSimilarities(events, targetFeatures)
	c.normalizeSimilarities(similarities)

	rankedEvents := make(golib.RankedSlice, 0, len(events))
	for i, event := range events {
		rankedEvents.Append(similarities[i], event)
	}
	rankedEvents.SortReverse()
	return rankedEvents
}

func copyAnomalyFeatures(f AnomalyData) AnomalyData {
	res := make(AnomalyData, len(f))
	copy(res, f)
	return res
}

func (c *SimilarityComputation) computeSimilarities(events []AnomalyData, target AnomalyData) []float64 {
	res := make([]float64, len(events))
	for i, event := range events {
		res[i] = c.anomalyEventSimilarity(target, event)
	}
	return res
}

func (c *SimilarityComputation) preprocess(events []AnomalyData) {
	if !(c.ScaleStddev || c.ScaleMinMax) {
		return
	}

	stats := make(map[string]*steps.FeatureStats, len(events))
	for _, event := range events {
		for _, feature := range event {
			featureStats, ok := stats[feature.Name]
			if !ok {
				featureStats = new(steps.FeatureStats)
				stats[feature.Name] = featureStats
			}
			featureStats.Push(feature.Value)
		}
	}
	for _, event := range events {
		for i, feature := range event {
			if c.ScaleStddev {
				event[i].Value = stats[feature.Name].ScaleStddev(event[i].Value)
			} else if c.ScaleMinMax {
				event[i].Value = stats[feature.Name].ScaleMinMax(event[i].Value, c.ScaleMinMaxFrom, c.ScaleMinMaxTo)
			}
		}
	}
}

func (c *SimilarityComputation) normalizeSimilarities(res ...[]float64) {
	if c.NormalizeSimilarities {
		var stat steps.FeatureStats
		for _, similSlice := range res {
			for _, simil := range similSlice {
				stat.Push(simil)
			}
		}
		for i, similSlice := range res {
			for j, simil := range similSlice {
				res[i][j] = stat.ScaleStddev(simil)
			}
		}
	}
}

func (c *SimilarityComputation) anomalyEventSimilarity(anomaly1 AnomalyData, anomaly2 AnomalyData) float64 {

	// TODO include meta-features, like number of previous attempts? Difference in the features between first attempt and current attempt of running anomaly?

	map1 := make(map[string]float64, len(anomaly1))
	for _, feature := range anomaly1 {
		map1[feature.Name] = feature.Value
	}

	// Collect feature-vectors of all features shared by both anomalies. Other features are ignored.
	var values1, values2 []float64
	for _, feature2 := range anomaly2 {
		value1, ok := map1[feature2.Name]
		if !ok {
			// Only take into account features that are present in both anomalies
			continue
		}
		value2 := feature2.Value
		values1 = append(values1, value1)
		values2 = append(values2, value2)
	}
	if len(values1) == 0 {
		// Not a single feature shared by both anomalies... Return minimum possible similarity
		return -1
	}

	// Result in range [-1 .. 1]
	return clustering.CosineSimilarity(values1, values2)
}
