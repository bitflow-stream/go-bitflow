package main

import (
	"errors"
	"strconv"

	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/dbscan"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/go-bitflow-pipeline/regression"
)

func RegisterMathAnalyses(b *query.PipelineBuilder) {
	b.RegisterAnalysis("dbscan", dbscan_rtree, "Perform a dbscan clustering on a batch of samples")
	b.RegisterAnalysis("dbscan_parallel", dbscan_parallel, "Perform a parallelized dbscan clustering on a batch of samples")

	b.RegisterAnalysis("regression", linear_regression, "Perform a linear regression analysis on a batch of samples")
	b.RegisterAnalysis("regression_brute", linear_regression_bruteforce, "In a batch of samples, perform a linear regression analysis for every possible combination of metrics")

	b.RegisterAnalysisParamsErr("pca", pca_analysis, "Create a PCA model of a batch of samples and project all samples into a number of principal components with a total contained variance given by the parameter", []string{"var"})
	b.RegisterAnalysisParams("pca_store", pca_analysis_store, "Create a PCA model of a batch of samples and store it to the given file", []string{"file"})
	b.RegisterAnalysisParamsErr("pca_load", pca_analysis_load, "Load a PCA model from the given file and project all samples into a number of principal components with a total contained variance given by the parameter", []string{"var", "file"})
	b.RegisterAnalysisParamsErr("pca_load_stream", pca_analysis_load_stream, "Like pca_load, but process every sample individually, instead of batching them up", []string{"var", "file"})

	b.RegisterAnalysisParamsErr("sphere", add_sphere, "Treat every sample as the center of a multi-dimensional sphere, and output a number of random points on the hull of the resulting sphere. The radius can either be fixed or given as one of the metrics", []string{"points"}, "seed", "radius", "radius_metric")
	b.RegisterAnalysis("convex_hull", filter_convex_hull, "Filter out the convex hull for a two-dimensional batch of samples")
	b.RegisterAnalysis("convex_hull_sort", sort_convex_hull, "Sort a two-dimensional batch of samples in order around their center")
}

func linear_regression(p *SamplePipeline) {
	p.Batch(&regression.LinearRegressionBatchProcessor{})
}

func linear_regression_bruteforce(p *SamplePipeline) {
	p.Batch(&regression.LinearRegressionBruteForce{})
}

func pca_analysis(p *SamplePipeline, params map[string]string) error {
	variance, err := parse_pca_variance(params)
	if err == nil {
		p.Batch(ComputeAndProjectPCA(variance))
	}
	return err
}

func pca_analysis_store(p *SamplePipeline, params map[string]string) {
	p.Batch(StorePCAModel(params["file"]))
}

func pca_analysis_load(p *SamplePipeline, params map[string]string) error {
	variance, err := parse_pca_variance(params)
	if err == nil {
		var step BatchProcessingStep
		step, err = LoadBatchPCAModel(params["file"], variance)
		if err == nil {
			p.Batch(step)
		}
	}
	return err
}

func pca_analysis_load_stream(p *SamplePipeline, params map[string]string) error {
	variance, err := parse_pca_variance(params)
	if err == nil {
		var step bitflow.SampleProcessor
		step, err = LoadStreamingPCAModel(params["file"], variance)
		if err == nil {
			p.Add(step)
		}
	}
	return err
}

func parse_pca_variance(params map[string]string) (float64, error) {
	variance, err := strconv.ParseFloat(params["var"], 64)
	if err != nil {
		err = parameterError("var", err)
	}
	return variance, err
}

func dbscan_rtree(p *SamplePipeline) {
	p.Batch(&dbscan.DbscanBatchClusterer{
		Dbscan:          dbscan.Dbscan{Eps: 0.1, MinPts: 5},
		TreeMinChildren: 25,
		TreeMaxChildren: 50,
		TreePointWidth:  0.0001,
	})
}

func dbscan_parallel(p *SamplePipeline) {
	p.Batch(&dbscan.ParallelDbscanBatchClusterer{Eps: 0.3, MinPts: 5})
}

func add_sphere(p *SamplePipeline, params map[string]string) error {
	var err error
	points, err := strconv.Atoi(params["points"])
	if err != nil {
		return parameterError("points", err)
	}
	seed := int64(1)
	if seedStr, ok := params["seed"]; ok {
		seed, err = strconv.ParseInt(seedStr, 10, 64)
		if err != nil {
			return parameterError("seed", err)
		}
	}
	radiusStr, hasRadius := params["radius"]
	radiusMetricStr, hasRadiusMetric := params["radius_metric"]
	if hasRadius == hasRadiusMetric {
		return errors.New("Need either 'radius' or 'radius_metric' parameter")
	}

	sphere := &SpherePoints{
		RandomSeed: seed,
		NumPoints:  points,
	}
	if hasRadius {
		sphere.RadiusMetric = -1
		sphere.Radius, err = strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			return parameterError("radius", err)
		}
	} else {
		sphere.RadiusMetric, err = strconv.Atoi(radiusMetricStr)
		if err != nil {
			return parameterError("radius_metric", err)
		}
	}
	p.Add(sphere)
	return nil
}

func filter_convex_hull(p *SamplePipeline) {
	p.Batch(BatchConvexHull(false))
}

func sort_convex_hull(p *SamplePipeline) {
	p.Batch(BatchConvexHull(true))
}
