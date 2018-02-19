package denstream

import "time"

var _ ClusterSpace = new(LinearClusterSpace)

type LinearClusterSpace struct {
	clusters      map[*BasicMicroCluster]bool
	nextClusterId int
}

func NewLinearClusterSpace(numDimensions, minChildren, maxChildren int) *LinearClusterSpace {
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

func (s *LinearClusterSpace) NearestCluster(point []float64) MicroCluster {

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
	if _, ok := s.clusters[cluster]; !ok {
		panic("Cluster not in cluster space during: " + reason)
	}
	delete(s.clusters, cluster)
}

func (s *LinearClusterSpace) TransferCluster(cluster MicroCluster, otherSpace ClusterSpace) {
	s.Delete(clust, "transfering")
	newSpace.Insert(clust)
}

func (s *LinearClusterSpace) UpdateCluster(cluster MicroCluster, do func() (reinsertCluster bool)) {
	basic := clust.(*BasicMicroCluster)
	s.Delete(basic, "update")
	if do() {
		s.clusters[basic] = true
	}
}
