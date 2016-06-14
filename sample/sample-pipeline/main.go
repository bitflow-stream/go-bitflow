package main

import (
	"flag"
	"os"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

var handlePipeline func(p *sample.CmdSamplePipeline)

// Example pipeline with just the basic input and output parts (no data processing)
// This can read data from a source and output it to one or more sinks.
// Setting handlePipeline can be used e.g. to add SampleProcessors to the pipeline.
func do_main() int {
	var p sample.CmdSamplePipeline
	p.ParseFlags()
	flag.Parse()
	defer golib.ProfileCpu()()
	p.Init()
	if handlePipeline != nil {
		handlePipeline(&p)
	}
	return p.StartAndWait()
}

func main() {
	os.Exit(do_main())
}
