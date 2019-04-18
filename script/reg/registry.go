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

type BatchStepFunc func(params map[string]string) (bitflow.BatchProcessingStep, error)

type ForkFunc func(subpiplines []Subpipeline, params map[string]string) (fork.Distributor, error)

type RegisteredParameters struct {
	Required []string
	Optional []string
}

type RegisteredStep struct {
	Name        string
	Description string
	Params      RegisteredParameters
}

type RegisteredPipelineStep struct {
	RegisteredStep
	Func AnalysisFunc
}

type RegisteredBatchStep struct {
	RegisteredStep
	Func BatchStepFunc
}

type RegisteredFork struct {
	RegisteredStep
	Func ForkFunc
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

	stepRegistry  map[string]RegisteredPipelineStep
	batchRegistry map[string]RegisteredBatchStep
	forkRegistry  map[string]RegisteredFork
}

func NewProcessorRegistry() ProcessorRegistry {
	reg := ProcessorRegistry{
		ProcessorRegistryImpl: &ProcessorRegistryImpl{
			Endpoints:     *bitflow.NewEndpointFactory(),
			stepRegistry:  make(map[string]RegisteredPipelineStep),
			batchRegistry: make(map[string]RegisteredBatchStep),
			forkRegistry:  make(map[string]RegisteredFork),
		},
	}
	RegisterMultiplexFork(reg)
	return reg
}

func (r *ProcessorRegistryImpl) GetStep(name string) (analysisProcessor RegisteredPipelineStep, ok bool) {
	analysisProcessor, ok = r.stepRegistry[name]
	return
}

func (r *ProcessorRegistryImpl) GetFork(name string) (fork RegisteredFork, ok bool) {
	fork, ok = r.forkRegistry[name]
	return
}

func (r *ProcessorRegistryImpl) GetBatchStep(name string) (batchStep RegisteredBatchStep, ok bool) {
	batchStep, ok = r.batchRegistry[name]
	return
}

type ParameterOption func(*RegisteredParameters)

func RequiredParams(params ...string) ParameterOption {
	return func(opts *RegisteredParameters) {
		opts.Required = append(opts.Required, params...)
	}
}

func OptionalParams(params ...string) ParameterOption {
	return func(opts *RegisteredParameters) {
		opts.Optional = append(opts.Optional, params...)
	}
}

func VariableParams() ParameterOption {
	return func(opts *RegisteredParameters) {
		// Two empty (but non-nil) slices are a marker, that this step accepts any parameters without validation
		opts.Optional = []string{}
		opts.Required = []string{}
	}
}

func applyParameterOptions(options []ParameterOption) RegisteredParameters {
	opts := RegisteredParameters{}
	for _, o := range options {
		o(&opts)
	}
	return opts
}

func (r *ProcessorRegistryImpl) RegisterStep(name string, setupPipeline AnalysisFunc, description string, options ...ParameterOption) {
	if _, ok := r.stepRegistry[name]; ok {
		panic("Analysis already registered: " + name)
	}
	params := applyParameterOptions(options)
	r.stepRegistry[name] = RegisteredPipelineStep{
		RegisteredStep: RegisteredStep{
			Name:        name,
			Description: description,
			Params:      params,
		},
		Func: setupPipeline,
	}
}

func (r *ProcessorRegistryImpl) RegisterFork(name string, createFork ForkFunc, description string, options ...ParameterOption) {
	if _, ok := r.forkRegistry[name]; ok {
		panic("Fork already registered: " + name)
	}
	opts := applyParameterOptions(options)
	params := RegisteredParameters{opts.Required, opts.Optional}
	r.forkRegistry[name] = RegisteredFork{RegisteredStep{name, description, params}, createFork}
}

func (r *ProcessorRegistryImpl) RegisterBatchStep(name string, createBatchStep BatchStepFunc, description string, options ...ParameterOption) {
	if _, ok := r.batchRegistry[name]; ok {
		panic("Batch step already registered: " + name)
	}
	opts := applyParameterOptions(options)
	params := RegisteredParameters{opts.Required, opts.Optional}
	r.batchRegistry[name] = RegisteredBatchStep{RegisteredStep{name, description, params}, createBatchStep}
}

func (params RegisteredParameters) AcceptsVariableParameters() bool {
	return params.Required != nil && params.Optional != nil && len(params.Required) == 0 && len(params.Optional) == 0
}

func (params RegisteredParameters) Verify(input map[string]string) error {
	if params.AcceptsVariableParameters() {
		return nil
	}
	checked := map[string]bool{}
	for _, opt := range params.Optional {
		checked[opt] = true
	}
	for _, req := range params.Required {
		if _, ok := input[req]; !ok {
			return fmt.Errorf("Missing Required parameter '%v'", req)
		}
		checked[req] = true
	}
	for key := range input {
		if _, ok := checked[key]; !ok {
			return fmt.Errorf("Unexpected parameter '%v'", key)
		}
	}
	return nil
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
	all := make(AvailableProcessorSlice, 0, len(r.stepRegistry))
	for _, step := range r.stepRegistry {
		if step.Func == nil {
			continue
		}
		all = append(all, AvailableProcessor{
			Name:           step.Name,
			IsFork:         false,
			Description:    step.Description,
			RequiredParams: step.Params.Required,
			OptionalParams: step.Params.Optional,
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
			RequiredParams: registeredFork.Params.Required,
			OptionalParams: registeredFork.Params.Optional,
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
