package denstream

import (
	"fmt"
	"strconv"

	bitflow "github.com/antongulenko/go-bitflow"
	log "github.com/sirupsen/logrus"
)

var _ bitflow.SampleProcessor = new(DenstreamClusterProcessor)

type DenstreamClusterProcessor struct {
	bitflow.AbstractProcessor
	DenstreamClusterer

	// If set to >0, will log the denstream clusterer state every OutputStateModulo samples
	OutputStateModulo  int
	CreateClusterSpace func(numDimensions int) ClusterSpace

	numDimensions    int
	processedSamples int
}

func (p *DenstreamClusterProcessor) String() string {
	return fmt.Sprintf("denstream (λ=%v, ε=%v, βµ=%v)", p.HistoryFading, p.Epsilon, p.MaxOutlierWeight)
}

func (p *DenstreamClusterProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	if p.numDimensions == 0 {
		p.numDimensions = len(sample.Values)
		log.Printf("Initializaing denstream processor to %v dimensions", p.numDimensions)
		p.pClusters = p.CreateClusterSpace(p.numDimensions)
		p.oClusters = p.CreateClusterSpace(p.numDimensions)
	} else if p.numDimensions != len(sample.Values) {
		return fmt.Errorf("Denstream has already been initialized to %v dimensions, cannot handle sample with %v dimensions",
			p.numDimensions, len(sample.Values))
	}

	point := make([]float64, len(sample.Values))
	for i, v := range sample.Values {
		point[i] = float64(v)
	}
	clusterId := p.Insert(point, sample.Time)

	if p.OutputStateModulo > 0 && p.processedSamples%p.OutputStateModulo == 0 {
		log.Println("Denstream processed", p.processedSamples, "points in:", p.DenstreamClusterer.String())
	}
	p.processedSamples++

	sample.SetTag("cluster", strconv.Itoa(clusterId))
	return p.OutgoingSink.Sample(sample, header)
}
