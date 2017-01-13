package pipeline

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

// ====================================== Generate random points on the hull of a sphere ======================================

type SpherePoints struct {
	bitflow.AbstractProcessor
	RandomSeed int64
	Radius     float64
	NumPoints  int

	rand *rand.Rand
}

func (p *SpherePoints) Start(wg *sync.WaitGroup) golib.StopChan {
	p.rand = rand.New(rand.NewSource(p.RandomSeed))
	return p.AbstractProcessor.Start(wg)
}

func (p *SpherePoints) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	for i := 0; i < p.NumPoints; i++ {
		out := sample.Clone()
		out.Values = p.randomSpherePoint(sample.Values)
		if err := p.OutgoingSink.Sample(out, header); err != nil {
			return err
		}
	}
	return nil
}

// https://de.wikipedia.org/wiki/Kugelkoordinaten#Verallgemeinerung_auf_n-dimensionale_Kugelkoordinaten
func (p *SpherePoints) randomSpherePoint(values []bitflow.Value) []bitflow.Value {
	sinValues := make([]float64, len(values))
	cosValues := make([]float64, len(values))
	for i := range values {
		angle := p.randomAngle()
		sinValues[i] = math.Sin(angle)
		cosValues[i] = math.Cos(angle)
	}

	// Calculate point for a sphere around the point (0, 0, 0, ...)
	result := make([]bitflow.Value, len(values))
	for i := range values {
		coord := p.Radius
		for j := 0; j < i; j++ {
			coord *= sinValues[j]
		}
		if i < len(values)-1 {
			coord *= cosValues[i]
		}
		result[i] = bitflow.Value(coord)
	}
	// Move the point so it is part of the sphere around the given center
	for i, val := range values {
		result[i] += val
	}
	return result
}

func (p *SpherePoints) randomAngle() float64 {
	return p.rand.Float64() * 2 * math.Pi // Random angle in 0..90 degrees
}

// ====================================== Filter out points that are not on the convex hull of a point set ======================================

type BatchConvexHull struct {
	Sort bool
}

func (b *BatchConvexHull) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	if len(header.Fields) != 2 {
		return nil, nil, fmt.Errorf("Cannot compute convex hull for %v dimensions, only 2-dimensional data is allowed", len(header.Fields))
	}
	points := make([]Point, len(samples))
	for i, sample := range samples {
		points[i].X = float64(sample.Values[0])
		points[i].Y = float64(sample.Values[1])
	}

	var hull ConvexHull
	if b.Sort {
		hull = SortByAngle(points)
	} else {
		hull = ComputeConvexHull(points)
	}

	for i, point := range hull {
		samples[i].Values[0] = bitflow.Value(point.X)
		samples[i].Values[1] = bitflow.Value(point.Y)
	}
	log.Println("Convex hull reduced samples from", len(samples), "to", len(hull))
	return header, samples[:len(hull)], nil
}

func (b *BatchConvexHull) String() string {
	res := "convex hull"
	if b.Sort {
		res += " sort"
	}
	return res
}
