package main

import (
	"errors"
	"fmt"

	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

type Pipeline struct {
	*pipeline.SamplePipeline
}

var builder = query.NewPipelineBuilder()

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

func RegisterFork(name string, createFork query.ForkFunc, description string, requiredParams []string, optionalParams ...string) {
	if len(requiredParams) > 0 {
		description += fmt.Sprintf(". Required parameters: %v", requiredParams)
	} else if requiredParams == nil {
		description += ". Variable parameters"
	}
	if len(optionalParams) > 0 {
		description += fmt.Sprintf(". Optional parameters: %v", optionalParams)
	}

	builder.RegisterFork(name, func(params map[string]string) (fmt.Stringer, error) {
		if requiredParams != nil {
			if err := query.CheckParameters(params, optionalParams, requiredParams); err != nil {
				return nil, err
			}
		}
		return createFork(params)
	}, description)
}
