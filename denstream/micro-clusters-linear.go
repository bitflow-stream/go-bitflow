package denstream

import (
	"time"
)

var _ ClusterSpace = new(LinearClusterSpace)

type LinearClusterSpace struct {
	clusters      map[*BasicMicroCluster]bool
	nextClusterId int
}

func NewLinearClusterSpace() *LinearClusterSpace {
	res := new(LinearClusterSpace)
	res.Init()
	return res
}

func (s *LinearClusterSpace) Init() {
	s.clusters = make(map[*BasicMicroCluster]bool)
	s.nextClusterId = 0
}

func (s *LinearClusterSpace) NumClusters() int {
	return len(s.clusters)
}

func (s *LinearClusterSpace) NearestCluster(point []float64) (nearestCluster MicroCluster) {
	var closestDistance float64
	for clust := range s.clusters {
		// dist can be negative, if the point is inside a cluster
		dist := euclideanDistance(point, clust.Center()) - clust.Radius()
		if nearestCluster == nil || dist < closestDistance {
			nearestCluster = clust
			closestDistance = dist
		}
	}
	return
}

func (s *LinearClusterSpace) ClustersDo(do func(cluster MicroCluster)) {
	for cluster := range s.clusters {
		do(cluster)
	}
}

func (s *LinearClusterSpace) NewCluster(point []float64, creationTime time.Time) MicroCluster {
	clust := &BasicMicroCluster{
		cf1:          make([]float64, len(point)),
		cf2:          make([]float64, len(point)),
		creationTime: creationTime,
	}
	clust.Merge(point)
	s.Insert(clust)
	return clust
}

func (s *LinearClusterSpace) Insert(cluster MicroCluster) {
	basic := cluster.(*BasicMicroCluster)
	s.clusters[basic] = true
	basic.id = s.nextClusterId
	s.nextClusterId++
}

func (s *LinearClusterSpace) Delete(cluster MicroCluster, reason string) {
	b := cluster.(*BasicMicroCluster)
	if _, ok := s.clusters[b]; !ok {
		panic("Cluster not in cluster space during: " + reason)
	}
	delete(s.clusters, b)
}

func (s *LinearClusterSpace) TransferCluster(cluster MicroCluster, otherSpace ClusterSpace) {
	s.Delete(cluster, "transfering")
	otherSpace.Insert(cluster)
}

func (s *LinearClusterSpace) UpdateCluster(cluster MicroCluster, do func() (reinsertCluster bool)) {
	basic := cluster.(*BasicMicroCluster)
	s.Delete(basic, "update")
	if do() {
		s.clusters[basic] = true
	}
}
