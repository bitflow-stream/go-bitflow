package reg

import (
	"fmt"
	"sort"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/bitflow/fork"
)

const (
	MultiplexForkName = "multiplex"
)

type AnalysisFunc func(pipeline *bitflow.SamplePipeline, params map[string]string) error

type ForkFunc func(subpiplines []Subpipeline, params map[string]string) (fork.Distributor, error)

type RegisteredAnalysis struct {
	Name                     string
	Func                     AnalysisFunc
	Description              string
	Params                   registeredParameters
	SupportsBatchProcessing  bool
	SupportsStreamProcessing bool
}

type RegisteredFork struct {
	Name        string
	Func        ForkFunc
	Description string
	Params      registeredParameters
}

type registeredParameters struct {
	required []string
	optional []string
}

type Subpipeline interface {
	Build() (*bitflow.SamplePipeline, error)
	Keys() []string
}

// Avoid having to specify a pointer to a ProcessorRegistry at many locations.
type ProcessorRegistry struct {
	*ProcessorRegistryImpl
}

type ProcessorRegistryImpl struct {
	Endpoints bitflow.EndpointFactory

	analysisRegistry map[string]RegisteredAnalysis
	forkRegistry     map[string]RegisteredFork
}

func NewProcessorRegistry() ProcessorRegistry {
	reg := ProcessorRegistry{
		ProcessorRegistryImpl: &ProcessorRegistryImpl{
			Endpoints:        *bitflow.NewEndpointFactory(),
			analysisRegistry: make(map[string]RegisteredAnalysis),
			forkRegistry:     make(map[string]RegisteredFork),
		},
	}
	RegisterMultiplexFork(reg)
	return reg
}

func (r *ProcessorRegistryImpl) GetAnalysis(name string) (analysisProcessor RegisteredAnalysis, ok bool) {
	analysisProcessor, ok = r.analysisRegistry[name]
	return
}

func (r *ProcessorRegistryImpl) GetFork(name string) (fork RegisteredFork, ok bool) {
	fork, ok = r.forkRegistry[name]
	return
}

func (r *ProcessorRegistryImpl) RegisterAnalysisParamsErr(name string, setupPipeline AnalysisFunc, description string, options ...Option) {
	if _, ok := r.analysisRegistry[name]; ok {
		panic("Analysis already registered: " + name)
	}
	opts := GetOpts(options)
	params := registeredParameters{opts.RequiredParams, opts.OptionalParams}
	r.analysisRegistry[name] = RegisteredAnalysis{
		Name:                     name,
		Func:                     setupPipeline,
		Description:              params.makeDescription(description),
		Params:                   params,
		SupportsBatchProcessing:  opts.SupportBatchProcessing,
		SupportsStreamProcessing: !opts.EnforceBatchProcessing,
	}
}

func (r *ProcessorRegistryImpl) RegisterAnalysisParams(name string, setupPipeline func(pipeline *bitflow.SamplePipeline, params map[string]string), description string, options ...Option) {
	r.RegisterAnalysisParamsErr(name, func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
		setupPipeline(pipeline, params)
		return nil
	}, description, options...)
}

func (r *ProcessorRegistryImpl) RegisterAnalysis(name string, setupPipeline func(pipeline *bitflow.SamplePipeline), description string, options ...Option) {
	r.RegisterAnalysisParams(name, func(pipeline *bitflow.SamplePipeline, _ map[string]string) {
		setupPipeline(pipeline)
	}, description, options...)
}

func (r *ProcessorRegistryImpl) RegisterAnalysisErr(name string, setupPipeline func(pipeline *bitflow.SamplePipeline) error, description string, options ...Option) {
	r.RegisterAnalysisParamsErr(name, func(pipeline *bitflow.SamplePipeline, _ map[string]string) error {
		return setupPipeline(pipeline)
	}, description, options...)
}

func (r *ProcessorRegistryImpl) RegisterFork(name string, createFork ForkFunc, description string, options ...Option) {
	if _, ok := r.forkRegistry[name]; ok {
		panic("Fork already registered: " + name)
	}
	opts := GetOpts(options)
	params := registeredParameters{opts.RequiredParams, opts.OptionalParams}
	r.forkRegistry[name] = RegisteredFork{name, createFork, params.makeDescription(description), params}
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

func (r ProcessorRegistryImpl) GetAvailableProcessors() AvailableProcessorSlice {
	all := make(AvailableProcessorSlice, 0, len(r.analysisRegistry))
	for _, step := range r.analysisRegistry {
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
	for _, registeredFork := range r.forkRegistry {
		if registeredFork.Func == nil {
			continue
		}
		all = append(all, AvailableProcessor{
			Name:           registeredFork.Name,
			IsFork:         true,
			Description:    registeredFork.Description,
			RequiredParams: registeredFork.Params.required,
			OptionalParams: registeredFork.Params.optional,
		})
	}
	sort.Sort(all)
	return all
}

// default Fork
func RegisterMultiplexFork(builder ProcessorRegistry) {
	builder.RegisterFork(MultiplexForkName, createMultiplexFork, "Basic fork forwarding samples to all subpipelines. Subpipeline keys are ignored.")
}

func createMultiplexFork(subpipelines []Subpipeline, _ map[string]string) (fork.Distributor, error) {
	var res fork.MultiplexDistributor
	res.Subpipelines = make([]*bitflow.SamplePipeline, len(subpipelines))
	var err error
	for i, pipe := range subpipelines {
		res.Subpipelines[i], err = pipe.Build()
		if err != nil {
			return nil, err
		}
	}
	return &res, nil
}
