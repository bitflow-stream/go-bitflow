package main

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/dbscan"
	"github.com/antongulenko/go-bitflow-pipeline/regression"
)

func init() {
	RegisterAnalysis("dbscan", dbscan_rtree)
	RegisterAnalysis("dbscan_parallel", dbscan_parallel)
	RegisterAnalysisParams("pca", pca_analysis, "contained variance 0..1")
	RegisterAnalysis("regression", linear_regression)
	RegisterAnalysis("regression_brute", linear_regression_bruteforce)

	RegisterAnalysisParams("sphere", add_sphere, "<number of points>,<radius>,<random-seed>")
	RegisterAnalysis("convex_hull", filter_convex_hull)
	RegisterAnalysis("convex_hull_sort", sort_convex_hull)
}

func linear_regression(p *SamplePipeline) {
	p.Batch(&regression.LinearRegressionBatchProcessor{})
}

func linear_regression_bruteforce(p *SamplePipeline) {
	p.Batch(&regression.LinearRegressionBruteForce{})
}

func pca_analysis(pipe *SamplePipeline, params string) {
	variance := 0.99
	if params != "" {
		var err error
		if variance, err = strconv.ParseFloat(params, 64); err != nil {
			log.Fatalln("Failed to parse parameter for -e pca:", err)
		}
	} else {
		log.Warnln("No parameter for -e pca, default contained variance:", variance)
	}
	pipe.Batch(&PCABatchProcessing{ContainedVariance: variance})
}

func dbscan_rtree(pipe *SamplePipeline) {
	pipe.Batch(&dbscan.DbscanBatchClusterer{
		Dbscan:          dbscan.Dbscan{Eps: 0.1, MinPts: 5},
		TreeMinChildren: 25,
		TreeMaxChildren: 50,
		TreePointWidth:  0.0001,
	})
}

func dbscan_parallel(pipe *SamplePipeline) {
	pipe.Batch(&dbscan.ParallelDbscanBatchClusterer{Eps: 0.3, MinPts: 5})
}

func add_sphere(pipe *SamplePipeline, params string) {
	parts := strings.Split(params, ",")
	if len(parts) != 3 {
		log.Fatalln("-e sphere needs 3 parameters: <number of points>,<radius>,<random-seed>")
	}
	points, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Fatalln("Error parsing -e sphere number-of-points parameter:", err)
	}
	radius, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		log.Fatalln("Error parsing -e sphere radius parameter:", err)
	}
	seed, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		log.Fatalln("Error parsing -e random-seed parameter:", err)
	}
	pipe.Add(&SpherePoints{
		RandomSeed: seed,
		Radius:     radius,
		NumPoints:  points,
	})
}

func filter_convex_hull(pipe *SamplePipeline) {
	pipe.Batch(&BatchConvexHull{Sort: false})
}

func sort_convex_hull(pipe *SamplePipeline) {
	pipe.Batch(&BatchConvexHull{Sort: true})
}
