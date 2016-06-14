package main

import (
	"flag"
	"os"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

var handlePipeline func(p *sample.CmdSamplePipeline)

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
