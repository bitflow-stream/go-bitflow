package main

import (
	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/analysis/regression"
	"github.com/antongulenko/data2go/sample"
)

func init() {
	RegisterAnalysis("regression", nil, add_linear_regression)
	RegisterAnalysis("regression_brute", nil, add_linear_regression_bruteforce)
}

func add_linear_regression(p *sample.CmdSamplePipeline) {
	p.Add(new(BatchProcessor).Add(&regression.LinearRegressionBatchProcessor{Fields: []string{"cpu", "mem/percent", "net-io/bytes"}}))
}

func add_linear_regression_bruteforce(p *sample.CmdSamplePipeline) {
	p.Add(new(BatchProcessor).Add(new(MinMaxScaling)).Add(&regression.LinearRegressionBruteForce{}))
}
