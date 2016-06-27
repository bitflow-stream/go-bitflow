package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

// Can be filled/configured from init() functions
var DefaultAnalysis = ""
var registry = make(map[string]analysis)

type analysis struct {
	SampleHandler sample.ReadSampleHandler
	SetupPipeline func(*sample.CmdSamplePipeline)
}

func RegisterAnalysis(name string, sampleHandler sample.ReadSampleHandler, setupPipeline func(*sample.CmdSamplePipeline)) {
	registry[name] = analysis{
		SampleHandler: sampleHandler,
		SetupPipeline: setupPipeline,
	}
}

func main() {
	os.Exit(do_main())
}

func do_main() int {
	analysisName := flag.String("e", DefaultAnalysis, fmt.Sprintf("Select one of the following analysis pipelines to execute: %v", allAnalyses()))

	var p sample.CmdSamplePipeline
	p.ParseFlags()
	flag.Parse()
	analysis, ok := registry[*analysisName]
	if !ok {
		log.Fatalf("Analysis pipeline '%v' not registered. Available: %v\n", *analysisName, allAnalyses())
	}
	p.ReadSampleHandler = analysis.SampleHandler
	defer golib.ProfileCpu()()
	p.Init()
	if setup := analysis.SetupPipeline; setup != nil {
		setup(&p)
	}
	return p.StartAndWait()
}

func allAnalyses() []string {
	all := make([]string, 0, len(registry))
	for name := range registry {
		all = append(all, name)
	}
	sort.Strings(all)
	return all
}
