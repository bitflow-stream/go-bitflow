package recovery

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
	"github.com/antongulenko/go-bitflow-pipeline/clustering"
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
	randSeed := reg.IntParam(params, "rand-seed", 1, true, &err)
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

	log.Println("===================== SELECTION DONE =====================")
	for _, recovery := range rankedRecoveries {
		log.Printf("Score %v: %v", recovery.Rank, recovery.Item)
	}

	return rankedRecoveries[0].Item.(weightedRecoveryInfo).name
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
	normalizeRanks(rankedEvents)
	rankedEvents.SortReverse() // TODO maybe cut off events with a similarity that is too low
	return rankedEvents
}

type weightedRecoveryInfo struct {
	name string

	overallWeight float64
	successWeight float64

	executionSuccess float64
	recoverySuccess  float64 // Recovery efficiency on successful execution
	overallSuccess   float64 // Successful execution and recovery

	duration           float64 // Seconds, for all executions
	successfulDuration float64 // Seconds, in case of successful execution

	// Un-weighted counters
	executions           int
	successfulExecutions int
	successfulRecoveries int
}

func (i weightedRecoveryInfo) String() string {
	return fmt.Sprintf("%v: %v/%v executions successful (%v recoveries) execution success %.2f%%, recovery success %.2f%%, overall success %.2f%%, duration %.2fs, success duration %.2fs",
		i.name, i.successfulExecutions, i.executions, i.successfulRecoveries, i.executionSuccess*100, i.recoverySuccess*100, i.overallSuccess*100, i.duration, i.successfulDuration)
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
	for _, event := range rankedEvents {
		anomaly := event.Item.(*AnomalyEvent)
		weight := event.Rank
		executions := history.GetExecutions(anomaly)
		for _, execution := range executions {
			index, ok := recoveryNames[execution.Recovery]
			if !ok {
				// The recovery has been executed in the past, but is not available for the presently given anomaly
				continue
			}
			rec := &recoveries[index]

			executionSuccess := execution.Error == nil
			overallSuccess := executionSuccess && execution.Successful
			rec.overallWeight += weight

			rec.executionSuccess += weight * asFloat(executionSuccess)
			rec.overallSuccess += weight * asFloat(overallSuccess)
			durationSeconds := execution.Duration.Seconds()
			rec.duration += weight * durationSeconds
			rec.executions++

			if executionSuccess {
				rec.successWeight += weight
				rec.recoverySuccess += weight * asFloat(execution.Successful)
				rec.successfulDuration += weight * durationSeconds
				rec.successfulExecutions++
			}
			if overallSuccess {
				rec.successfulRecoveries++
			}
		}
	}
	for i := range recoveries {
		rec := &recoveries[i]
		rec.overallSuccess /= rec.overallWeight
		rec.executionSuccess /= rec.overallWeight
		rec.duration /= rec.overallWeight
		rec.successfulDuration /= rec.successWeight
		rec.recoverySuccess /= rec.successWeight
	}
	return recoveries
}

func asFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
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
	result := make(golib.RankedSlice, len(recoveries))
	for i, recovery := range recoveries {
		score := recovery.computeScore()
		result[i].Item = recovery
		result[i].Rank = score
	}
	result.SortReverse()
	return result
}

func (i weightedRecoveryInfo) computeScore() float64 {
	// TODO take other aspects into account: duration, curiosity

	if i.executions == 0 {
		return 0.6
	}

	return i.overallSuccess
}

func normalizeRanks(s golib.RankedSlice) {
	if len(s) == 0 {
		return
	}
	min := s[0].Rank
	max := min
	for _, item := range s[1:] {
		if item.Rank < min {
			min = item.Rank
		}
		if item.Rank > max {
			max = item.Rank
		}
	}
	diff := max - min
	for _, item := range s {
		if diff == 0 {
			item.Rank = 0.5
		} else {
			item.Rank = (item.Rank - min) / diff
		}
	}
}
