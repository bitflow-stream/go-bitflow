package math

import (
	"fmt"
	"sort"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

// Graham Scan for computing the convex hull of a point set
// https://en.wikipedia.org/wiki/Graham_scan

type Point struct {
	X, Y float64
}

func (p Point) Distance(other Point) float64 {
	return (p.X-other.X)*(p.X-other.X) + (p.Y-other.Y)*(p.Y-other.Y)
}

type ConvexHull []Point

func ComputeConvexHull(points []Point) ConvexHull {
	p := ConvexHull(points)
	if len(p) < 3 {
		return p
	}
	lowest := p.ComputeLowestPoint()
	sort.Sort(AngleBasedSort{Reference: lowest, Points: p})

	hull := make(ConvexHull, 0, len(p)+1)
	p1 := p[0]
	p2 := p[1]
	p3 := p[2]
	hull = append(hull, p1)

	if lowest != p1 {
		panic(fmt.Errorf("Lowest point not at beginning of hull. Lowest %v, first: %v", lowest, p1))
	}

	for i := 2; i < len(p); {
		if isLeftTurn(p1, p2, p3) {
			hull = append(hull, p2)
			i++
			if i >= len(p) {
				if isLeftTurn(p2, p3, lowest) {
					hull = append(hull, p3)
				}
				break
			}
			p1, p2, p3 = p2, p3, p[i]
		} else {
			if len(hull) <= 1 {
				// TODO this is probably a bug, debug and fix
				log.Warnln("Illegal convex hull with", len(p), "points")
				return p
			}
			p2 = p1
			p1 = hull[len(hull)-2]
			hull = hull[:len(hull)-1]
		}
	}

	hull = append(hull, lowest) // Close the "circle"
	return hull
}

func (p ConvexHull) ComputeLowestPoint() (low Point) {
	low = p[0]
	for _, point := range p {
		if point.Y < low.Y || (point.Y == low.Y && point.X < low.X) {
			low = point
		}
	}
	return
}

func isLeftTurn(a, b, c Point) bool {
	// > 0: left turn
	// == 0: collinear points
	// < 0: right turn
	return crossProductZ(a, b, c) > 0
}

func crossProductZ(a, b, c Point) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (c.X-a.X)*(b.Y-a.Y)
}

// ==================== Sort by the angle from a reference point ====================

func SortByAngle(points []Point) []Point {
	p := ConvexHull(points)
	l := p.ComputeLowestPoint()
	sort.Sort(AngleBasedSort{Reference: l, Points: p})
	return p
}

type AngleBasedSort struct {
	Reference Point
	Points    []Point
}

func (s AngleBasedSort) Len() int {
	return len(s.Points)
}

func (s AngleBasedSort) Swap(i, j int) {
	p := s.Points
	p[i], p[j] = p[j], p[i]
}

func (s AngleBasedSort) Less(i, j int) bool {
	a, b := s.Points[i], s.Points[j]
	z := crossProductZ(s.Reference, a, b)
	if z == 0 {
		// Collinear points: use distance to reference as second argument
		return s.Reference.Distance(a) < s.Reference.Distance(b)
	}
	return z > 0
}

// ====================================== Batch processor ======================================

func BatchConvexHull(sortOnly bool) bitflow.BatchProcessingStep {
	desc := "convex hull"
	if sortOnly {
		desc += " sort"
	}
	return &bitflow.SimpleBatchProcessingStep{
		Description: desc,
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			if len(header.Fields) != 2 {
				return nil, nil, fmt.Errorf(
					"Cannot compute convex hull for %v dimension(s), only 2-dimensional data is allowed", len(header.Fields))
			}
			points := make([]Point, len(samples))
			for i, sample := range samples {
				points[i].X = float64(sample.Values[0])
				points[i].Y = float64(sample.Values[1])
			}

			var hull ConvexHull
			if sortOnly {
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
		},
	}
}

func RegisterConvexHull(b reg.ProcessorRegistry) {
	b.RegisterBatchStep("convex_hull",
		func(_ map[string]string) (bitflow.BatchProcessingStep, error) {
			return BatchConvexHull(false), nil
		},
		"Filter out the convex hull for a two-dimensional batch of samples")
}

func RegisterConvexHullSort(b reg.ProcessorRegistry) {
	b.RegisterBatchStep("convex_hull_sort",
		func(_ map[string]string) (bitflow.BatchProcessingStep, error) {
			return BatchConvexHull(true), nil
		},
		"Sort a two-dimensional batch of samples in order around their center")
}
