package dbscan

import (
	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/analysis-pipeline/analysis"
)

type Point interface {
	SetCluster(id int)
	GetCluster() int
}

type SetOfPoints interface {
	RegionQuery(point Point, eps float64) map[Point]bool
	AllPoints() []Point
}

type Dbscan struct {
	Eps    float64
	MinPts int
}

func (d *Dbscan) Cluster(points SetOfPoints) {
	clusterId := 1
	for i, point := range points.AllPoints() {

		log.Println("Checking point", i)

		if point.GetCluster() == analysis.ClusterUnclassified {
			if d.expandCluster(points, point, clusterId) {
				clusterId++
			}
		}
	}
}

func (d *Dbscan) expandCluster(points SetOfPoints, point Point, clusterId int) bool {
	seeds := points.RegionQuery(point, d.Eps)
	if len(seeds) < d.MinPts {
		point.SetCluster(analysis.ClusterNoise)
		return false
	} else {
		d.setClusterIds(seeds, clusterId)
		delete(seeds, point)
		for len(seeds) > 0 {
			current := d.getAny(seeds)
			result := points.RegionQuery(current, d.Eps)
			if len(result) >= d.MinPts {
				for resultP := range result {
					resultPCluster := resultP.GetCluster()
					if resultPCluster == analysis.ClusterUnclassified || resultPCluster == analysis.ClusterNoise {
						if resultPCluster == analysis.ClusterUnclassified {
							seeds[resultP] = true
						}
						resultP.SetCluster(clusterId)
					}
				}
			}
			delete(seeds, current)
		}
		return true
	}
}

func (d *Dbscan) setClusterIds(points map[Point]bool, clusterId int) {
	for p := range points {
		p.SetCluster(clusterId)
	}
}

func (d *Dbscan) getAny(points map[Point]bool) Point {
	for p := range points {
		return p
	}
	return nil
}
