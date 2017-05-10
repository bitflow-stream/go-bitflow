package query

import (
	"bytes"
	"fmt"
	"sort"

	log "github.com/Sirupsen/logrus"
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
	return &PipelineBuilder{
		Endpoints:         *bitflow.NewEndpointFactory(),
		analysis_registry: make(map[string]registeredAnalysis),
		fork_registry:     make(map[string]registeredFork),
	}
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
		case Fork:
			err = builder.addFork(res, step)
		default:
			err = ParserError{
				Pos:     step.Pos(),
				Message: fmt.Sprintf("Unsupported pipeline step type: %T", step),
			}
		}
		if err != nil {
			break
		}
	}
	return res, err
}

func (builder PipelineBuilder) addStep(pipe *pipeline.SamplePipeline, step Step) error {
	stepFunc, err := builder.getAnalysis(step.Name)
	if err != nil {
		return err
	}
	params := builder.makeParams(step.Params)
	err = stepFunc(pipe, params)
	if err != nil {
		err = ParserError{
			Pos:     step.Name,
			Message: fmt.Sprintf("%v: %v", step.Name.Content(), err),
		}
	}
	return err
}

func (builder PipelineBuilder) makeParams(params map[Token]Token) map[string]string {
	res := make(map[string]string, len(params))
	for key, val := range params {
		res[key.Content()] = val.Content()
	}
	return res
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
	subPipelines := make(fork.MultiplexPipelineBuilder, num)
	for i, subPipe := range pipes {
		subPipe, err := builder.makePipeline(subPipe, false)
		if err != nil {
			return err
		}
		subPipelines[i] = subPipe
	}

	pipe.Add(&fork.MetricFork{
		ParallelClose: true,
		Distributor:   fork.NewMultiplexDistributor(num),
		Builder:       subPipelines,
	})
	return nil
}

func (builder PipelineBuilder) createMultiInput(pipes Pipelines) (bitflow.MetricSource, error) {
	subPipelines := &fork.MultiMetricSource{
		ParallelClose: true,
	}
	for _, subPipe := range pipes {
		subPipe, err := builder.makePipeline(subPipe, true)
		if err != nil {
			return nil, err
		}
		subPipelines.Add(subPipe)
	}
	return subPipelines, nil
}

func (builder PipelineBuilder) addFork(pipe *pipeline.SamplePipeline, f Fork) error {
	forkFunc, err := builder.getFork(f.Name)
	if err != nil {
		return err
	}
	params := builder.makeParams(f.Params)
	resDist, err := forkFunc(params)
	if err != nil {
		return err
	}
	forkBuilder, err := builder.makeCustomBuilder(f.Pipelines)
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

func (builder PipelineBuilder) makeCustomBuilder(pipelines Pipelines) (*CustomPipelineBuilder, error) {
	builderPipes := make(map[string]*pipeline.SamplePipeline)
	for _, pipe := range pipelines {
		inputs := pipe[0].(Input)
		builtPipe, err := builder.makePipeline(pipe[1:], false)
		if err != nil {
			return nil, err
		}
		for _, input := range inputs {
			builderPipes[input.Content()] = builtPipe
		}
	}
	return &CustomPipelineBuilder{
		pipelines: builderPipes,
	}, nil
}

func (builder PipelineBuilder) getFork(name_tok Token) (ForkFunc, error) {
	name := name_tok.Content()
	if registeredFork, ok := builder.fork_registry[name]; ok {
		return registeredFork.Func, nil
	} else {
		return nil, ParserError{
			Pos:     name_tok,
			Message: fmt.Sprintf("Pipeline fork '%v' is unknown", name),
		}
	}
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

type CustomPipelineBuilder struct {
	pipelines map[string]*pipeline.SamplePipeline
}

func (b *CustomPipelineBuilder) BuildPipeline(key interface{}, _ *fork.ForkMerger) *bitflow.SamplePipeline {
	strKey := ""
	if key != nil {
		strKey = fmt.Sprintf("%v", key)
	}
	pipe, ok := b.pipelines[strKey]
	if !ok {
		keys := make([]string, 0, len(b.pipelines))
		for key := range b.pipelines {
			keys = append(keys, key)
		}
		log.Warnf("No subpipeline defined for key '%v' (type %T). Using empty pipeline (Have pipelines: %v)", strKey, key, keys)
		pipe = new(pipeline.SamplePipeline)
	}
	return &pipe.SamplePipeline
}

func (b *CustomPipelineBuilder) String() string {
	return fmt.Sprintf("Pipeline builder, %v subpipelines", len(b.pipelines))
}

func (b *CustomPipelineBuilder) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, 0, len(b.pipelines))
	for key, pipe := range b.pipelines {
		var title string
		if key == "" {
			title = "Default pipeline"
		} else {
			title = fmt.Sprintf("Pipeline %v", key)
		}
		res = append(res, &titledSubPipeline{
			SamplePipeline: pipe,
			title:          title,
		})
	}
	sort.Sort(sortedStringers(res))
	return res
}

type titledSubPipeline struct {
	*pipeline.SamplePipeline
	title string
}

func (t *titledSubPipeline) String() string {
	return t.title
}

type sortedStringers []fmt.Stringer

func (t sortedStringers) Len() int {
	return len(t)
}

func (t sortedStringers) Less(a, b int) bool {
	return t[a].String() < t[b].String()
}

func (t sortedStringers) Swap(a, b int) {
	t[a], t[b] = t[b], t[a]
}
