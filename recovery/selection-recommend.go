package recovery

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/reg"
)

type Selection interface {
	SelectRecovery(now time.Time, node *SimilarityNode, anomalyFeatures AnomalyData, possibleRecoveries []string, history History) string
}

type RecommendingSelection struct {
	SimilarityComputation
	MaxCuriosity          float64
	CuriosityGrowthThresh float64 // This percentage of MaxCuriosity will be reached after CuriosityGrowthThresh seconds
	CuriosityGrowthTime   time.Duration

	k float64 // for curiosity calculation
}

func NewRecommendingSelection(params map[string]string) (*RecommendingSelection, error) {
	var err error
	sel := &RecommendingSelection{
		MaxCuriosity:          reg.FloatParam(params, "max-curiosity", 1, true, &err),
		CuriosityGrowthThresh: reg.FloatParam(params, "curiosity-growth", 0.99, true, &err),
		CuriosityGrowthTime:   reg.DurationParam(params, "curiosity-growth-time", 24*60*60, true, &err),
	}
	sel.SimilarityComputation.parseParams(params, &err)
	if err == nil {
		sel.k = sel.computeCuriosityK()
		if sel.k <= 0 || math.IsNaN(sel.k) || math.IsInf(sel.k, 0) {
			err = fmt.Errorf("Computation of curiosity growth rate K resulted in non-positive value %.2f. Choose curiosity-growth > 0.5 and bigger curiosity-growth-time", sel.k)
		}
	}
	return sel, err
}

func (r *RecommendingSelection) SelectRecovery(now time.Time, node *SimilarityNode, anomalyFeatures AnomalyData, possibleRecoveries []string, history History) string {
	events := history.GetExecutions(node.Name)

	// TODO include events of neighbor nodes (with relevance/similarity shrinking according to distance)
	// TODO maybe cut off events with a similarity that is too low

	rankedEvents := r.rankHistoricEvents(events, anomalyFeatures)
	recoveries := r.collectWeightedRecoveryInfo(rankedEvents, history, possibleRecoveries)
	r.normalizeRecoveryInfo(recoveries)
	rankedRecoveries := r.rankRecoveries(recoveries, now)

	log.Println("===================== SELECTION DONE =====================")
	for _, recovery := range rankedRecoveries {
		log.Printf("Score %v: %v", recovery.Rank, recovery.Item)
	}

	return rankedRecoveries[0].Item.(*weightedRecoveryInfo).name
}

type weightedRecoveryInfo struct {
	name string

	lastExecution         time.Time
	lastExecutionDuration time.Duration
	lastComputedCuriosity float64

	all       info // All executions of this recovery, including errors
	executed  info // All successful executions without errors (but including executions that did not recover the anomaly)
	recovered info // Successful executions that did actually recover the anomaly
}

type info struct {
	// Weighted stats and averages
	weight            float64
	executionSuccess  float64
	recoverySuccess   float64
	executionDuration float64
	overallDuration   float64

	// Un-weighted counters and averages
	count                    int
	averageExecutionDuration float64
	averageOverallDuration   float64
}

func (i *weightedRecoveryInfo) String() string {
	return fmt.Sprintf("%v (w=%.2f, dur=%v, cur: %.2f) %v/%v exec (%.2f%% -> %.2f%%), recovered %v/%v (%.2f%% -> %.2f%%), duration %.2fs -> %.2fs, success duration %.2fs -> %.2fs",
		i.name, i.all.weight, i.lastExecutionDuration, i.lastComputedCuriosity,
		i.executed.count, i.all.count, 100*float64(i.executed.count)/float64(i.all.count), i.all.executionSuccess*100,
		i.recovered.count, i.executed.count, 100*float64(i.recovered.count)/float64(i.executed.count), i.all.recoverySuccess*100,
		i.all.averageExecutionDuration, i.all.executionDuration, i.recovered.averageOverallDuration, i.recovered.overallDuration)
}

func (i *weightedRecoveryInfo) collectMinMax(min, max *weightedRecoveryInfo) {
	i.all.collectMinMax(&min.all, &max.all)
	i.executed.collectMinMax(&min.executed, &max.executed)
	i.recovered.collectMinMax(&min.recovered, &max.recovered)
}

func (i *weightedRecoveryInfo) normalize(min, max *weightedRecoveryInfo) {
	i.all.normalize(&min.all, &max.all)
	i.executed.normalize(&min.executed, &max.executed)
	i.recovered.normalize(&min.recovered, &max.recovered)
}

func (i *info) Add(weight float64, event *ExecutionEvent) {
	i.weight += weight
	i.count++
	if event.Successful {
		i.recoverySuccess += weight
	}
	if event.Error == nil {
		i.executionSuccess += weight
	}
	execSeconds := event.ExecutionDuration.Seconds()
	overallSeconds := event.Ended.Sub(event.Started).Seconds()
	i.executionDuration += weight * execSeconds
	i.overallDuration += weight * overallSeconds
	i.averageExecutionDuration += execSeconds
	i.averageOverallDuration += overallSeconds
}

func (i *info) Finalize() {
	if i.weight > 0 {
		i.recoverySuccess /= i.weight
		i.executionSuccess /= i.weight
		i.executionDuration /= i.weight
		i.overallDuration /= i.weight
	}
	if i.count > 0 {
		i.averageExecutionDuration /= float64(i.count)
		i.averageOverallDuration /= float64(i.count)
	}
}

func (i *info) collectMinMax(min, max *info) {
	min.weight = math.Min(i.weight, min.weight)
	min.executionSuccess = math.Min(i.executionSuccess, min.executionSuccess)
	min.recoverySuccess = math.Min(i.recoverySuccess, min.recoverySuccess)
	min.executionDuration = math.Min(i.executionDuration, min.executionDuration)
	min.overallDuration = math.Min(i.overallDuration, min.overallDuration)
	min.averageExecutionDuration = math.Min(i.averageExecutionDuration, min.averageExecutionDuration)
	min.averageOverallDuration = math.Min(i.averageOverallDuration, min.averageOverallDuration)
	max.weight = math.Max(i.weight, max.weight)
	max.executionSuccess = math.Max(i.executionSuccess, max.executionSuccess)
	max.recoverySuccess = math.Max(i.recoverySuccess, max.recoverySuccess)
	max.executionDuration = math.Max(i.executionDuration, max.executionDuration)
	max.overallDuration = math.Max(i.overallDuration, max.overallDuration)
	max.averageExecutionDuration = math.Max(i.averageExecutionDuration, max.averageExecutionDuration)
	max.averageOverallDuration = math.Max(i.averageOverallDuration, max.averageOverallDuration)
}

func (i *info) normalize(min, max *info) {
	i.weight = (i.weight - min.weight) / (max.weight - min.weight)
	i.executionSuccess = (i.executionSuccess - min.executionSuccess) / (max.executionSuccess - min.executionSuccess)
	i.recoverySuccess = (i.recoverySuccess - min.recoverySuccess) / (max.recoverySuccess - min.recoverySuccess)
	i.executionDuration = (i.executionDuration - min.executionDuration) / (max.executionDuration - min.executionDuration)
	i.overallDuration = (i.overallDuration - min.overallDuration) / (max.overallDuration - min.overallDuration)
	i.averageExecutionDuration = (i.averageExecutionDuration - min.averageExecutionDuration) / (max.averageExecutionDuration - min.averageExecutionDuration)
	i.averageOverallDuration = (i.averageOverallDuration - min.averageOverallDuration) / (max.averageOverallDuration - min.averageOverallDuration)
}

func (r *RecommendingSelection) collectWeightedRecoveryInfo(rankedEvents golib.RankedSlice, history History, possibleRecoveries []string) []*weightedRecoveryInfo {
	recoveries := make([]*weightedRecoveryInfo, len(possibleRecoveries))
	recoveryNames := make(map[string]int, len(possibleRecoveries))
	for i, recoveryName := range possibleRecoveries {
		recoveries[i] = &weightedRecoveryInfo{
			name: recoveryName,
		}
		recoveryNames[recoveryName] = i
	}

	// Compute a number of weighted average values for each recovery action, using the anomaly similarity as weights
	for _, event := range rankedEvents {
		execution := event.Item.(*ExecutionEvent)
		weight := event.Rank

		index, ok := recoveryNames[execution.Recovery]
		if !ok {
			// The recovery has been executed in the past, but is not available for the presently given anomaly
			continue
		}
		rec := recoveries[index]

		rec.all.Add(weight, execution)
		if execution.Error == nil {
			rec.executed.Add(weight, execution)
		}
		if execution.Successful {
			rec.recovered.Add(weight, execution)
		}
		if execution.Ended.After(rec.lastExecution) {
			rec.lastExecution = execution.Ended
		}
	}
	for i := range recoveries {
		rec := recoveries[i]
		rec.all.Finalize()
		rec.executed.Finalize()
		rec.recovered.Finalize()
	}
	return recoveries
}

func (r *RecommendingSelection) normalizeRecoveryInfo(recoveries []*weightedRecoveryInfo) {
	if len(recoveries) <= 1 {
		return
	}
	min := *recoveries[0]
	max := *recoveries[0]
	for _, i := range recoveries {
		i.collectMinMax(&min, &max)
	}
	for _, i := range recoveries {
		i.normalize(&min, &max)
	}
}

func (r *RecommendingSelection) rankRecoveries(recoveries []*weightedRecoveryInfo, now time.Time) golib.RankedSlice {
	result := make(golib.RankedSlice, len(recoveries))
	for i, recovery := range recoveries {
		score := recovery.computeScore(r, now)
		result[i].Item = recovery
		result[i].Rank = score
	}
	result.SortReverse()
	return result
}

func (r *RecommendingSelection) computeCuriosityK() float64 {
	// Derived from the formula for computeCuriosity(), using f(x)=p*MAX
	// K = log( (1-p)/p ) / -(MAX*TIME)

	return math.Log((1-r.CuriosityGrowthThresh)/r.CuriosityGrowthThresh) / -(r.MaxCuriosity * r.curiosityTime(r.CuriosityGrowthTime))
}

func (r *RecommendingSelection) curiosityTime(d time.Duration) float64 {
	// Should probably be changed to hours or days for a real scenario
	return d.Seconds()
}

func (i *weightedRecoveryInfo) computeScore(r *RecommendingSelection, now time.Time) float64 {
	dur1 := i.all.executionDuration
	dur2 := i.recovered.overallDuration
	if dur1 == 0 {
		// If there are no previous executions, use the maximum execution time (since it is normalized in 0..1)
		dur1 = 1
	}
	if dur2 == 0 {
		// If there are no previous executions, use the maximum execution time (since it is normalized in 0..1)
		dur2 = 1
	}
	// TODO scale duration logarithmically to reduce impact on overall score
	return (i.all.recoverySuccess + i.computeCuriosity(r, now)) / (dur1 + dur2)
}

func (i *weightedRecoveryInfo) computeCuriosity(r *RecommendingSelection, now time.Time) float64 {
	// f(t) = MAX / (1 + e^( -MAX*K*t ))
	// MAX is the upper bound for curiosity, K is computed in computeCuriosityK() and determines the growth rate

	// TODO the lastExecution takes into account executions during any anomaly. Possible to compute weighted average based on similarity?

	dur := now.Sub(i.lastExecution)
	t := r.curiosityTime(dur)
	i.lastExecutionDuration = dur
	cur := r.MaxCuriosity / (1 + math.Exp(-r.MaxCuriosity*r.k*t))
	i.lastComputedCuriosity = cur
	return cur
}
