package pipeline

import (
	"math"
	"math/rand"

	"github.com/antongulenko/go-bitflow"
)

type SpherePoints struct {
	bitflow.AbstractProcessor
	Radius    float64
	NumPoints int
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
	return rand.Float64() * 0.5 * math.Pi // Random angle in 0..90 degrees
}
