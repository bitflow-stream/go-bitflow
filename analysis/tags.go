package analysis

import "strconv"

const (
	ClassTag  = "cls"
	SourceTag = "src"

	ClusterTag    = "cluster"
	ClusterPrefix = "Cluster-"

	ClusterUnclassified = 0
	ClusterNoise        = -1
)

func ClusterName(clusterNum int) string {
	return ClusterPrefix + strconv.Itoa(clusterNum)
}
