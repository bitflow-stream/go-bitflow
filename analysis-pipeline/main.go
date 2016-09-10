package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

type AnalysisFunc func(pipeline *SamplePipeline, params string)

// Can be filled from init() functions using RegisterAnalysis()
var analysis_registry = map[string]AnalysisFunc{
	"": nil,
}
var handler_registry = map[string]sample.ReadSampleHandler{
	"": nil,
}

func RegisterSampleHandler(name string, sampleHandler sample.ReadSampleHandler) {
	if _, ok := handler_registry[name]; ok {
		log.Fatalln("Sample handler already registered:", name)
	}
	handler_registry[name] = sampleHandler
}

func RegisterAnalysis(name string, setupPipeline AnalysisFunc) {
	if _, ok := analysis_registry[name]; ok {
		log.Fatalln("Analysis already registered:", name)
	}
	analysis_registry[name] = setupPipeline
}

func main() {
	os.Exit(do_main())
}

func do_main() int {
	var analysisNames golib.StringSlice
	var readSampleHandler string
	flag.Var(&analysisNames, "e", fmt.Sprintf("Select one or more of the following analysis pipelines to execute: %v", allAnalyses()))
	flag.StringVar(&readSampleHandler, "h", "", fmt.Sprintf("Select an optional sample handler for handling incoming samples: %v", allHandlers()))

	var p SamplePipeline
	p.ParseFlags()
	flag.Parse()
	analyses := resolvePipeline(analysisNames)
	handler, ok := handler_registry[readSampleHandler]
	if !ok {
		log.Fatalf("Sample handler '%v' not registered. Available: %v", readSampleHandler, allHandlers())
	}
	defer golib.ProfileCpu()()
	p.Init()

	p.ReadSampleHandler = handler
	for _, analysis := range analyses {
		if setup := analysis.setup; setup != nil {
			setup(&p, analysis.params)
		}
	}
	printPipeline(p.Processors)
	return p.StartAndWait()
}

type parameterizedAnalysis struct {
	setup  AnalysisFunc
	params string
}

func resolvePipeline(analysisNames golib.StringSlice) []parameterizedAnalysis {
	if len(analysisNames) == 0 {
		analysisNames = append(analysisNames, "") // The default
	}
	result := make([]parameterizedAnalysis, len(analysisNames))
	for i, name := range analysisNames {
		params := ""
		if index := strings.IndexRune(name, ','); index >= 0 {
			full := name
			name = full[:index]
			params = full[index+1:]
		}
		analysisFunc, ok := analysis_registry[name]
		if !ok {
			log.Fatalf("Analysis pipeline '%v' not registered. Available: %v", name, allAnalyses())
		}
		result[i] = parameterizedAnalysis{analysisFunc, params}
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
			log.Println(indent + proc.String())
		}
	}
}

func allAnalyses() []string {
	all := make([]string, 0, len(analysis_registry))
	for name := range analysis_registry {
		all = append(all, name)
	}
	sort.Strings(all)
	return all
}

func allHandlers() []string {
	all := make([]string, 0, len(handler_registry))
	for name := range handler_registry {
		if name != "" {
			all = append(all, name)
		}
	}
	sort.Strings(all)
	return all
}

type SamplePipeline struct {
	sample.CmdSamplePipeline
	batch *analysis.BatchProcessor
}

func (pipeline *SamplePipeline) Add(step sample.SampleProcessor) *SamplePipeline {
	pipeline.batch = nil
	pipeline.CmdSamplePipeline.Add(step)
	return pipeline
}

func (pipeline *SamplePipeline) Batch(proc analysis.BatchProcessingStep) *SamplePipeline {
	if pipeline.batch == nil {
		pipeline.batch = new(analysis.BatchProcessor)
		pipeline.CmdSamplePipeline.Add(pipeline.batch)
	}
	pipeline.batch.Add(proc)
	return pipeline
}
