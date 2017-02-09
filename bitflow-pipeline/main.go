package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
)

type Pipeline struct {
	*pipeline.SamplePipeline
}

var builder query.PipelineBuilder

func RegisterAnalysis(name string, setupPipeline func(pipeline *Pipeline), description string) {
	builder.RegisterAnalysis(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		if len(params) > 0 {
			return errors.New("No parmeters expected")
		}
		setupPipeline(&Pipeline{pipeline})
		return nil
	}, description)
}

func RegisterAnalysisErr(name string, setupPipeline func(pipeline *Pipeline) error, description string) {
	builder.RegisterAnalysis(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		if len(params) > 0 {
			return errors.New("No parmeters expected")
		}
		return setupPipeline(&Pipeline{pipeline})
	}, description)
}

func RegisterAnalysisParams(name string, setupPipeline func(pipeline *Pipeline, params map[string]string), description string, requiredParams []string, optionalParams ...string) {
	RegisterAnalysisParamsErr(name, func(pipeline *Pipeline, params map[string]string) error {
		setupPipeline(pipeline, params)
		return nil
	}, description, requiredParams, optionalParams...)
}

func RegisterAnalysisParamsErr(name string, setupPipeline func(pipeline *Pipeline, params map[string]string) error, description string, requiredParams []string, optionalParams ...string) {
	if len(requiredParams) > 0 {
		description += fmt.Sprintf(". Required parameters: %v", requiredParams)
	} else if requiredParams == nil {
		description += ". Variable parameters"
	}
	if len(optionalParams) > 0 {
		description += fmt.Sprintf(". Optional parameters: %v", optionalParams)
	}

	builder.RegisterAnalysis(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		if requiredParams != nil {
			if err := query.CheckParameters(params, optionalParams, requiredParams); err != nil {
				return err
			}
		}
		return setupPipeline(&Pipeline{pipeline}, params)
	}, description)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <flags> <bitflow script>\nAll flags must be defined before the first non-flag parameter.\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	os.Exit(do_main())
}

func do_main() int {
	printAnalyses := flag.Bool("print-analyses", false, "Print a list of available analyses and exit.")
	printPipeline := flag.Bool("print-pipeline", false, "Print the parsed pipeline and exit. Can be used to verify the input script.")

	bitflow.RegisterGolibFlags()
	builder.Endpoints.RegisterFlags()
	flag.Parse()
	golib.ConfigureLogging()
	if *printAnalyses {
		fmt.Printf("Available analysis steps:\n%v\n", builder.PrintAllAnalyses())
		return 0
	}

	pipeline, err := make_pipeline()
	if err != nil {
		log.Errorln(err)
		log.Fatalln("Use -print-analyses to print all available analysis steps.")
	}
	defer golib.ProfileCpu()()
	print_pipeline(pipeline)
	if *printPipeline {
		return 0
	}
	return pipeline.StartAndWait()
}

func make_pipeline() (*pipeline.SamplePipeline, error) {
	script := strings.Join(flag.Args(), " ")
	if strings.TrimSpace(script) == "" {
		return nil, errors.New("Please provide a bitflow pipeline script")
	}
	parser := query.NewParser(bytes.NewReader([]byte(script)))
	pipe, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return builder.MakePipeline(pipe)
}

func print_pipeline(pipe *pipeline.SamplePipeline) {
	printer := pipeline.IndentPrinter{
		OuterIndent:  "│ ",
		InnerIndent:  "├─",
		CornerIndent: "└─",
		FillerIndent: "  ",
	}
	lines := printer.PrintLines(pipe)
	for _, str := range lines {
		log.Println(str)
	}
}
