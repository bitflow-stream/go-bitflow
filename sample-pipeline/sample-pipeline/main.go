package main

import (
	"flag"

	"github.com/antongulenko/data2go/sample-pipeline"
	"github.com/antongulenko/golib"
)

func main() {
	var p pipeline.SamplePipeline
	p.ParseFlags()
	flag.Parse()
	defer golib.ProfileCpu()()
	p.Init()
	p.StartAndWait()
}
