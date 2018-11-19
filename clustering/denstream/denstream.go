// Paper: http://www.siam.org/meetings/sdm06/proceedings/030caof.pdf
package denstream

import (
	"fmt"
	"math"
	"time"

	"github.com/bitflow-stream/go-bitflow-pipeline/clustering"
	log "github.com/sirupsen/logrus"
)

var inputCount int = 0

var dmin float64 = math.MaxFloat64
var percentageNoise = 5.0

const (
	Outlier       = -1
	NewOutlier    = -2
	OutlierStr    = "-1"
	NewOutlierStr = "-2"
)

type ClusterSpace interface {
	NumClusters() int
	TotalWeight() float64

	NearestCluster(point []float64) clustering.SphericalCluster
	ClustersDo(func(cluster clustering.SphericalCluster))

	NewCluster(point []float64, timestamp time.Time) clustering.SphericalCluster
	Insert(cluster clustering.SphericalCluster)
	Delete(cluster clustering.SphericalCluster, reason string)
	TransferCluster(cluster clustering.SphericalCluster, otherSpace ClusterSpace)
	UpdateCluster(cluster clustering.SphericalCluster, do func() (reinsertCluster bool))
	checkClusterForOpt(epsilon float64) float64
}

type Clusterer struct {
	pClusters ClusterSpace
	oClusters ClusterSpace

	HistoryFading    float64 // "λ" in the paper. >0. The higher it is, the less important are history points.
	Epsilon          float64 // "ε" in the paper. >0. The epsilon neighborhood makes core objects merge into one density area.
	MaxOutlierWeight float64 // "βµ" in the paper. [0..µ]. Minimum weight of a p-cluster.
	EnableEpsTuning  bool
	ColdStart        bool
	MinNumClusters   int

	lastPeriodicCheck     time.Time
	computedDecayInterval time.Duration
	decayCheckCounter     int
}

func (c *Clusterer) String() string {
	return fmt.Sprintf("denstream-clusterer (λ=%v, ε=%v, βµ=%v, %v p-clusters, %v o-clusters, decay checked %v times, every %v)",
		c.HistoryFading, c.Epsilon, c.MaxOutlierWeight, c.pClusters.NumClusters(), c.oClusters.NumClusters(), c.decayCheckCounter, c.computedDecayInterval)
}

// Set the lambda parameter (HistoryFading) to a value that makes the weights of all clusters decay by 1 after the given interval.
// The computation is based on the MaxOutlierWeight parameter.
func (c *Clusterer) SetDecayTimeUnit(delta time.Duration) {
	c.HistoryFading = math.Log(c.MaxOutlierWeight/(c.MaxOutlierWeight-1)) / (delta.Seconds() * 1000)
}

func (c *Clusterer) SetEpsilon(eps float64) {
	c.Epsilon = eps
}

func (c *Clusterer) Insert(point []float64, timestamp time.Time) (clusterId int) {
	inputCount++
	if c.computedDecayInterval == 0 {
		// Lazy initialize
		c.computedDecayInterval = c.decayCheckInterval()
	}

	// First insert the point into a p-cluster, an o-cluster, or create a new o-cluster
	clusterId = c.insertPoint(point)
	if clusterId <= NewOutlier {
		c.oClusters.NewCluster(point, timestamp)
	}

	// Periodically decay and check all micro-clusters
	now := timestamp
	delta := now.Sub(c.lastPeriodicCheck)
	if c.lastPeriodicCheck.IsZero() {
		c.lastPeriodicCheck = now
	} else if delta > c.computedDecayInterval {
		c.periodicCheck(now, delta)
		c.decayCheckCounter++
		c.lastPeriodicCheck = now
	}
	//Add Seema - Check periodically for epsilon after cold start
	if inputCount%10000 == 0 {
		epsilondiff := c.pClusters.checkClusterForOpt(c.Epsilon)
		log.Println("epsilon diff: ", epsilondiff)
		c.Epsilon += math.Round(epsilondiff*100) / 100
		log.Println("modified epsilon to ", c.Epsilon)
		log.Println("Pcluster: Weight and Cluster", c.pClusters.TotalWeight(), c.pClusters.NumClusters())
		log.Println("Ocluster: Weight and Cluster", c.oClusters.TotalWeight(), c.oClusters.NumClusters())
		log.Println("modified Maxoutlierweight", c.pClusters.TotalWeight()/float64(c.pClusters.NumClusters()))
		c.doMaxOutlierWeightCheck()
	}
	return
}

func (c *Clusterer) GetCluster(point []float64) (clusterId int) {
	_, nearest := c.testMergeNearest(point, c.pClusters)
	if nearest != nil {
		return nearest.Id()
	}
	_, nearest = c.testMergeNearest(point, c.oClusters)
	if nearest != nil {
		return Outlier
	}
	return NewOutlier
}

// ========================================================================================================
// ==== Internal ====
// ========================================================================================================

func (c *Clusterer) insertPoint(point []float64) int {
	// 1. try to merge into closest p-cluster. check if new radius small enough.
	if c.EnableEpsTuning && c.ColdStart {
		if c.pClusters.NumClusters()+c.oClusters.NumClusters() < c.MinNumClusters {
			// 	dist := clustering.EuclideanDistance(point, nearestClust.Center())
			// 	if dist != 0 && dist < dmin {
			// 		dmin = dist
			// 	}
			log.Println("inputcount:", inputCount, c.pClusters.NumClusters()+c.oClusters.NumClusters())
			return NewOutlier
		}
		c.doEpsilonCheck(point)

	}
	clust := c.mergeNearest(point, c.pClusters)
	if clust != nil {
		return clust.Id()
	}

	// 2. try to merge into closest o-cluster. check if new radius small enough. then check new weight: if large enough, convert to p-cluster.
	clust = c.mergeNearest(point, c.oClusters)
	if clust != nil {
		if clust.W() > c.MaxOutlierWeight {
			// Promote the o-cluster to a p-cluster
			c.oClusters.TransferCluster(clust, c.pClusters)
			return clust.Id()
		} else {
			return Outlier
		}
	}

	return NewOutlier
}

func (c *Clusterer) testMergeNearest(point []float64, space ClusterSpace) (testCluster clustering.SphericalCluster, realCluster clustering.SphericalCluster) {
	realCluster = space.NearestCluster(point)
	if realCluster != nil {

		testCluster = realCluster.Clone() // Copy the cluster to check the radius after mergin the incoming point
		testCluster.Merge(point)
		if radius := testCluster.Radius(); radius > c.Epsilon {
			realCluster = nil
			testCluster = nil
		}
	}
	return
}

func (c *Clusterer) mergeNearest(point []float64, space ClusterSpace) clustering.SphericalCluster {
	test, realCluster := c.testMergeNearest(point, space)
	if realCluster != nil {
		space.UpdateCluster(realCluster, func() bool {
			realCluster.CopyFrom(test)
			return true
		})
	}
	return realCluster
}

func (c *Clusterer) weightDecay(delta time.Duration) float64 {
	timeUnits := delta.Seconds() * 1000
	res := math.Pow(2, -c.HistoryFading*timeUnits)
	return res
}

// The interval, in which the clusters should decay their weight and be checked for too small weight
// This is the time, after which the weight of a cluster decays by 1
func (c *Clusterer) decayCheckInterval() time.Duration {
	resFloat := float64(time.Millisecond) * (1 / c.HistoryFading) * math.Log(c.MaxOutlierWeight/(c.MaxOutlierWeight-1))
	res := time.Duration(resFloat)
	if res < 0 {
		panic(fmt.Sprintf("Lambda value (%v) too low, leads to decay check interval overflow: %v", c.HistoryFading, res))
	}
	return res
}

func (c *Clusterer) periodicCheck(curTime time.Time, delta time.Duration) {
	decayFactor := c.weightDecay(delta)

	// 1. decay weight of p-clusters. delete, if too small.
	c.pClusters.ClustersDo(func(clust clustering.SphericalCluster) {
		c.pClusters.UpdateCluster(clust, func() bool {
			clust.Decay(decayFactor)
			return clust.W() >= c.MaxOutlierWeight
		})
	})

	// 2. decay weight of o-clusters. delete, if too small.
	c.oClusters.ClustersDo(func(clust clustering.SphericalCluster) {
		if clust.CreationTime().Equal(curTime) {
			// If the cluster has JUST been created, do not decay it yet
			return
		}
		c.oClusters.UpdateCluster(clust, func() bool {
			clust.Decay(decayFactor)
			t1 := curTime.Sub(clust.CreationTime()) + delta
			minWeight := (c.weightDecay(t1) - 1) / (decayFactor - 1)
			return clust.W() >= minWeight
		})
	})

}

func (c *Clusterer) adjustEpsilon() {
	epsilondiff := 0.0
	if c.pClusters.NumClusters() > c.oClusters.NumClusters() {
		epsilondiff = c.pClusters.checkClusterForOpt(c.Epsilon)
	} else {
		epsilondiff = c.oClusters.checkClusterForOpt(c.Epsilon)
	}

	log.Println("epsilon diff: ", epsilondiff)
	c.Epsilon += math.Round(epsilondiff*100) / 100
	log.Println("modified epsilon to ", c.Epsilon)
}

func (c *Clusterer) doEpsilonCheck(point []float64) { /*, nearestClust clustering.SphericalCluster) {*/
	//coldstart
	log.Println("doEpsilonCheck", c.Epsilon, inputCount)
	if inputCount == c.MinNumClusters+1 {

		c.adjustEpsilon()
		log.Println("minNumclusters reached eps=", c.Epsilon)
		// coldStart = false
	}

	if c.pClusters.NumClusters()+c.oClusters.NumClusters() >= c.MinNumClusters && (inputCount < 1000 && inputCount%100 == 0) {
		c.adjustEpsilon()
	}
	if inputCount > 1000 {
		c.ColdStart = false
	}

	// if c.pClusters.NumClusters()+c.oClusters.NumClusters() < minNumClusters {
	// 	dist := clustering.EuclideanDistance(point, nearestClust.Center())
	// 	if dist != 0 && dist < dmin {
	// 		dmin = dist
	// 	}
	// 	return NewOutlier
	// }

	//Find the minimum distance between 2 points and set epsilon as minimum distance
	/*if c.pClusters.NumClusters()+c.oClusters.NumClusters() < minNumClusters {

		if dist == 0 && dmin == math.MaxFloat64 {
			c.SetEpsilon(math.MaxFloat64)
		}
		if dist != 0 && dist < dmin {
			dmin = dist
			c.SetEpsilon(dmin / 2)
		}
		log.Println("doEpsilonCheck: ", c.Epsilon, c.pClusters.NumClusters()+c.oClusters.NumClusters())
	}*/

	// if c.pClusters.NumClusters()+c.oClusters.NumClusters() == minNumClusters {
	// 	c.adjustEpsilon()
	// }
	// if c.pClusters.NumClusters()+c.oClusters.NumClusters() >= minNumClusters && inputCount%100 == 0 && hmt <= 30 {
	// 	log.Println("adjust Epsilon")
	// 	hmt++
	// 	c.adjustEpsilon()
	// }
	//based on a function
	//every 10k points
}

func (c *Clusterer) doMaxOutlierWeightCheck() {

	pClusterWeight := c.pClusters.TotalWeight()
	oClusterWeight := c.oClusters.TotalWeight()

	currOutlierPercentage := (oClusterWeight / (pClusterWeight + oClusterWeight)) * 100
	log.Println("currOutlierPercentage: ", currOutlierPercentage)
	if currOutlierPercentage > percentageNoise {
		c.MaxOutlierWeight--
	} else if currOutlierPercentage < percentageNoise {
		c.MaxOutlierWeight++
	}
	log.Println("modified MaxOutlierWeight: ", c.MaxOutlierWeight)
}
