package denstream

import (
	"flag"
	"fmt"
	"math"
	"time"
)

var (
	ForceComponentRadius = false
	HoldZeroRadius       = false
)

// TODO move these parameters elsewhere
func init() {
	flag.BoolVar(&ForceComponentRadius, "microclusters-component-radius", false, "Always use the component radius computation for Denstream")
	flag.BoolVar(&HoldZeroRadius, "microclusters-hold-zero-radius", false, "Until the radius can be calculated, treat the radius of Denstream microclusters as zero")
}

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
	cf2 float64
	w   float64

	// Values computed from above values
	radius float64
	center []float64

	// Only used if this is an o-cluster (and not a p-cluster)
	creationTime time.Time
}

func NewBasicMicroCluster(numDimensions int) BasicMicroCluster {
	return BasicMicroCluster{
		cf1:    make([]float64, numDimensions),
		center: make([]float64, numDimensions),
	}
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
	}
	c.cf2 += vectorSquared(p, 1)
	c.w++
	c.Update()
}

func (c *BasicMicroCluster) Decay(decayFactor float64) {
	for i := range c.cf1 {
		c.cf1[i] *= decayFactor
	}
	c.cf2 *= decayFactor
	c.w *= decayFactor
	c.Update()
}

func (c *BasicMicroCluster) Clone() MicroCluster {
	res := &BasicMicroCluster{
		id:           c.id,
		cf1:          make([]float64, len(c.cf1)),
		cf2:          c.cf2,
		w:            c.w,
		radius:       c.radius,
		center:       make([]float64, len(c.center)),
		creationTime: c.creationTime,
	}
	copy(res.cf1, c.cf1)
	copy(res.center, c.center)
	return res
}

func (c *BasicMicroCluster) CopyFrom(other MicroCluster) {
	b := other.(*BasicMicroCluster)
	c.id = b.id
	c.cf2 = b.cf2
	c.w = b.w
	c.radius = b.radius
	c.creationTime = b.creationTime

	// Set len to zero and use append to reuse the preallocated memory if possible
	c.center = append(c.center[0:0], b.center...)
	c.cf1 = append(c.cf1[0:0], b.cf1...)
}

func (c *BasicMicroCluster) Update() {
	c.radius = c.computeRadius()
	c.center = c.computeCenter(c.center)
}

func (c *BasicMicroCluster) Add(other *BasicMicroCluster) {
	for i := range other.cf1 {
		c.cf1[i] += other.cf1[i]
	}
	c.cf2 += other.cf2
	c.w += other.w
	c.Update()
}

func (c *BasicMicroCluster) Subtract(other *BasicMicroCluster) {
	for i := range other.cf1 {
		c.cf1[i] -= other.cf1[i]
	}
	c.cf2 -= other.cf2
	c.w -= other.w
	c.Update()
}

func (c *BasicMicroCluster) computeRadius() float64 {
	if c.w <= 0 {
		return 0
	}
	cf2 := c.cf2 / c.w
	cf1 := vectorSquared(c.cf1, 1/c.w)
	if cf2 < cf1 {
		panic(fmt.Sprintf("Radius part negative (%v < %v), cf1 = %v, cf2 = %v, w = %v", cf2, cf1, c.cf1, c.cf2, c.w))
	} else {
		return math.Sqrt(cf2 - cf1)
	}
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
	c.cf2 = 0
	c.w = 0
	c.radius = 0
	c.center = []float64{}
	c.creationTime = time.Time{}
}

func vectorSquared(p []float64, componentFactor float64) float64 {
	// Squaring a vector means multiplying it with its transposed, resulting in a scalar.
	var res float64
	for _, v := range p {
		v *= componentFactor
		res += v * v
	}
	return res
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
