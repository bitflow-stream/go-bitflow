package denstream

import (
	"fmt"
	"strconv"

	bitflow "github.com/antongulenko/go-bitflow"
	log "github.com/sirupsen/logrus"
)

var _ bitflow.SampleProcessor = new(DenstreamClusterProcessor)

const ClusterRadiusSuffic = "/radius"

type DenstreamClusterProcessor struct {
	bitflow.NoopProcessor
	DenstreamClusterer

	// If set to >0, will log the denstream clusterer state every OutputStateModulo samples
	OutputStateModulo  int
	CreateClusterSpace func(numDimensions int) ClusterSpace

	// If set to true, this processing step will output all clusters as special samples when closing.
	// The first metric will be the radius of the cluster, followed by the cluster center in every dimension.
	FlushClustersAtClose bool

	numDimensions       int
	processedSamples    int
	lastProcessedSample *bitflow.Sample
	lastProcessedHeader *bitflow.Header
}

func (p *DenstreamClusterProcessor) String() string {
	return fmt.Sprintf("denstream (λ=%v, ε=%v, βµ=%v)", p.HistoryFading, p.Epsilon, p.MaxOutlierWeight)
}

func (p *DenstreamClusterProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if p.numDimensions == 0 {
		p.numDimensions = len(sample.Values)
		log.Printf("Initializing denstream processor to %v dimensions", p.numDimensions)
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
	p.lastProcessedSample = sample
	p.lastProcessedHeader = header
	return p.NoopProcessor.Sample(sample, header)
}

func (p *DenstreamClusterProcessor) Close() {
	if p.FlushClustersAtClose && p.lastProcessedSample != nil {
		newFields := make([]string, len(p.lastProcessedHeader.Fields)+1)
		newFields[0] = "radius"
		copy(newFields[1:], p.lastProcessedHeader.Fields)
		header := p.lastProcessedHeader.Clone(newFields)

		var err error
		p.oClusters.ClustersDo(func(cluster MicroCluster) {
			p.outputCluster(cluster, "outlier", header, &err)
		})
		p.oClusters.ClustersDo(func(cluster MicroCluster) {
			p.outputCluster(cluster, "real", header, &err)
		})
	}
	p.NoopProcessor.Close()
}

func (p *DenstreamClusterProcessor) outputCluster(cluster MicroCluster, clusterType string, header *bitflow.Header, err *error) {
	if *err != nil {
		return
	}
	center := cluster.Center()
	values := make([]bitflow.Value, len(center)+1)
	values[0] = bitflow.Value(cluster.Radius())
	for i, centerDim := range center {
		values[i+1] = bitflow.Value(centerDim)
	}

	sample := &bitflow.Sample{
		Time:   p.lastProcessedSample.Time,
		Values: values,
	}
	sample.SetTag("cluster-type", clusterType)
	sample.SetTag("cluster-id", strconv.Itoa(cluster.Id()))
	*err = p.GetSink().Sample(sample, header)
	if *err != nil {
		log.Errorln("Error outputting cluster centers:", *err)
	}
}
