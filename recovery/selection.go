package recovery

import "math/rand"

type RecoverySelection interface {
	SelectRecovery(node *SimilarityNode, anomalyFeatures []AnomalyFeature, possibleRecoveries []string, history History) string
}

type RandomRecoverySelection struct {
}

func (r *RandomRecoverySelection) SelectRecovery(node *SimilarityNode, anomalyFeatures []AnomalyFeature, possibleRecoveries []string, history History) string {
	selected := rand.Int31n(int32(len(possibleRecoveries)))
	return possibleRecoveries[selected]
}
