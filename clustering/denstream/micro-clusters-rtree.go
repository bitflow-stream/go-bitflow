package denstream

import (
	"fmt"
	"time"

	"github.com/dhconnelly/rtreego"
)

type RtreeClusterSpace struct {
	clusters      map[*RtreeMicroCluster]bool
	tree          *rtreego.Rtree
	nextClusterId int
}

func NewRtreeClusterSpace(numDimensions, minChildren, maxChildren int) *RtreeClusterSpace {
	res := new(RtreeClusterSpace)
	res.Init(numDimensions, minChildren, maxChildren)
	return res
}

func (s *RtreeClusterSpace) Init(numDimensions, minChildren, maxChildren int) {
	s.clusters = make(map[*RtreeMicroCluster]bool)
	s.tree = rtreego.NewTree(numDimensions, minChildren, maxChildren)
}

func (s *RtreeClusterSpace) NumClusters() int {
	return len(s.clusters)
}

func (s *RtreeClusterSpace) TreeSize() int {
	return s.tree.Size()
}

func (s *RtreeClusterSpace) NearestCluster(p []float64) MicroCluster {
	point := make(rtreego.Point, len(p))
	for i, v := range p {
		point[i] = v
	}
	res := s.tree.NearestNeighbor(point)
	if res != nil {
		return res.(MicroCluster)
	}
	return nil
}

func (s *RtreeClusterSpace) TransferCluster(clust MicroCluster, newSpace ClusterSpace) {
	s.Delete(clust, "transfering")
	newSpace.Insert(clust)
}

func (s *RtreeClusterSpace) NewCluster(point []float64, creationTime time.Time) MicroCluster {
	clust := &RtreeMicroCluster{
		BasicMicroCluster: NewBasicMicroCluster(len(point)),
	}
	clust.Merge(point)
	s.Insert(clust)
	return clust
}

func (s *RtreeClusterSpace) ClustersDo(do func(MicroCluster)) {
	for clust := range s.clusters {
		do(clust)
	}
}

func (s *RtreeClusterSpace) Insert(clust MicroCluster) {
	r := clust.(*RtreeMicroCluster)
	s.clusters[r] = true
	s.tree.Insert(r)
	r.id = s.nextClusterId
	s.nextClusterId++
}

func (s *RtreeClusterSpace) DeleteFromTree(clust MicroCluster, reason string) {
	c := clust.(*RtreeMicroCluster)
	if !s.clusters[c] {
		panic("cluster is not in micro-cluster-space during: " + reason)
	}
	if !s.tree.DeleteWithComparator(c, s.clusterDeleteComparator) {
		panic("cluster not found in tree during: " + reason)
	}
}

func (s *RtreeClusterSpace) Delete(clust MicroCluster, reason string) {
	s.DeleteFromTree(clust, reason)
	delete(s.clusters, clust.(*RtreeMicroCluster))
}

// See MicroCluster.Invalidate() for the return value
func (s *RtreeClusterSpace) UpdateCluster(clust MicroCluster, do func() bool) {
	r := clust.(*RtreeMicroCluster)
	s.DeleteFromTree(r, "update")
	if do() {
		s.tree.Insert(r)
	}
}

func (s *RtreeClusterSpace) clusterDeleteComparator(obj1, obj2 rtreego.Spatial) bool {
	return obj1.(*RtreeMicroCluster) == obj2.(*RtreeMicroCluster)
}

type RtreeMicroCluster struct {
	BasicMicroCluster
	bounds *rtreego.Rect
}

func (c *RtreeMicroCluster) Bounds() *rtreego.Rect {
	return c.bounds
}

func (c *RtreeMicroCluster) Merge(point []float64) {
	c.BasicMicroCluster.Merge(point)
	c.updateBounds()
}

func (c *RtreeMicroCluster) Decay(factor float64) {
	c.BasicMicroCluster.Decay(factor)
	c.updateBounds()
}

func (c *RtreeMicroCluster) CopyFrom(other MicroCluster) {
	otherRtree := other.(*RtreeMicroCluster)
	c.BasicMicroCluster.CopyFrom(&otherRtree.BasicMicroCluster)
	c.bounds = otherRtree.bounds
}

func (c *RtreeMicroCluster) Clone() MicroCluster {
	return &RtreeMicroCluster{
		BasicMicroCluster: *c.BasicMicroCluster.Clone().(*BasicMicroCluster),
		bounds:            c.bounds,
	}
}

func (c *RtreeMicroCluster) updateBounds() {
	c.bounds = c.computeBounds()
}

func (c *RtreeMicroCluster) computeBounds() *rtreego.Rect {
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
