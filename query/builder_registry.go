package query

import (
	"fmt"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
)

type Subpipeline struct {
	keys []string

	builder *PipelineBuilder
	pipe    Pipeline
}

func (s *Subpipeline) Keys() []string {
	return s.keys
}

func (s *Subpipeline) Build() (*pipeline.SamplePipeline, error) {
	return s.builder.makePipelineTail(s.pipe)
}

func (b *PipelineBuilder) RegisterAnalysisParamsErr(name string, setupPipeline builder.AnalysisFunc, description string, options ...builder.Option) {
	if _, ok := b.analysis_registry[name]; ok {
		panic("Analysis already registered: " + name)
	}
	opts := builder.GetOpts(options)
	params := registeredParameters{opts.RequiredParams, opts.OptionalParams}
	b.analysis_registry[name] = registeredAnalysis{name, setupPipeline, params.makeDescription(description), params}
}

func (builder *PipelineBuilder) RegisterAnalysisParams(name string, setupPipeline func(pipeline *pipeline.SamplePipeline, params map[string]string), description string, options ...builder.Option) {
	builder.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		setupPipeline(pipeline, params)
		return nil
	}, description, options...)
}

func (builder *PipelineBuilder) RegisterAnalysis(name string, setupPipeline func(pipeline *pipeline.SamplePipeline), description string, options ...builder.Option) {
	builder.RegisterAnalysisParams(name, func(pipeline *pipeline.SamplePipeline, _ map[string]string) {
		setupPipeline(pipeline)
	}, description, options...)
}

func (builder *PipelineBuilder) RegisterAnalysisErr(name string, setupPipeline func(pipeline *pipeline.SamplePipeline) error, description string, options ...builder.Option) {
	builder.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, _ map[string]string) error {
		return setupPipeline(pipeline)
	}, description, options...)
}

func (b *PipelineBuilder) RegisterFork(name string, createFork builder.ForkFunc, description string, options ...builder.Option) {
	if _, ok := b.fork_registry[name]; ok {
		panic("Fork already registered: " + name)
	}
	opts := builder.GetOpts(options)
	params := registeredParameters{opts.RequiredParams, opts.OptionalParams}
	b.fork_registry[name] = registeredFork{name, createFork, params.makeDescription(description), params}
}

type registeredAnalysis struct {
	Name        string
	Func        builder.AnalysisFunc
	Description string
	Params      registeredParameters
}

type registeredFork struct {
	Name        string
	Func        builder.ForkFunc
	Description string
	Params      registeredParameters
}

type registeredParameters struct {
	required []string
	optional []string
}

func (params registeredParameters) Verify(input map[string]string) error {
	checked := map[string]bool{}
	for _, opt := range params.optional {
		checked[opt] = true
	}
	for _, req := range params.required {
		if _, ok := input[req]; !ok {
			return fmt.Errorf("Missing required parameter '%v'", req)
		}
		checked[req] = true
	}
	if params.required != nil {
		for key := range input {
			if _, ok := checked[key]; !ok {
				return fmt.Errorf("Unexpected parameter '%v'", key)
			}
		}
	}
	return nil
}

func (params registeredParameters) makeDescription(description string) string {
	if len(params.required) > 0 {
		description += fmt.Sprintf(". Required parameters: %v", params.required)
	} else if params.required == nil {
		description += ". Variable parameters"
	}
	if len(params.optional) > 0 {
		description += fmt.Sprintf(". Optional parameters: %v", params.optional)
	}
	return description
}
