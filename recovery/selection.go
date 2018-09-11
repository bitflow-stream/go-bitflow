package recovery

import (
	"math/rand"

	"github.com/antongulenko/go-bitflow-pipeline/query"
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
