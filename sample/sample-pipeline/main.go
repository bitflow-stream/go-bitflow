package main

import (
	"flag"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

// Example pipeline with just the basic input and output parts (no data processing)
// This can read data from a source and output it to one or more sinks.
func main() {
	var p sample.CmdSamplePipeline
	p.ParseFlags()
	flag.Parse()
	defer golib.ProfileCpu()()
	p.Init()
	p.StartAndWait()
}
