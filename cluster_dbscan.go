package analysis

import (
	"fmt"
	"log"
	"sync"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/go-DBSCAN"
	"github.com/carbocation/runningvariance"
)

const ClusterTag = "cluster"

// Implements dbscan.ClusterablePoint
type DbscanPoint struct {
	sample  *sample.Sample
	convert sync.Once
	point   []float64
}

func (p *DbscanPoint) String() string {
	return fmt.Sprintf("Point[%v](%p)", len(p.sample.Values), p)
}

func (p *DbscanPoint) GetPoint() []float64 {
	p.convert.Do(func() {
		p.point = make([]float64, len(p.sample.Values))
		for i, val := range p.sample.Values {
			p.point[i] = float64(val)
		}
	})
	return p.point
}

type DbscanBatchClusterer struct {
	Eps    float64
	MinPts int
}

func (c *DbscanBatchClusterer) cluster(points []dbscan.ClusterablePoint) [][]dbscan.ClusterablePoint {
	clusterer := dbscan.NewDBSCANClusterer(c.Eps, c.MinPts)
	return clusterer.Cluster(points)
}

func (c *DbscanBatchClusterer) printSummary(clusters [][]dbscan.ClusterablePoint) {
	var stats runningvariance.RunningStat
	for _, cluster := range clusters {
		stats.Push(float64(len(cluster)))
	}
	log.Printf("%v clusters, avg size %v, size stddev %v\n", len(clusters), stats.Mean(), stats.StandardDeviation())
}

func (c *DbscanBatchClusterer) ProcessBatch(header *sample.Header, samples []*sample.Sample) (*sample.Header, []*sample.Sample) {
	points := make([]dbscan.ClusterablePoint, len(samples))
	for i, sample := range samples {
		points[i] = &DbscanPoint{sample: sample}
	}
	clusters := c.cluster(points)
	c.printSummary(clusters)
	outSamples := make([]*sample.Sample, 0, len(samples))
	for i, clust := range clusters {
		clusterName := fmt.Sprintf("Cluster-%v", i)
		for _, p := range clust {
			point, ok := p.(*DbscanPoint)
			if !ok {
				panic(fmt.Sprintf("Wrong dbscan.ClusterablePoint implementation (%T): %v", p, p))
			}
			outSample := point.sample
			outSample.Tags[ClusterTag] = clusterName
			outSamples = append(outSamples, outSample)
		}
	}
	return header, outSamples
}

func (c *DbscanBatchClusterer) String() string {
	return fmt.Sprintf("Dbscan(eps: %v, minpts: %v)", c.Eps, c.MinPts)
}
