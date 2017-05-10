package query

import (
	"errors"
	"fmt"

	"github.com/antongulenko/go-bitflow-pipeline"
)

type AnalysisFunc func(pipeline *pipeline.SamplePipeline, params map[string]string) error
type ForkFunc func(params map[string]string) (fmt.Stringer, error) // Can return a fork.ForkDistributor or a fork.RemapDistributor

type registeredAnalysis struct {
	Name        string
	Func        AnalysisFunc
	Description string
}

type registeredFork struct {
	Name        string
	Func        ForkFunc
	Description string
}

func (b *PipelineBuilder) RegisterAnalysisParamsErr(name string, setupPipeline AnalysisFunc, description string, requiredParams []string, optionalParams ...string) {
	if _, ok := b.analysis_registry[name]; ok {
		panic("Analysis already registered: " + name)
	}

	if len(requiredParams) > 0 {
		description += fmt.Sprintf(". Required parameters: %v", requiredParams)
	} else if requiredParams == nil {
		description += ". Variable parameters"
	}
	if len(optionalParams) > 0 {
		description += fmt.Sprintf(". Optional parameters: %v", optionalParams)
	}

	doSetup := func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		if requiredParams != nil {
			if err := checkParameters(params, optionalParams, requiredParams); err != nil {
				return err
			}
		}
		return setupPipeline(pipeline, params)
	}
	b.analysis_registry[name] = registeredAnalysis{name, doSetup, description}
}

func (b *PipelineBuilder) RegisterAnalysisParams(name string, setupPipeline func(pipeline *pipeline.SamplePipeline, params map[string]string), description string, requiredParams []string, optionalParams ...string) {
	b.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		setupPipeline(pipeline, params)
		return nil
	}, description, requiredParams, optionalParams...)
}

func (b *PipelineBuilder) RegisterAnalysis(name string, setupPipeline func(pipeline *pipeline.SamplePipeline), description string) {
	b.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		if len(params) > 0 {
			return errors.New("No parmeters expected")
		}
		setupPipeline(pipeline)
		return nil
	}, description, nil)
}

func (b *PipelineBuilder) RegisterAnalysisErr(name string, setupPipeline func(pipeline *pipeline.SamplePipeline) error, description string) {
	b.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		if len(params) > 0 {
			return errors.New("No parmeters expected")
		}
		return setupPipeline(pipeline)
	}, description, nil)
}

func (b *PipelineBuilder) RegisterFork(name string, createFork ForkFunc, description string, requiredParams []string, optionalParams ...string) {
	if _, ok := b.fork_registry[name]; ok {
		panic("Fork already registered: " + name)
	}

	if len(requiredParams) > 0 {
		description += fmt.Sprintf(". Required parameters: %v", requiredParams)
	} else if requiredParams == nil {
		description += ". Variable parameters"
	}
	if len(optionalParams) > 0 {
		description += fmt.Sprintf(". Optional parameters: %v", optionalParams)
	}

	doSetup := func(params map[string]string) (fmt.Stringer, error) {
		if requiredParams != nil {
			if err := checkParameters(params, optionalParams, requiredParams); err != nil {
				return nil, err
			}
		}
		return createFork(params)
	}
	b.fork_registry[name] = registeredFork{name, doSetup, description}
}

func checkParameters(params map[string]string, optional []string, required []string) error {
	checked := map[string]bool{}
	for _, opt := range optional {
		checked[opt] = true
	}
	for _, req := range required {
		if _, ok := params[req]; !ok {
			return fmt.Errorf("Missing required parameter '%v'", req)
		}
		checked[req] = true
	}
	for key := range params {
		if _, ok := checked[key]; !ok {
			return fmt.Errorf("Unexpected parameter '%v'", key)
		}
	}
	return nil
}
