package query

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
)

type AnalysisFunc func(pipeline *pipeline.SamplePipeline, params map[string]string) error

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

func (builder PipelineBuilder) MakePipeline(pipe Pipeline) (*pipeline.SamplePipeline, error) {
	res, err := builder.makePipeline(pipe, true)
	if err == nil {
		if res.Source == nil {
			res.Source = new(bitflow.EmptyMetricSource)
		}
		if res.Sink == nil {
			res.Sink = new(bitflow.EmptyMetricSink)
		}
	}
	return res, err
}

func (builder PipelineBuilder) makePipeline(pipe Pipeline, isInput bool) (*pipeline.SamplePipeline, error) {
	if len(pipe) == 0 {
		return nil, ParserError{
			Pos:     pipe.Pos(),
			Message: "Empty pipeline is not allowed",
		}
	}
	res := new(pipeline.SamplePipeline)
	var err error

	switch input := pipe[0].(type) {
	case Input:
		pipe = pipe[1:]
		inputs := make([]string, len(input))
		for i, in := range input {
			inputs[i] = in.Content()
		}
		res.Source, err = builder.Endpoints.CreateInput(inputs...)
		if err != nil {
			return nil, err
		}
	case Pipelines:
		if isInput {
			pipe = pipe[1:]
			res.Source, err = builder.createMultiInput(input)
			if err != nil {
				return nil, err
			}
		}
	}
	if len(pipe) >= 1 {
		if output, ok := pipe[len(pipe)-1].(Output); ok {
			pipe = pipe[:len(pipe)-1]
			res.Sink, err = builder.Endpoints.CreateOutput(Token(output).Content())
			if err != nil {
				return nil, err
			}
		}
	}

	for _, step := range pipe {
		switch step := step.(type) {
		case Step:
			err = builder.addStep(res, step)
		case Pipelines:
			err = builder.addMultiplex(res, step)
		default:
			err = ParserError{
				Pos:     step.Pos(),
				Message: fmt.Sprintf("Unsupported pipeline step type: %T", step),
			}
		}
	}
	return res, err
}

func (builder PipelineBuilder) addStep(pipe *pipeline.SamplePipeline, step Step) error {
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

func (builder PipelineBuilder) addMultiplex(pipe *pipeline.SamplePipeline, pipes Pipelines) error {
	num := len(pipes) // Must be the same for the builder and the distributor
	subpipelines := make(MultiplexPipelineBuilder, num)
	for i, subpipe := range pipes {
		subpipe, err := builder.makePipeline(subpipe, false)
		if err != nil {
			return err
		}
		subpipelines[i] = subpipe
	}

	// TODO control/configure parallelism of the Fork

	pipe.Add(&pipeline.MetricFork{
		MultiPipeline: pipeline.MultiPipeline{
			ParallelClose: true,
		},
		Distributor: pipeline.NewMultiplexDistributor(num),
		Builder:     subpipelines,
	})
	return nil
}

func (builder PipelineBuilder) createMultiInput(pipes Pipelines) (bitflow.MetricSource, error) {

	// TODO control/configure parallelism of the Fork

	subpipelines := &pipeline.MultiMetricSource{
		MultiPipeline: pipeline.MultiPipeline{
			ParallelClose: true,
		},
	}
	for _, subpipe := range pipes {
		subpipe, err := builder.makePipeline(subpipe, true)
		if err != nil {
			return nil, err
		}
		subpipelines.Add(subpipe)
	}
	return subpipelines, nil
}

func (builder PipelineBuilder) PrintAllAnalyses() string {
	all := make(SortedAnalyses, 0, len(builder.analysis_registry))
	for _, analysis := range builder.analysis_registry {
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
		buf.WriteString(":\n")
		buf.WriteString("      ")
		buf.WriteString(analysis.Description)
	}
	return buf.String()
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

type MultiplexPipelineBuilder []*pipeline.SamplePipeline

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
