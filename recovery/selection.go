package recovery

import (
	"math/rand"

	"github.com/antongulenko/go-bitflow-pipeline/clustering"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
)

type Selection interface {
	SelectRecovery(node *SimilarityNode, anomalyFeatures []AnomalyFeature, possibleRecoveries []string, history History) string
}

type RandomSelection struct {
	random *rand.Rand
}

func NewRandomSelectionParams(params map[string]string) (*RandomSelection, error) {
	var err error
	randSeed := query.IntParam(params, "rand-seed", 1, true, &err)
	return NewRandomSelection(randSeed), err
}

func NewRandomSelection(seed int) *RandomSelection {
	return &RandomSelection{
		random: rand.New(rand.NewSource(int64(seed))),
	}
}

func (r *RandomSelection) SelectRecovery(node *SimilarityNode, anomalyFeatures []AnomalyFeature, possibleRecoveries []string, history History) string {
	selected := r.random.Int31n(int32(len(possibleRecoveries)))
	return possibleRecoveries[selected]
}

type RecommendingSelection struct {
}

func (r *RecommendingSelection) SelectRecovery(node *SimilarityNode, anomalyFeatures []AnomalyFeature, possibleRecoveries []string, history History) string {
	rankedEvents := r.rankHistoricEvents(history, node, anomalyFeatures)
	recoveries := r.collectWeightedRecoveryInfo(rankedEvents, history, possibleRecoveries)
	rankedRecoveries := r.rankRecoveries(recoveries)
	return rankedRecoveries[0].Item.(string)
}

func (r *RecommendingSelection) rankHistoricEvents(history History, node *SimilarityNode, anomalyFeatures []AnomalyFeature) golib.RankedSlice {

	// TODO include events of neighbor nodes (with relevance/similarity shrinking according to distance)

	// Rank all past events based on their similarity to the current event
	pastEvents := history.GetAnomalies(node.Name)
	rankedEvents := make(golib.RankedSlice, 0, len(pastEvents))
	for _, event := range pastEvents {
		if !event.IsResolved() {
			continue
		}
		similarity := r.anomalyEventSimilarity(event.Features, anomalyFeatures)
		rankedEvents.Append(similarity, event)
	}
	rankedEvents.Sort()
	return rankedEvents
}

func (r *RecommendingSelection) collectWeightedRecoveryInfo(rankedEvents golib.RankedSlice, history History, possibleRecoveries []string) []weightedRecoveryInfo {
	recoveries := make([]weightedRecoveryInfo, len(possibleRecoveries))
	recoveryNames := make(map[string]int, len(possibleRecoveries))
	for i, recoveryName := range possibleRecoveries {
		recoveries[i].name = recoveryName
		recoveryNames[recoveryName] = i
	}

	// Compute a number of weighted average values for each recovery action, using the anomaly similarity as weights
	// Values include: success rate, duration
	var totalSimilarityWeight float64
	for _, event := range rankedEvents {
		anomaly := event.Item.(*AnomalyEvent)
		weight := event.Rank
		totalSimilarityWeight += weight
		executions := history.GetExecutions(anomaly)
		for _, execution := range executions {
			index, ok := recoveryNames[execution.Recovery]
			if !ok {
				// The recovery has been executed in the past, but is not available for the presently given anomaly
				continue
			}

			// Success rate
			var successFactor float64
			if execution.Successful && execution.Error == nil {
				successFactor = 1
			} else {
				successFactor = 0
			}
			recoveries[index].success += successFactor * weight

			// Duration
			durationSeconds := execution.Duration.Seconds()
			recoveries[index].duration += durationSeconds * weight
		}
	}
	for index := range recoveries {
		recoveries[index].success /= totalSimilarityWeight
		recoveries[index].duration /= totalSimilarityWeight
	}
	return recoveries
}

func (r *RecommendingSelection) anomalyEventSimilarity(anomaly1 []AnomalyFeature, anomaly2 []AnomalyFeature) float64 {
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
		return 0
	}

	// Result in range [-1 .. 1]
	return clustering.CosineSimilarity(values1, values2)
}

func (r *RecommendingSelection) rankRecoveries(recoveries []weightedRecoveryInfo) golib.RankedSlice {
	result := make(golib.RankedSlice, 0, len(recoveries))
	for i, recovery := range recoveries {
		score := r.computeRecoveryScore(recovery)
		result[i].Item = recovery.name
		result[i].Rank = score
	}
	return result
}

func (r *RecommendingSelection) computeRecoveryScore(recovery weightedRecoveryInfo) float64 {

	// TODO take other aspects into account: duration, curiosity

	return recovery.success
}

type weightedRecoveryInfo struct {
	name     string
	success  float64
	duration float64 // Seconds
}
