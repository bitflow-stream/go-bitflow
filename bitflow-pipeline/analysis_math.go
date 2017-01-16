package main

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/dbscan"
	"github.com/antongulenko/go-bitflow-pipeline/regression"
	"github.com/antongulenko/golib"
)

func init() {
	RegisterAnalysis("dbscan", dbscan_rtree)
	RegisterAnalysis("dbscan_parallel", dbscan_parallel)

	RegisterAnalysis("regression", linear_regression)
	RegisterAnalysis("regression_brute", linear_regression_bruteforce)

	RegisterAnalysisParams("pca", pca_analysis, "contained variance 0..1")
	RegisterAnalysisParams("pca_store", pca_analysis_store, "filename to store the PCA model")
	RegisterAnalysisParams("pca_load", pca_analysis_load, "<filename containing PCA model>,<contained variance 0..1>")
	RegisterAnalysisParams("pca_load_stream", pca_analysis_load_stream, "<filename containing PCA model>,<contained variance 0..1>")

	RegisterAnalysisParams("sphere", add_sphere, "<number of points>,<fixed radius>,<random-seed>")
	RegisterAnalysisParams("sphere_metric", add_sphere_metric, "<number of points>,<radius-metric>,<random-seed>")
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
	pipe.Batch(ComputeAndProjectPCA(parse_pca_variance(params)))
}

func pca_analysis_store(pipe *SamplePipeline, params string) {
	if params == "" {
		log.Fatalln("pca_store needs non-empty parameter: filename for storing PCA model")
	}
	pipe.Batch(StorePCAModel(params))
}

func pca_analysis_load(pipe *SamplePipeline, params string) {
	step, err := LoadBatchPCAModel(parse_pca_variance_filename(params))
	golib.Checkerr(err)
	pipe.Batch(step)
}

func pca_analysis_load_stream(pipe *SamplePipeline, params string) {
	step, err := LoadStreamingPCAModel(parse_pca_variance_filename(params))
	golib.Checkerr(err)
	pipe.Add(step)
}

func parse_pca_variance(str string) float64 {
	variance := DefaultContainedVariance
	if str != "" {
		var err error
		if variance, err = strconv.ParseFloat(str, 64); err != nil {
			log.Fatalln("Failed to parse PCA variance parameter:", err)
		}
	} else {
		log.Warnln("No parameter PCA, default contained variance:", variance)
	}
	return variance
}

func parse_pca_variance_filename(params string) (string, float64) {
	parts := strings.Split(params, ",")
	if len(parts) != 2 {
		log.Fatalln("Illegal PCA parameter. Need: <filename containing PCA model>,<contained variance 0..1>. Got:", params)
	}
	return parts[0], parse_pca_variance(parts[1])
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
	do_add_sphere(pipe, params, true)
}

func add_sphere_metric(pipe *SamplePipeline, params string) {
	do_add_sphere(pipe, params, false)
}

func do_add_sphere(pipe *SamplePipeline, params string, fixed_radius bool) {
	parts := strings.Split(params, ",")
	if len(parts) != 3 {
		log.Fatalln("sphere needs 3 parameters: <number of points>,<radius>,<random-seed>")
	}
	points, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Fatalln("Error parsing -e sphere number-of-points parameter:", err)
	}
	seed, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		log.Fatalln("Error parsing -e random-seed parameter:", err)
	}

	sphere := &SpherePoints{
		RandomSeed: seed,
		NumPoints:  points,
	}

	if fixed_radius {
		radius, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			log.Fatalln("Error parsing -e sphere radius parameter:", err)
		}
		sphere.RadiusMetric = -1
		sphere.Radius = radius
	} else {
		radius, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Fatalln("Error parsing -e sphere radius parameter:", err)
		}
		sphere.RadiusMetric = radius
	}
	pipe.Add(sphere)
}

func filter_convex_hull(pipe *SamplePipeline) {
	pipe.Batch(BatchConvexHull(false))
}

func sort_convex_hull(pipe *SamplePipeline) {
	pipe.Batch(BatchConvexHull(true))
}
