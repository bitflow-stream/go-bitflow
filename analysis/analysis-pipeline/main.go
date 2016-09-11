package main

import (
	"bytes"
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

type registeredAnalysis struct {
	Name   string
	Func   AnalysisFunc
	Params string
}

// Can be filled from init() functions using RegisterAnalysis()
var analysis_registry = map[string]registeredAnalysis{
	"": registeredAnalysis{"", nil, ""},
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
	RegisterAnalysisParams(name, setupPipeline, "")
}

func RegisterAnalysisParams(name string, setupPipeline AnalysisFunc, paramDescription string) {
	if _, ok := analysis_registry[name]; ok {
		log.Fatalln("Analysis already registered:", name)
	}
	analysis_registry[name] = registeredAnalysis{name, setupPipeline, paramDescription}
}

func main() {
	os.Exit(do_main())
}

func do_main() int {
	var analysisNames golib.StringSlice
	flag.Var(&analysisNames, "e", fmt.Sprintf("Select one or more analyses to execute. Use -analysis for listing all analyses."))
	readSampleHandler := flag.String("h", "", fmt.Sprintf("Select an optional sample handler for handling incoming samples: %v", allHandlers()))
	printAnalyses := flag.Bool("analyses", false, "Print a list of available analyses and exit.")

	var p SamplePipeline
	p.ParseFlags()
	flag.Parse()
	if *printAnalyses {
		fmt.Printf("Available analyses:%v\n", allAnalyses())
		return 0
	}
	analyses := resolvePipeline(analysisNames)
	handler, ok := handler_registry[*readSampleHandler]
	if !ok {
		log.Fatalf("Sample handler '%v' not registered. Available: %v", *readSampleHandler, allHandlers())
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
		analysis, ok := analysis_registry[name]
		if !ok {
			log.Fatalf("Analysis '%v' not registered. Available analyses:%v", name, allAnalyses())
		}
		result[i] = parameterizedAnalysis{analysis.Func, params}
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

func allAnalyses() string {
	all := make(SortedAnalyses, 0, len(analysis_registry))
	for _, analysis := range analysis_registry {
		all = append(all, analysis)
	}
	sort.Sort(all)
	var buf bytes.Buffer
	for i, analysis := range all {
		if analysis.Func == nil {
			continue
		}
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(" - ")
		buf.WriteString(analysis.Name)
		if analysis.Params != "" {
			buf.WriteString(" (Params: ")
			buf.WriteString(analysis.Params)
			buf.WriteString(")")
		}
	}
	return buf.String()
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

type SortedAnalyses []registeredAnalysis

func (slice SortedAnalyses) Len() int {
	return len(slice)
}

func (slice SortedAnalyses) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice SortedAnalyses) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
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
