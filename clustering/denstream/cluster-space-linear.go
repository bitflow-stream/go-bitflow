package denstream

import (
	"time"

	"github.com/bitflow-stream/go-bitflow-pipeline/clustering"
)

var _ ClusterSpace = new(LinearClusterSpace)

type LinearClusterSpace struct {
	clusters      map[*clustering.Coreset]bool
	nextClusterId int
}

func NewLinearClusterSpace() *LinearClusterSpace {
	res := new(LinearClusterSpace)
	res.Init()
	return res
}

func (s *LinearClusterSpace) Init() {
	s.clusters = make(map[*clustering.Coreset]bool)
	s.nextClusterId = 0
}

func (s *LinearClusterSpace) NumClusters() int {
	return len(s.clusters)
}

func (s *LinearClusterSpace) TotalWeight() float64 {
	return 0
}
func (s *LinearClusterSpace) NearestCluster(point []float64) (nearestCluster clustering.SphericalCluster) {
	var closestDistance float64
	for clust := range s.clusters {
		// dist can be negative, if the point is inside a cluster
		dist := clustering.EuclideanDistance(point, clust.Center()) - clust.Radius()
		if nearestCluster == nil || dist < closestDistance {
			nearestCluster = clust
			closestDistance = dist
		}
	}
	return
}

func (s *LinearClusterSpace) ClustersDo(do func(cluster clustering.SphericalCluster)) {
	for cluster := range s.clusters {
		do(cluster)
	}
}

func (s *LinearClusterSpace) NewCluster(point []float64, creationTime time.Time) clustering.SphericalCluster {
	clust := clustering.NewCoreset(len(point))
	clust.Merge(point)
	s.Insert(&clust)
	return &clust
}

func (s *LinearClusterSpace) Insert(cluster clustering.SphericalCluster) {
	basic := cluster.(*clustering.Coreset)
	s.clusters[basic] = true
	basic.SetId(s.nextClusterId)
	s.nextClusterId++
}

func (s *LinearClusterSpace) Delete(cluster clustering.SphericalCluster, reason string) {
	b := cluster.(*clustering.Coreset)
	if _, ok := s.clusters[b]; !ok {
		panic("Cluster not in cluster space during: " + reason)
	}
	delete(s.clusters, b)
}

func (s *LinearClusterSpace) TransferCluster(cluster clustering.SphericalCluster, otherSpace ClusterSpace) {
	s.Delete(cluster, "transfering")
	otherSpace.Insert(cluster)
}

func (s *LinearClusterSpace) UpdateCluster(cluster clustering.SphericalCluster, do func() (reinsertCluster bool)) {
	basic := cluster.(*clustering.Coreset)
	s.Delete(basic, "update")
	if do() {
		s.clusters[basic] = true
	}
}

func (s *LinearClusterSpace) checkClusterForOpt(epsilon float64) float64 {
	return 0
}
