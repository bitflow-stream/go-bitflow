package script

import (
	"fmt"
	"sort"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
	"github.com/antongulenko/go-bitflow-pipeline/fork"
)

const (
	MultiplexForkName = "multiplex"
)

var _ builder.PipelineBuilder = NewProcessorRegistry()

type registeredAnalysis struct {
	Name                     string
	Func                     builder.AnalysisFunc
	Description              string
	Params                   registeredParameters
	SupportsBatchProcessing  bool
	SupportsStreamProcessing bool
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

type Subpipeline struct {
	keys []string
	pipe *pipeline.SamplePipeline
}

func (s *Subpipeline) Keys() []string {
	return s.keys
}

func (s *Subpipeline) Build() (*pipeline.SamplePipeline, error) {
	return s.pipe, nil
}

type ProcessorRegistry struct {
	Endpoints bitflow.EndpointFactory

	analysisRegistry map[string]registeredAnalysis
	forkRegistry     map[string]registeredFork
}

func NewProcessorRegistry() *ProcessorRegistry {
	registry := &ProcessorRegistry{
		Endpoints:        *bitflow.NewEndpointFactory(),
		analysisRegistry: make(map[string]registeredAnalysis),
		forkRegistry:     make(map[string]registeredFork),
	}
	return registry
}

func (reg ProcessorRegistry) getAnalysis(name string) (analysisProcessor registeredAnalysis, ok bool) {
	analysisProcessor, ok = reg.analysisRegistry[name]
	return
}

func (reg ProcessorRegistry) getFork(name string) (fork registeredFork, ok bool) {
	fork, ok = reg.forkRegistry[name]
	return
}

func parameterError(name string, err error) error {
	return fmt.Errorf("Failed to parse '%v' parameter: %v", name, err)
}

func (reg *ProcessorRegistry) RegisterAnalysisParamsErr(name string, setupPipeline builder.AnalysisFunc, description string, options ...builder.Option) {
	if _, ok := reg.analysisRegistry[name]; ok {
		panic("Analysis already registered: " + name)
	}
	opts := builder.GetOpts(options)
	params := registeredParameters{opts.RequiredParams, opts.OptionalParams}
	reg.analysisRegistry[name] = registeredAnalysis{
		Name:                     name,
		Func:                     setupPipeline,
		Description:              params.makeDescription(description),
		Params:                   params,
		SupportsBatchProcessing:  opts.SupportBatchProcessing,
		SupportsStreamProcessing: !opts.EnforceBatchProcessing,
	}
}

func (reg *ProcessorRegistry) RegisterAnalysisParams(name string, setupPipeline func(pipeline *pipeline.SamplePipeline, params map[string]string), description string, options ...builder.Option) {
	reg.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		setupPipeline(pipeline, params)
		return nil
	}, description, options...)
}

func (reg *ProcessorRegistry) RegisterAnalysis(name string, setupPipeline func(pipeline *pipeline.SamplePipeline), description string, options ...builder.Option) {
	reg.RegisterAnalysisParams(name, func(pipeline *pipeline.SamplePipeline, _ map[string]string) {
		setupPipeline(pipeline)
	}, description, options...)
}

func (reg *ProcessorRegistry) RegisterAnalysisErr(name string, setupPipeline func(pipeline *pipeline.SamplePipeline) error, description string, options ...builder.Option) {
	reg.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, _ map[string]string) error {
		return setupPipeline(pipeline)
	}, description, options...)
}

func (reg *ProcessorRegistry) RegisterFork(name string, createFork builder.ForkFunc, description string, options ...builder.Option) {
	if _, ok := reg.forkRegistry[name]; ok {
		panic("Fork already registered: " + name)
	}
	opts := builder.GetOpts(options)
	params := registeredParameters{opts.RequiredParams, opts.OptionalParams}
	reg.forkRegistry[name] = registeredFork{name, createFork, params.makeDescription(description), params}
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

type AvailableProcessor struct {
	Name           string
	IsFork         bool
	Description    string
	RequiredParams []string
	OptionalParams []string
}

type AvailableProcessorSlice []AvailableProcessor

func (slice AvailableProcessorSlice) Len() int {
	return len(slice)
}

func (slice AvailableProcessorSlice) Less(i, j int) bool {
	// Sort the forks after the regular processing steps
	return (!slice[i].IsFork && slice[j].IsFork) || slice[i].Name < slice[j].Name
}

func (slice AvailableProcessorSlice) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (reg ProcessorRegistry) GetAvailableProcessors() AvailableProcessorSlice {
	all := make(AvailableProcessorSlice, 0, len(reg.analysisRegistry))
	for _, step := range reg.analysisRegistry {
		if step.Func == nil {
			continue
		}
		all = append(all, AvailableProcessor{
			Name:           step.Name,
			IsFork:         false,
			Description:    step.Description,
			RequiredParams: step.Params.required,
			OptionalParams: step.Params.optional,
		})
	}
	for _, fork := range reg.forkRegistry {
		if fork.Func == nil {
			continue
		}
		all = append(all, AvailableProcessor{
			Name:           fork.Name,
			IsFork:         true,
			Description:    fork.Description,
			RequiredParams: fork.Params.required,
			OptionalParams: fork.Params.optional,
		})
	}
	sort.Sort(all)
	return all
}

// default Fork
func RegisterMultiplexFork(builder builder.PipelineBuilder) {
	builder.RegisterFork(MultiplexForkName, createMultiplexFork, "Basic fork forwarding samples to all subpipelines. Subpipeline keys are ignored.")
}

func createMultiplexFork(subpipelines []builder.Subpipeline, _ map[string]string) (fork.Distributor, error) {
	var res fork.MultiplexDistributor
	res.Subpipelines = make([]*pipeline.SamplePipeline, len(subpipelines))
	var err error
	for i, pipe := range subpipelines {
		res.Subpipelines[i], err = pipe.Build()
		if err != nil {
			return nil, err
		}
	}
	return &res, nil
}
