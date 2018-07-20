package denstream

import (
	"fmt"
	"math"
	"time"
)

type ClusterSpace interface {
	NumClusters() int

	NearestCluster(point []float64) MicroCluster
	ClustersDo(func(cluster MicroCluster))

	NewCluster(point []float64, timestamp time.Time) MicroCluster
	Insert(cluster MicroCluster)
	Delete(cluster MicroCluster, reason string)
	TransferCluster(cluster MicroCluster, otherSpace ClusterSpace)
	UpdateCluster(cluster MicroCluster, do func() (reinsertCluster bool))
}

type MicroCluster interface {
	Id() int
	W() float64
	Radius() float64
	Center() []float64
	CreationTime() time.Time

	Decay(factor float64)
	Merge(points []float64)
	CopyFrom(other MicroCluster)

	Clone() MicroCluster
}

type BasicMicroCluster struct {
	id int

	cf1 []float64
	cf2 []float64
	w   float64

	// Values computed from above values
	radius float64
	center []float64

	// Only used if this is an o-cluster (and not a p-cluster)
	creationTime time.Time
}

func (c *BasicMicroCluster) CreationTime() time.Time {
	return c.creationTime
}

func (c *BasicMicroCluster) Id() int {
	return c.id
}

func (c *BasicMicroCluster) W() float64 {
	return c.w
}

func (c *BasicMicroCluster) Radius() float64 {
	return c.radius
}

func (c *BasicMicroCluster) Center() []float64 {
	return c.center
}

func (c *BasicMicroCluster) Merge(p []float64) {
	for i, v := range p {
		c.cf1[i] += v
		c.cf2[i] += v * v
	}
	c.w++
	c.Update()
}

func (c *BasicMicroCluster) Decay(decayFactor float64) {
	for i := range c.cf1 {
		c.cf1[i] *= decayFactor
	}
	for i := range c.cf2 {
		c.cf2[i] *= decayFactor
	}
	c.w *= decayFactor
	c.Update()
}

func (c *BasicMicroCluster) Clone() MicroCluster {
	return &BasicMicroCluster{
		id:           c.id,
		cf1:          c.cf1,
		cf2:          c.cf2,
		w:            c.w,
		radius:       c.radius,
		center:       c.center,
		creationTime: c.creationTime,
	}
}

func (c *BasicMicroCluster) CopyFrom(other MicroCluster) {
	b := other.(*BasicMicroCluster)
	c.id = b.id
	c.cf1 = b.cf1
	c.cf2 = b.cf2
	c.w = b.w
	c.radius = b.radius
	c.center = b.center
	c.creationTime = b.creationTime
}

func (c *BasicMicroCluster) Update() {
	c.radius = c.computeRadius()
	c.center = c.computeCenter(c.center)
}

func (c *BasicMicroCluster) computeRadius() float64 {
	if c.w <= 0 {
		return 0
	}
	var radius float64
	for i := range c.cf1 {
		v1 := c.cf1[i] / c.w
		v2 := c.cf2[i] / c.w
		r := v2 - v1*v1
		if r < 0 {
			panic(fmt.Sprintf("Negatie radius component %v, cf1 = %v, cf2 = %v, w = %v", i, c.cf1, c.cf2, c.w))
		}
		r = math.Sqrt(r)
		if radius < r {
			// We use the largest radius of any component as the overall radius. TODO There could be other strategies.
			radius = r
		}
	}
	return radius

}

func (c *BasicMicroCluster) computeCenter(center []float64) []float64 {
	center = center[0:0]
	for _, cf1 := range c.cf1 {
		center = append(center, cf1/c.w)
	}
	if math.IsNaN(center[0]) {
		panic(fmt.Errorf("NaN in center: %v. W: %v, CF1: %v", center, c.w, c.cf1))
	}
	return center
}

func (c *BasicMicroCluster) reset() {
	c.id = -1
	c.cf1 = []float64{}
	c.cf2 = []float64{}
	c.w = 0
	c.radius = 0
	c.center = []float64{}
	c.creationTime = time.Time{}
}
func vectorLength(p []float64) float64 {
	var res float64
	for _, v := range p {
		res += v * v
	}
	return math.Sqrt(res)
}

func euclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		panic(fmt.Sprintf("Mismatched point dimensions for euclidean distance: %v vs %v", len(a), len(b)))
	}
	var res float64
	for i, v1 := range a {
		v2 := b[i]
		diff := v1 - v2
		res += diff * diff
	}
	return math.Sqrt(res)
}
