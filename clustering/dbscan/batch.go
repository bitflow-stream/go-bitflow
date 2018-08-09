package dbscan

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/go-onlinestats"
	log "github.com/sirupsen/logrus"
)

type BatchClusterer struct {
	Dbscan

	TreeMinChildren int     // 25
	TreeMaxChildren int     // 50
	TreePointWidth  float64 // 0.0001
}

func (c *BatchClusterer) printSummary(clusters map[string][]*bitflow.Sample) {
	var stats onlinestats.Running
	for _, cluster := range clusters {
		stats.Push(float64(len(cluster)))
	}
	log.Printf("%v clusters, avg size %v, size stddev %v", len(clusters), stats.Mean(), stats.Stddev())
}

func (c *BatchClusterer) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	log.Println("Building RTree...")

	tree := NewRtreeSetOfPoints(len(header.Fields), c.TreeMinChildren, c.TreeMaxChildren, c.TreePointWidth)
	for i, sample := range samples {
		if i%5000 == 0 {
			log.Printf("Inserted %v out of %v samples (%.2f%%)", i, len(samples), float32(i)/float32(len(samples))*100)
		}
		tree.Add(sample)
	}

	log.Println("Clustering ...")
	clusters := tree.Cluster(&c.Dbscan)
	c.printSummary(clusters)
	return header, samples, nil
}

func (c *BatchClusterer) String() string {
	return fmt.Sprintf("Rtree-Dbscan(eps: %v, minpts: %v, tree: %v-%v, width: %v)",
		c.Eps, c.MinPts, c.TreeMinChildren, c.TreeMaxChildren, c.TreePointWidth)
}

func RegisterDbscan(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline) {
		p.Batch(&BatchClusterer{
			Dbscan:          Dbscan{Eps: 0.1, MinPts: 5},
			TreeMinChildren: 25,
			TreeMaxChildren: 50,
			TreePointWidth:  0.0001,
		})
	}
	b.RegisterAnalysis("dbscan", create, "Perform a dbscan clustering on a batch of samples")
}

func RegisterDbscanParallel(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline) {
		p.Batch(&ParallelDbscanBatchClusterer{Eps: 0.3, MinPts: 5})
	}
	b.RegisterAnalysis("dbscan_parallel", create, "Perform a parallel dbscan clustering on a batch of samples")
}
