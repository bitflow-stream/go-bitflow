package clustering

import (
	"math"

	"gonum.org/v1/gonum/floats"
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

// aka magnitude
func VectorLength(p []float64) float64 {
	var res float64
	for _, v := range p {
		res += v * v
	}
	return math.Sqrt(res)
}

func EuclideanDistance(a, b []float64) float64 {
	return floats.Distance(a, b, 2)
}

func CosineSimilarity(a, b []float64) float64 {
	return floats.Dot(a, b) / (VectorLength(a) * VectorLength(b))
}
