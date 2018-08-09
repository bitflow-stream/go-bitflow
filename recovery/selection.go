package recovery

import "math/rand"

type Selection interface {
	SelectRecovery(node *SimilarityNode, anomalyFeatures []AnomalyFeature, possibleRecoveries []string, history History) string
}

type RandomSelection struct {
}

func (r *RandomSelection) SelectRecovery(node *SimilarityNode, anomalyFeatures []AnomalyFeature, possibleRecoveries []string, history History) string {
	selected := rand.Int31n(int32(len(possibleRecoveries)))
	return possibleRecoveries[selected]
}
