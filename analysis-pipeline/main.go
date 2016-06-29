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

// Can be filled from init() functions using RegisterAnalysis()
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
	var analysisNames golib.StringSlice
	flag.Var(&analysisNames, "e", fmt.Sprintf("Select one or more of the following analysis pipelines to execute: %v", allAnalyses()))

	var p sample.CmdSamplePipeline
	p.ParseFlags()
	flag.Parse()
	analyses := resolvePipeline(&p, analysisNames)
	if len(analyses) > 0 {
		// If multiple analyses are selected, use the SampleHandler of the first one
		p.ReadSampleHandler = analyses[0].SampleHandler
	}
	defer golib.ProfileCpu()()
	p.Init()

	for _, analysis := range analyses {
		if setup := analysis.SetupPipeline; setup != nil {
			setup(&p)
		}
	}
	printPipeline(p.Processors)
	return p.StartAndWait()
}

func resolvePipeline(p *sample.CmdSamplePipeline, analysisNames golib.StringSlice) []analysis {
	result := make([]analysis, len(analysisNames))
	for i, name := range analysisNames {
		analysis, ok := registry[name]
		if !ok {
			log.Fatalf("Analysis pipeline '%v' not registered. Available: %v\n", name, allAnalyses())
		}
		result[i] = analysis
	}
	return result
}

func printPipeline(p []sample.SampleProcessor) {
	if len(p) == 0 {
		log.Println("Empty analysis pipeline")
	} else if len(p) == 1 {
		log.Println("Analysis:", p[0])
	} else {
		log.Println("Analysis pipeline:")
		for i, proc := range p {
			indent := "├─"
			if i == len(p)-1 {
				indent = "└─"
			}
			log.Printf("%s %v\n", indent, proc)
		}
	}
}

func allAnalyses() []string {
	all := make([]string, 0, len(registry))
	for name := range registry {
		all = append(all, name)
	}
	sort.Strings(all)
	return all
}
