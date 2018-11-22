package script_go

import (
	"fmt"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/bitflow/fork"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

type PipelineBuilder struct {
	reg.ProcessorRegistry
}

type subpipeline struct {
	keys []string

	builder *PipelineBuilder
	pipe    Pipeline
}

func (s *subpipeline) Keys() []string {
	return s.keys
}

func (s *subpipeline) Build() (*bitflow.SamplePipeline, error) {
	return s.builder.makePipelineTail(s.pipe)
}

func (b PipelineBuilder) MakePipeline(pipe Pipeline) (*bitflow.SamplePipeline, error) {
	pipe, err := pipe.Transform(b)
	if err != nil {
		return nil, err
	}
	return b.makePipeline(pipe)
}

func (b PipelineBuilder) makePipeline(pipe Pipeline) (res *bitflow.SamplePipeline, err error) {
	var source bitflow.SampleSource

	switch input := pipe[0].(type) {
	case Input:
		inputs := make([]string, len(input))
		for i, in := range input {
			inputs[i] = in.Content()
		}
		source, err = b.Endpoints.CreateInput(inputs...)
	case MultiInput:
		source, err = b.createMultiInput(input)
	default:
		return nil, fmt.Errorf("Illegal input type %T: %v", input, input)
	}

	if err == nil {
		res, err = b.makePipelineTail(pipe[1:])
		res.Source = source
	}
	return
}

func (b PipelineBuilder) makePipelineTail(pipe Pipeline) (*bitflow.SamplePipeline, error) {
	res := new(bitflow.SamplePipeline)
	var err error
	for _, step := range pipe {
		switch step := step.(type) {
		case Output:
			err = b.addOutputStep(res, step)
		case Step:
			err = b.addStep(res, step)
		case Fork:
			err = b.addFork(res, step)
		default:
			err = ParserError{
				Pos:     step.Pos(),
				Message: fmt.Sprintf("Unsupported pipeline step type %T: %v", step, step),
			}
		}
		if err != nil {
			break
		}
	}
	return res, err
}

func (b PipelineBuilder) addOutputStep(pipe *bitflow.SamplePipeline, output Output) error {
	endpoint, err := b.Endpoints.CreateOutput(Token(output).Content())
	if err == nil {
		pipe.Add(endpoint)
	}
	return err
}

func (b PipelineBuilder) addStep(pipe *bitflow.SamplePipeline, step Step) error {
	analysis, err := b.getAnalysis(step.Name)
	if err != nil {
		return err
	}
	params := step.ParamsMap()
	err = analysis.Params.Verify(params)
	if err == nil {
		err = analysis.Func(pipe, params)
	}
	if err != nil {
		err = ParserError{
			Pos:     step.Name,
			Message: fmt.Sprintf("%v: %v", step.Name.Content(), err),
		}
	}
	return err
}

func (b PipelineBuilder) getAnalysis(name_tok Token) (reg.RegisteredAnalysis, error) {
	name := name_tok.Content()
	if analysis, ok := b.GetAnalysis(name); ok {
		return analysis, nil
	} else {
		return reg.RegisteredAnalysis{}, ParserError{
			Pos:     name_tok,
			Message: fmt.Sprintf("Pipeline step '%v' is unknown", name),
		}
	}
}

func (b PipelineBuilder) createMultiInput(pipes MultiInput) (bitflow.SampleSource, error) {
	subPipelines := new(fork.MultiMetricSource)
	for _, subPipe := range pipes.Pipelines {
		subPipe, err := b.makePipeline(subPipe)
		if err != nil {
			return nil, err
		}
		subPipelines.Add(subPipe)
	}
	return subPipelines, nil
}

func (b PipelineBuilder) addFork(pipe *bitflow.SamplePipeline, f Fork) error {
	forkStep, err := b.getFork(f.Name)
	var distributor fork.Distributor
	if err == nil {
		params := f.ParamsMap()
		err = forkStep.Params.Verify(params)
		if err == nil {
			subpipelines := b.prepareSubpipelines(f.Pipelines)
			regSubpipelines := make([]reg.Subpipeline, len(subpipelines))
			for i := range subpipelines {
				regSubpipelines[i] = &subpipelines[i]
			}
			distributor, err = forkStep.Func(regSubpipelines, params)
		}
	}
	if err != nil {
		return ParserError{
			Pos:     f.Name,
			Message: fmt.Sprintf("%v: %v", f.Name.Content(), err),
		}
	}
	pipe.Add(&fork.SampleFork{
		Distributor: distributor,
	})
	return nil
}

func (b PipelineBuilder) getFork(name_tok Token) (reg.RegisteredFork, error) {
	name := name_tok.Content()
	if res, ok := b.GetFork(name); ok {
		return res, nil
	} else {
		return reg.RegisteredFork{}, ParserError{
			Pos:     name_tok,
			Message: fmt.Sprintf("Pipeline fork '%v' is unknown", name),
		}
	}
}

func (b PipelineBuilder) prepareSubpipelines(pipelines Pipelines) []subpipeline {
	res := make([]subpipeline, len(pipelines))
	for i, pipe := range pipelines {
		inputs := pipe[0].(Input)
		res[i].keys = make([]string, len(inputs))
		for j, input := range inputs {
			res[i].keys[j] = input.Content()
		}
		res[i].pipe = pipe[1:]
		res[i].builder = &b
	}
	return res
}

// Implement the PipelineVerification interface

func (b PipelineBuilder) VerifyInput(inputs []string) error {
	// This allocates some objects, but no system resources.
	// TODO add some Verify* methods to EndpointFactory to avoid this.
	_, err := b.Endpoints.CreateInput(inputs...)
	return err
}

func (b PipelineBuilder) VerifyOutput(output string) error {
	_, err := b.Endpoints.CreateOutput(output)
	return err
}

func (b PipelineBuilder) VerifyStep(name Token, params map[string]string) error {
	analysis, err := b.getAnalysis(name)
	if err != nil {
		return err
	}
	return analysis.Params.Verify(params)
}

func (b PipelineBuilder) VerifyFork(name Token, params map[string]string) error {
	forkStep, err := b.getFork(name)
	if err != nil {
		return err
	}
	return forkStep.Params.Verify(params)
}
