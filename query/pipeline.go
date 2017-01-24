package query

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type AnalysisFunc func(pipeline *SamplePipeline)
type ParameterizedAnalysisFunc func(pipeline *SamplePipeline, params string)
type HandlerFunc func(string) bitflow.ReadSampleHandler

type PipelineBuilder struct {
	Endpoints bitflow.EndpointFactory

	analysis_registry map[string]registeredAnalysis
	handler_registry  map[string]HandlerFunc
}

func NewPipelineBuilder() PipelineBuilder {
	return PipelineBuilder{
		analysis_registry: make(map[string]registeredAnalysis),
		handler_registry:  make(map[string]HandlerFunc),
	}
}

type registeredAnalysis struct {
	Name   string
	Func   ParameterizedAnalysisFunc
	Params string
}

func (b PipelineBuilder) RegisterSampleHandler(name string, sampleHandler HandlerFunc) {
	if _, ok := handler_registry[name]; ok {
		panic("Sample handler already registered: " + name)
	}
	b.handler_registry[name] = sampleHandler
}

func (b PipelineBuilder) RegisterAnalysis(name string, setupPipeline AnalysisFunc) {
	b.RegisterAnalysisParams(name, func(pipeline *SamplePipeline, _ string) {
		setupPipeline(pipeline)
	}, "")
}

func (b PipelineBuilder) RegisterAnalysisParams(name string, setupPipeline ParameterizedAnalysisFunc, paramDescription string) {
	if _, ok := b.analysis_registry[name]; ok {
		panic("Analysis already registered: " + name)
	}
	b.analysis_registry[name] = registeredAnalysis{name, setupPipeline, paramDescription}
}

func (builder PipelineBuilder) getAnalysis(name_tok Token) (AnalysisFunc, HandlerFunc, error) {
	name := name_tok.Content()
	if analysis, ok := analysis_registry[name]; ok {
		return analysis, nil, nil
	} else {
		handler, ok := handler_registry[name]
		if !ok {
			return nil, nil, ParserError{
				Pos:     name_tok,
				Message: fmt.Sprintf("Pipeline step '%v' is unknown", name),
			}
		}
		return nil, handler, nil
	}
}

func (builder PipelineBuilder) MakePipelines(pipes Pipelines) (res []bitflow.SamplePipeline, err error) {
	for _, pipe := range pipes {
		for _, step := range pipe {
			var currentPipeline bitflow.SamplePipeline
			if err := step.ExtendPipeline(&currentPipeline, builder); err != nil {
				return nil, err
			}
			res = append(res, currentPipeline)
		}
	}
	return
}

func (p Pipelines) ExtendPipeline(pipe *bitflow.SamplePipelin, builder PipelineBuildere) error {

}

func (p Fork) ExtendPipeline(pipe *bitflow.SamplePipeline, builder PipelineBuilder) error {
	return ParserError{
		Pos:     p.Pos(),
		Message: "Forks are not implemented",
	}
}

func (p Step) ExtendPipeline(pipe *bitflow.SamplePipeline, builder PipelineBuilder) error {

}

func (p Input) ExtendPipeline(pipe *bitflow.SamplePipeline, builder PipelineBuilder) error {
	if pipe.Source != nil {
		return ParserError{
			Pos:     p.Pos(),
			Message: "Multiple inputs defined for the pipeline",
		}
	}
	for _, in := range p {
		builder.Endpoints.FlagInputs = append(builder.Endpoints.FlagInputs, in.Content())
	}
	source, err := builder.Endpoints.CreateInput()
	if err != nil {
		return err
	}
	pipe.Source = source
	return nil
}

func (p Output) ExtendPipeline(pipe *bitflow.SamplePipeline, builder PipelineBuilder) error {
	if pipe.Sink != nil {
		return ParserError{
			Pos:     p.Pos(),
			Message: "Multiple outputs defined for the pipeline",
		}
	}

	builder.Endpoints.FlagOutputs = golib.StringSlice{p.Content()}
	sink, err := builder.Endpoints.CreateOutput()
	if err != nil {
		return err
	}
	pipe.Sink = sink
	return nil
}
