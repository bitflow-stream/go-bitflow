package pipeline

import (
	"fmt"
	"sort"
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

func (p AngleBasedSort) Len() int {
	return len(p.Points)
}

func (s AngleBasedSort) Swap(i, j int) {
	p := s.Points
	p[i], p[j] = p[j], p[i]
}

func (p AngleBasedSort) Less(i, j int) bool {
	a, b := p.Points[i], p.Points[j]
	z := crossProductZ(p.Reference, a, b)
	if z == 0 {
		// Collinear points: use distance to reference as second argument
		return p.Reference.Distance(a) < p.Reference.Distance(b)
	}
	return z > 0
}
