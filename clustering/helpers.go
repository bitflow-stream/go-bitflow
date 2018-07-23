package clustering

import (
	"fmt"
	"math"
)

func VectorSquared(p []float64, componentFactor float64) float64 {
	// Squaring a vector means multiplying it with its transposed, resulting in a scalar.
	var res float64
	for _, v := range p {
		v *= componentFactor
		res += v * v
	}
	return res
}

func VectorLength(p []float64) float64 {
	var res float64
	for _, v := range p {
		res += v * v
	}
	return math.Sqrt(res)
}

func EuclideanDistance(a, b []float64) float64 {
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
