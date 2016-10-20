package main

import (
	"strconv"

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
