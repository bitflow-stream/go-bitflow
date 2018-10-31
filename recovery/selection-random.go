package recovery

import (
	"math/rand"
	"time"

	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
)

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

func (r *RandomSelection) SelectRecovery(now time.Time, node *SimilarityNode, anomalyFeatures AnomalyData, possibleRecoveries []string, history History) string {
	selected := r.random.Int31n(int32(len(possibleRecoveries)))
	return possibleRecoveries[selected]
}
