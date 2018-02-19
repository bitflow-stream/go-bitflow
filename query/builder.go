package query

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/fork"
)

type PipelineBuilder struct {
	Endpoints bitflow.EndpointFactory

	analysis_registry map[string]registeredAnalysis
	fork_registry     map[string]registeredFork
}

func NewPipelineBuilder() *PipelineBuilder {
	builder := &PipelineBuilder{
		Endpoints:         *bitflow.NewEndpointFactory(),
		analysis_registry: make(map[string]registeredAnalysis),
		fork_registry:     make(map[string]registeredFork),
	}
	RegisterMultiplexFork(builder)
	return builder
}

func (builder PipelineBuilder) MakePipeline(pipe Pipeline) (*pipeline.SamplePipeline, error) {
	pipe, err := pipe.Transform(builder)
	if err != nil {
		return nil, err
	}
	return builder.makePipeline(pipe)
}

func (builder PipelineBuilder) makePipeline(pipe Pipeline) (res *pipeline.SamplePipeline, err error) {
	var source bitflow.MetricSource

	switch input := pipe[0].(type) {
	case Input:
		inputs := make([]string, len(input))
		for i, in := range input {
			inputs[i] = in.Content()
		}
		source, err = builder.Endpoints.CreateInput(inputs...)
	case MultiInput:
		source, err = builder.createMultiInput(input)
	default:
		return nil, fmt.Errorf("Illegal input type %T: %v", input, input)
	}

	if err == nil {
		res, err = builder.makePipelineTail(pipe[1:])
		res.Source = source
	}
	return
}

func (builder PipelineBuilder) makePipelineTail(pipe Pipeline) (res *pipeline.SamplePipeline, err error) {
	res = new(pipeline.SamplePipeline)

	// Output
	outputStep := pipe[len(pipe)-1]
	if output, ok := outputStep.(Output); ok {
		pipe = pipe[:len(pipe)-1]
		res.Sink, err = builder.Endpoints.CreateOutput(Token(output).Content())
		if err != nil {
			return
		}
	}

	// Steps
	for _, step := range pipe {
		switch step := step.(type) {
		case Step:
			err = builder.addStep(res, step)
		case Fork:
			err = builder.addFork(res, step)
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

func (builder PipelineBuilder) addStep(pipe *pipeline.SamplePipeline, step Step) error {
	analysis, err := builder.getAnalysis(step.Name)
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

func (builder PipelineBuilder) getAnalysis(name_tok Token) (registeredAnalysis, error) {
	name := name_tok.Content()
	if analysis, ok := builder.analysis_registry[name]; ok {
		return analysis, nil
	} else {
		return registeredAnalysis{}, ParserError{
			Pos:     name_tok,
			Message: fmt.Sprintf("Pipeline step '%v' is unknown", name),
		}
	}
}

func (builder PipelineBuilder) createMultiInput(pipes MultiInput) (bitflow.MetricSource, error) {
	subPipelines := &fork.MultiMetricSource{
		ParallelClose: true,
	}
	for _, subPipe := range pipes.Pipelines {
		subPipe, err := builder.makePipeline(subPipe)
		if err != nil {
			return nil, err
		}
		subPipelines.Add(subPipe)
	}
	return subPipelines, nil
}

func (builder PipelineBuilder) addFork(pipe *pipeline.SamplePipeline, f Fork) error {
	forkStep, err := builder.getFork(f.Name)
	if err != nil {
		return err
	}
	params := f.ParamsMap()
	err = forkStep.Params.Verify(params)
	if err != nil {
		return err
	}
	resDist, err := forkStep.Func(params)
	if err != nil {
		return err
	}
	forkBuilder, err := builder.makePipelineBuilder(f.Pipelines)
	if err != nil {
		return err
	}

	switch distributor := resDist.(type) {
	case fork.ForkDistributor:
		pipe.Add(&fork.MetricFork{
			ParallelClose: true,
			Builder:       forkBuilder,
			Distributor:   distributor,
		})
	case fork.RemapDistributor:
		pipe.Add(&fork.ForkRemapper{
			ParallelClose: true,
			Builder:       forkBuilder,
			Distributor:   distributor,
		})
	default:
		return ParserError{
			Pos: f.Name,
			Message: fmt.Sprintf("Fork func %v returned illegal result (need fork.ForkDistributor or fork.RemapDistributor): %T (%v)",
				f.Name.Content(), resDist, resDist),
		}
	}
	return nil
}

func (builder PipelineBuilder) getFork(name_tok Token) (registeredFork, error) {
	name := name_tok.Content()
	if res, ok := builder.fork_registry[name]; ok {
		return res, nil
	} else {
		return registeredFork{}, ParserError{
			Pos:     name_tok,
			Message: fmt.Sprintf("Pipeline fork '%v' is unknown", name),
		}
	}
}

func (builder PipelineBuilder) makePipelineBuilder(pipelines Pipelines) (fork.PipelineBuilder, error) {
	builderPipes := make(map[string]*pipeline.SamplePipeline)
	var defaultTail Pipeline
	for _, pipe := range pipelines {
		inputs := pipe[0].(Input)
		builtPipe, err := builder.makePipelineTail(pipe[1:])
		if err != nil {
			return nil, err
		}
		for _, input := range inputs {
			builderPipes[input.Content()] = builtPipe
			if input.Content() == "" {
				defaultTail = pipe[1:]
			}
		}
	}
	return &fork.StringPipelineBuilder{
		Pipelines: builderPipes,
		BuildMissingPipeline: func(string) (*pipeline.SamplePipeline, error) {
			if len(defaultTail) == 0 {
				return new(pipeline.SamplePipeline), nil
			} else {
				return builder.makePipelineTail(defaultTail)
			}
		},
	}, nil
}

// Implement the PipelineVerification interface

func (builder PipelineBuilder) VerifyInput(inputs []string) error {
	// This allocates some objects, but no system resources.
	// TODO add some Verify* methods to EndpointFactory to avoid this.
	_, err := builder.Endpoints.CreateInput(inputs...)
	return err
}

func (builder PipelineBuilder) VerifyOutput(output string) error {
	_, err := builder.Endpoints.CreateOutput(output)
	return err
}

func (builder PipelineBuilder) VerifyStep(name Token, params map[string]string) error {
	analysis, err := builder.getAnalysis(name)
	if err != nil {
		return err
	}
	return analysis.Params.Verify(params)
}

func (builder PipelineBuilder) VerifyFork(name Token, params map[string]string) error {
	forkStep, err := builder.getFork(name)
	if err != nil {
		return err
	}
	return forkStep.Params.Verify(params)
}
