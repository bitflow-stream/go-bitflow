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
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
)

type SamplePipeline struct {
	*query.SamplePipeline
}

func RegisterAnalysis(name string, setupPipeline func(pipeline *SamplePipeline), description string) {
	builder.RegisterAnalysis(name, func(pipeline *query.SamplePipeline, params map[string]string) error {
		if len(params) > 0 {
			return errors.New("No parmeters expected")
		}
		setupPipeline(&SamplePipeline{pipeline})
		return nil
	}, description)
}

func RegisterAnalysisErr(name string, setupPipeline func(pipeline *SamplePipeline) error, description string) {
	builder.RegisterAnalysis(name, func(pipeline *query.SamplePipeline, params map[string]string) error {
		if len(params) > 0 {
			return errors.New("No parmeters expected")
		}
		return setupPipeline(&SamplePipeline{pipeline})
	}, description)
}

func RegisterAnalysisParams(name string, setupPipeline func(pipeline *SamplePipeline, params map[string]string), description string, requiredParams []string, optionalParams ...string) {
	RegisterAnalysisParamsErr(name, func(pipeline *SamplePipeline, params map[string]string) error {
		setupPipeline(pipeline, params)
		return nil
	}, description, requiredParams, optionalParams...)
}

func RegisterAnalysisParamsErr(name string, setupPipeline func(pipeline *SamplePipeline, params map[string]string) error, description string, requiredParams []string, optionalParams ...string) {
	if len(requiredParams) > 0 {
		description += fmt.Sprintf(". Required parameters: %v", requiredParams)
	} else if requiredParams == nil {
		description += ". Variable parameters"
	}
	if len(optionalParams) > 0 {
		description += fmt.Sprintf(". Required parameters: %v", optionalParams)
	}

	builder.RegisterAnalysis(name, func(pipeline *query.SamplePipeline, params map[string]string) error {
		if requiredParams != nil {
			if err := query.CheckParameters(params, optionalParams, requiredParams); err != nil {
				return err
			}
		}
		return setupPipeline(&SamplePipeline{pipeline}, params)
	}, description)
}

var builder query.PipelineBuilder

func main() {
	os.Exit(do_main())
}

func do_main() int {
	printAnalyses := flag.Bool("print-analyses", false, "Print a list of available analyses and exit.")

	bitflow.RegisterGolibFlags()
	builder.Endpoints.RegisterConfigFlags()
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
	for _, str := range pipeline.Print() {
		log.Println(str)
	}
	return pipeline.StartAndWait()
}

func make_pipeline() (query.SamplePipeline, error) {
	script := strings.Join(flag.Args(), " ")
	parser := query.NewParser(bytes.NewReader([]byte(script)))
	pipes, err := parser.Parse()
	if err != nil {
		return query.SamplePipeline{}, err
	}
	return builder.MakePipeline(pipes)
}
