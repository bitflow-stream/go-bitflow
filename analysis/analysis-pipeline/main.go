package main

import (
	"flag"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

var _ = new(AbstractProcessor) // validate data2go/analysis import

func main() {
	var p sample.CmdSamplePipeline
	p.ParseFlags()
	flag.Parse()
	defer golib.ProfileCpu()()
	p.Init()
	registerProcessingSteps(&p.SamplePipeline)
	p.StartAndWait()
}

func registerProcessingSteps(p *sample.SamplePipeline) {
	// p.Add(new(BatchProcessor))
	// .Add(new(TimestampSort))
	// .Add(new(MinMaxScaling))
	// .Add(&PCABatchProcessing{ContainedVariance: 0.99})

	//	p.Add(new(SamplePrinter))
	//	p.Add(new(AbstractProcessor))
	//	p.Add(&DecouplingProcessor{ChannelBuffer: 200})
	//	p.Add(&Plotter{OutputFile: "/home/anton/test.jpg", ColorTag: "cls"})
}
