package denstream


import (
	"time"
)

var _ ClusterSpace = new(BirchTreeClusterSpace)

type BirchTreeClusterSpace struct {
	clusters      map[*BasicMicroCluster]bool
	nextClusterId int
}

func NewBirchTreeClusterSpace(numDimensions int) *BirchTreeClusterSpace {
	res := new(BirchTreeClusterSpace)
	res.Init()
	return res
}

func (s *BirchTreeClusterSpace) Init() {
	s.clusters = make(map[*BasicMicroCluster]bool)
	s.nextClusterId = 0
}

func (s *BirchTreeClusterSpace) NumClusters() int {
	return len(s.clusters)
}

func (s *BirchTreeClusterSpace) NearestCluster(point []float64) (nearestCluster MicroCluster) {
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

func (s *BirchTreeClusterSpace) ClustersDo(do func(cluster MicroCluster)) {
	for cluster := range s.clusters {
		do(cluster)
	}
}

func (s *BirchTreeClusterSpace) NewCluster(point []float64, creationTime time.Time) MicroCluster {
	clust := &BasicMicroCluster{
		cf1:          make([]float64, len(point)),
		cf2:          make([]float64, len(point)),
		creationTime: creationTime,
	}
	clust.Merge(point)
	s.Insert(clust)
	return clust
}

func (s *BirchTreeClusterSpace) Insert(cluster MicroCluster) {
	basic := cluster.(*BasicMicroCluster)
	s.clusters[basic] = true
	basic.id = s.nextClusterId
	s.nextClusterId++
}

func (s *BirchTreeClusterSpace) Delete(cluster MicroCluster, reason string) {
	b := cluster.(*BasicMicroCluster)
	if _, ok := s.clusters[b]; !ok {
		panic("Cluster not in cluster space during: " + reason)
	}
	delete(s.clusters, b)
}

func (s *BirchTreeClusterSpace) TransferCluster(cluster MicroCluster, otherSpace ClusterSpace) {
	s.Delete(cluster, "transfering")
	otherSpace.Insert(cluster)
}

func (s *BirchTreeClusterSpace) UpdateCluster(cluster MicroCluster, do func() (reinsertCluster bool)) {
	basic := cluster.(*BasicMicroCluster)
	s.Delete(basic, "update")
	if do() {
		s.clusters[basic] = true
	}
}
