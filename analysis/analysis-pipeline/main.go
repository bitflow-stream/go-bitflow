package main

import (
	"flag"
	"os"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

var _ = new(AbstractProcessor) // validate data2go/analysis import

func do_main() int {
	var p sample.CmdSamplePipeline
	p.ParseFlags()
	flag.Parse()
	defer golib.ProfileCpu()()
	p.Init()
	registerProcessingSteps(&p.SamplePipeline)
	return p.StartAndWait()
}

func main() {
	os.Exit(do_main())
}

func registerProcessingSteps(p *sample.SamplePipeline) {
	// p.Add(new(BatchProcessor))
	// .Add(new(TimestampSort))
	// .Add(new(MinMaxScaling))
	// .Add(&PCABatchProcessing{ContainedVariance: 0.99})

	// p.Add(new(SamplePrinter))
	// p.Add(new(AbstractProcessor))
	// p.Add(&DecouplingProcessor{ChannelBuffer: 150000})
	// p.Add(&Plotter{OutputFile: "/home/anton/test.jpg", ColorTag: "cls"})
}
