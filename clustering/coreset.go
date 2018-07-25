package clustering

import (
	"flag"
	"fmt"
	"math"
	"time"
)

var acceptableRadiusRoundingError = float64(0)

func init() {
	flag.Float64Var(&acceptableRadiusRoundingError, "coreset-radius-rounding", acceptableRadiusRoundingError, "Acceptable rounding error when computing the radius of coresets for Denstream")
}

type SphericalCluster interface {
	Id() int
	W() float64
	Radius() float64
	Center() []float64
	CreationTime() time.Time

	Decay(factor float64)
	Merge(points []float64)
	CopyFrom(other SphericalCluster)

	Clone() SphericalCluster
}

type Coreset struct {
	id int

	// Only used if this is an o-cluster (and not a p-cluster)
	creationTime time.Time

	cf1 []float64
	cf2 float64
	w   float64

	// Values computed from above values
	radius float64
	center []float64

	coresetDebugData
}

func NewCoreset(numDimensions int) Coreset {
	b := Coreset{
		cf1:    make([]float64, numDimensions),
		center: make([]float64, numDimensions),
	}
	b.debugCreate()
	return b
}

func (c Coreset) String() string {
	return fmt.Sprintf("[id: %v, w: %v, r: %v, center: %v%v]", c.id, c.w, c.radius, c.center, c.debugString())
}

func (c *Coreset) SetCreationTime(t time.Time) {
	c.creationTime = t
}

func (c *Coreset) CreationTime() time.Time {
	return c.creationTime
}

func (c *Coreset) SetId(id int) {
	c.id = id
}

func (c *Coreset) Id() int {
	return c.id
}

func (c *Coreset) W() float64 {
	return c.w
}

func (c *Coreset) Radius() float64 {
	return c.radius
}

func (c *Coreset) Center() []float64 {
	return c.center
}

func (c *Coreset) Merge(p []float64) {
	for i, v := range p {
		c.cf1[i] += v
	}
	c.cf2 += VectorSquared(p, 1)
	c.w++
	c.debugMerge(p)
	c.update()
}

func (c *Coreset) Subtract(p []float64) {
	for i, v := range p {
		c.cf1[i] -= v
	}
	c.cf2 -= VectorSquared(p, 1)
	c.w--
	c.debugSubtract(p)
	c.update()
}

func (c *Coreset) MergeCoreset(other *Coreset) {
	for i, v := range other.cf1 {
		c.cf1[i] += v
	}
	c.cf2 += other.cf2
	c.w += other.w
	c.debugMergeCoreset(other)
	c.update()
}

func (c *Coreset) SubtractCoreset(other *Coreset) {
	for i, v := range other.cf1 {
		c.cf1[i] -= v
	}
	c.cf2 -= other.cf2
	c.w -= other.w
	c.debugSubtractCoreset(other)
	c.update()
}

func (c *Coreset) Decay(decayFactor float64) {
	for i := range c.cf1 {
		c.cf1[i] *= decayFactor
	}
	c.cf2 *= decayFactor
	c.w *= decayFactor
	c.debugDecay(decayFactor)
	c.update()
}

func (c *Coreset) Clone() SphericalCluster {
	res := c.clone()
	res.debugClone()
	return res
}

func (c *Coreset) CopyFrom(other SphericalCluster) {
	b := other.(*Coreset)
	c.id = b.id
	c.cf2 = b.cf2
	c.w = b.w
	c.radius = b.radius
	c.creationTime = b.creationTime
	c.center = append(c.center[0:0], b.center...) // Reuse memory
	c.cf1 = append(c.cf1[0:0], b.cf1...)
	c.debugCopyFrom(other.(*Coreset))
}

func (c *Coreset) update() {
	c.radius = c.computeRadius()
	c.center = c.computeCenter(c.center)
}

func (c *Coreset) clone() *Coreset {
	res := &Coreset{
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
	res.copyDebugDataFrom(c)
	return res
}

func (c *Coreset) computeRadius() float64 {
	if c.w <= 1 {
		// Avoid computations for w ~= 1, they are zero mathematically, but sometimes cannot be computed due to inexact floating point calculations
		return 0
	}
	cf2 := c.cf2 / c.w
	cf1 := VectorSquared(c.cf1, 1/c.w)
	if cf2 < cf1 {
		if cf1-cf2 <= acceptableRadiusRoundingError {
			return 0
		}
		c.debugWrongRadiusComputation()
		panic(fmt.Sprintf("Radius part negative (%v < %v) for %v", cf2, cf1, c))
	} else {
		return math.Sqrt(cf2 - cf1)
	}
}

func (c *Coreset) computeCenter(center []float64) []float64 {
	if c.w <= 0 {
		center = center[:len(c.cf1)]
		for i := range c.cf1 {
			center[i] = 0
		}
		return center
	}

	center = center[0:0]
	for _, cf1 := range c.cf1 {
		center = append(center, cf1/c.w)
	}
	if math.IsNaN(center[0]) {
		panic(fmt.Errorf("NaN in center: %v. W: %v, CF1: %v", center, c.w, c.cf1))
	}
	return center
}
