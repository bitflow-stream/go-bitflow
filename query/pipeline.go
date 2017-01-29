package query

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/golib"
)

type AnalysisFunc func(pipeline *SamplePipeline, params map[string]string) error

type PipelineBuilder struct {
	Endpoints bitflow.EndpointFactory

	analysis_registry map[string]registeredAnalysis
}

type registeredAnalysis struct {
	Name        string
	Func        AnalysisFunc
	Description string
}

func (b *PipelineBuilder) RegisterAnalysis(name string, setupPipeline AnalysisFunc, description string) {
	if b.analysis_registry == nil {
		b.analysis_registry = make(map[string]registeredAnalysis)
	}
	if _, ok := b.analysis_registry[name]; ok {
		panic("Analysis already registered: " + name)
	}
	b.analysis_registry[name] = registeredAnalysis{name, setupPipeline, description}
}

func CheckParameters(params map[string]string, optional []string, required []string) error {
	checked := map[string]bool{}
	for _, opt := range optional {
		checked[opt] = true
	}
	for _, req := range required {
		if _, ok := params[req]; !ok {
			return fmt.Errorf("Missing required parameter %v", req)
		}
		checked[req] = true
	}
	for key := range params {
		if _, ok := checked[key]; !ok {
			return fmt.Errorf("Unexpected parameter %v", key)
		}
	}
	return nil
}

func (builder PipelineBuilder) MakePipeline(pipe Pipeline) (*SamplePipeline, error) {
	if len(pipe) == 0 {
		return nil, ParserError{
			Pos:     pipe.Pos(),
			Message: "Empty pipeline is not allowed",
		}
	}
	endpoints := builder.Endpoints // Copy all parmeters

	if input, ok := pipe[0].(Input); ok {
		pipe = pipe[1:]
		for _, in := range input {
			endpoints.FlagInputs = append(endpoints.FlagInputs, in.Content())
		}
	}
	if len(pipe) >= 1 {
		if output, ok := pipe[len(pipe)-1].(Output); ok {
			pipe = pipe[:len(pipe)-1]
			endpoints.FlagOutputs = golib.StringSlice{Token(output).Content()}
		}
	}
	res := new(SamplePipeline)
	if err := res.Configure(&endpoints); err != nil {
		return nil, err
	}

	for _, step := range pipe {
		switch step := step.(type) {
		case Step:
			if err := builder.addStep(res, step); err != nil {
				return nil, err
			}
		case Pipelines:
			if err := builder.addMultiplex(res, step); err != nil {
				return nil, err
			}
		default:
			return nil, ParserError{
				Pos:     step.Pos(),
				Message: fmt.Sprintf("Unsupported pipeline step type: %T", step),
			}
		}
	}
	return res, nil
}

func (builder PipelineBuilder) addStep(pipe *SamplePipeline, step Step) error {
	stepFunc, err := builder.getAnalysis(step.Name)
	if err != nil {
		return err
	}
	params := make(map[string]string, len(step.Params))
	for key, val := range step.Params {
		params[key.Content()] = val.Content()
	}
	err = stepFunc(pipe, params)
	if err != nil {
		err = ParserError{
			Pos:     step.Name,
			Message: fmt.Sprintf("%v: %v", step.Name.Content(), err),
		}
	}
	return err
}

func (builder PipelineBuilder) getAnalysis(name_tok Token) (AnalysisFunc, error) {
	name := name_tok.Content()
	if analysis, ok := builder.analysis_registry[name]; ok {
		return analysis.Func, nil
	} else {
		return nil, ParserError{
			Pos:     name_tok,
			Message: fmt.Sprintf("Pipeline step '%v' is unknown", name),
		}
	}
}

func (builder PipelineBuilder) addMultiplex(pipe *SamplePipeline, pipes Pipelines) error {
	num := len(pipes) // Must be the same for the builder and the distributor
	subpipelines := make(MultiplexPipelineBuilder, num)
	for i, subpipe := range pipes {
		subpipe, err := builder.MakePipeline(subpipe)
		if err != nil {
			return err
		}
		subpipelines[i] = subpipe
	}

	// TODO control/configure parallelism of the Fork

	pipe.Add(&pipeline.MetricFork{
		ParallelClose: true,
		Distributor:   pipeline.NewMultiplexDistributor(num),
		Builder:       subpipelines,
	})
	return nil
}

type MultiplexPipelineBuilder []*SamplePipeline

func (b MultiplexPipelineBuilder) BuildPipeline(key interface{}, output *pipeline.ForkMerger) *bitflow.SamplePipeline {
	pipe := &(b[key.(int)].SamplePipeline) // Type of key must be int, and index must be in range
	if pipe.Sink == nil {
		pipe.Sink = output
	}
	return pipe
}

func (b MultiplexPipelineBuilder) String() string {
	return fmt.Sprintf("Multiplex builder, %v subpipelines", len(b))
}

func (b MultiplexPipelineBuilder) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, len(b))
	for i, pipe := range b {
		res[i] = pipe
	}
	return res
}

type SamplePipeline struct {
	bitflow.SamplePipeline
	lastProcessor bitflow.SampleProcessor
}

func (p *SamplePipeline) Add(step bitflow.SampleProcessor) *SamplePipeline {
	if p.lastProcessor != nil {
		if merger, ok := p.lastProcessor.(pipeline.MergableProcessor); ok {
			if merger.MergeProcessor(step) {
				// Merge successful: drop the incoming step
				return p
			}
		}
	}
	p.lastProcessor = step
	p.SamplePipeline.Add(step)
	return p
}

func (p *SamplePipeline) Batch(steps ...pipeline.BatchProcessingStep) *SamplePipeline {
	batch := new(pipeline.BatchProcessor)
	for _, step := range steps {
		batch.Add(step)
	}
	return p.Add(batch)
}
