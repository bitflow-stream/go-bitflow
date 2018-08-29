package denstream

import (
	"fmt"
	"time"

	"github.com/antongulenko/go-bitflow-pipeline/clustering"
	"github.com/dhconnelly/rtreego"
)

type RtreeClusterSpace struct {
	clusters      map[*RtreeCluster]bool
	tree          *rtreego.Rtree
	nextClusterId int
}

func NewRtreeClusterSpace(numDimensions, minChildren, maxChildren int) *RtreeClusterSpace {
	res := new(RtreeClusterSpace)
	res.Init(numDimensions, minChildren, maxChildren)
	return res
}

func (s *RtreeClusterSpace) Init(numDimensions, minChildren, maxChildren int) {
	s.clusters = make(map[*RtreeCluster]bool)
	s.tree = rtreego.NewTree(numDimensions, minChildren, maxChildren)
}

func (s *RtreeClusterSpace) NumClusters() int {
	return len(s.clusters)
}

func (s *RtreeClusterSpace) TreeSize() int {
	return s.tree.Size()
}

func (s *RtreeClusterSpace) NearestCluster(p []float64) clustering.SphericalCluster {
	point := make(rtreego.Point, len(p))
	for i, v := range p {
		point[i] = v
	}
	res := s.tree.NearestNeighbor(point)
	if res != nil {
		return res.(clustering.SphericalCluster)
	}
	return nil
}

func (s *RtreeClusterSpace) TransferCluster(clust clustering.SphericalCluster, newSpace ClusterSpace) {
	s.Delete(clust, "transfering")
	newSpace.Insert(clust)
}

func (s *RtreeClusterSpace) NewCluster(point []float64, creationTime time.Time) clustering.SphericalCluster {
	clust := &RtreeCluster{
		Coreset: clustering.NewCoreset(len(point)),
	}
	clust.Merge(point)
	s.Insert(clust)
	return clust
}

func (s *RtreeClusterSpace) ClustersDo(do func(clustering.SphericalCluster)) {
	for clust := range s.clusters {
		do(clust)
	}
}

func (s *RtreeClusterSpace) Insert(clust clustering.SphericalCluster) {
	r := clust.(*RtreeCluster)
	s.clusters[r] = true
	s.tree.Insert(r)
	r.SetId(s.nextClusterId)
	s.nextClusterId++
}

func (s *RtreeClusterSpace) DeleteFromTree(clust clustering.SphericalCluster, reason string) {
	c := clust.(*RtreeCluster)
	if !s.clusters[c] {
		panic("cluster is not in micro-cluster-space during: " + reason)
	}
	if !s.tree.DeleteWithComparator(c, s.clusterDeleteComparator) {
		panic("cluster not found in tree during: " + reason)
	}
}

func (s *RtreeClusterSpace) Delete(clust clustering.SphericalCluster, reason string) {
	s.DeleteFromTree(clust, reason)
	delete(s.clusters, clust.(*RtreeCluster))
}

// See clustering.SphericalCluster.Invalidate() for the return value
func (s *RtreeClusterSpace) UpdateCluster(clust clustering.SphericalCluster, do func() bool) {
	r := clust.(*RtreeCluster)
	s.DeleteFromTree(r, "update")
	if do() {
		s.tree.Insert(r)
	}
}

func (s *RtreeClusterSpace) checkClusterForOpt(epsilon float64) float64 {
	return 0
}

func (s *RtreeClusterSpace) clusterDeleteComparator(obj1, obj2 rtreego.Spatial) bool {
	return obj1.(*RtreeCluster) == obj2.(*RtreeCluster)
}

type RtreeCluster struct {
	clustering.Coreset
	bounds *rtreego.Rect
}

func (c *RtreeCluster) Bounds() *rtreego.Rect {
	return c.bounds
}

func (c *RtreeCluster) Merge(point []float64) {
	c.Coreset.Merge(point)
	c.updateBounds()
}

func (c *RtreeCluster) Decay(factor float64) {
	c.Coreset.Decay(factor)
	c.updateBounds()
}

func (c *RtreeCluster) CopyFrom(other clustering.SphericalCluster) {
	otherRtree := other.(*RtreeCluster)
	c.Coreset.CopyFrom(&otherRtree.Coreset)
	c.bounds = otherRtree.bounds
}

func (c *RtreeCluster) Clone() clustering.SphericalCluster {
	return &RtreeCluster{
		Coreset: *c.Coreset.Clone().(*clustering.Coreset),
		bounds:  c.bounds,
	}
}

func (c *RtreeCluster) updateBounds() {
	c.bounds = c.computeBounds()
}

func (c *RtreeCluster) computeBounds() *rtreego.Rect {
	center := c.Center()
	radius := c.Radius()
	origin := make(rtreego.Point, len(center))
	lengths := make([]float64, len(center))
	for i, v := range center {
		origin[i] = v - radius
		lengths[i] = 2 * radius
	}
	bounds, err := rtreego.NewRect(origin, lengths)
	if err != nil {
		panic(fmt.Errorf("Illegal rect (W: %v, origin: %v, lengths: %v): %v", c.W(), origin, lengths, err))
	}
	return bounds
}
