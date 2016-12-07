package pipeline

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

// String is a trivial implementation of the fmt.Sringer interface
type String string

func (s String) String() string {
	return string(s)
}
