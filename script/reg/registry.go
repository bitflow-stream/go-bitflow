package reg

import (
	"fmt"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/bitflow/fork"
)

const (
	MultiplexForkName = "multiplex"
)

type AnalysisFunc func(pipeline *bitflow.SamplePipeline, params map[string]interface{}) error

type BatchStepFunc func(params map[string]interface{}) (bitflow.BatchProcessingStep, error)

type ForkFunc func(subpiplines []Subpipeline, params map[string]interface{}) (fork.Distributor, error)

type RegisteredParameters map[string]RegisteredParameter

type RegisteredParameter struct {
	Name     string
	Parser   ParameterParser
	Default  interface{}
	Required bool
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

	stepRegistry  map[string]*RegisteredPipelineStep
	batchRegistry map[string]*RegisteredBatchStep
	forkRegistry  map[string]*RegisteredFork
}

func NewProcessorRegistry() ProcessorRegistry {
	reg := ProcessorRegistry{
		ProcessorRegistryImpl: &ProcessorRegistryImpl{
			Endpoints:     *bitflow.NewEndpointFactory(),
			stepRegistry:  make(map[string]*RegisteredPipelineStep),
			batchRegistry: make(map[string]*RegisteredBatchStep),
			forkRegistry:  make(map[string]*RegisteredFork),
		},
	}
	RegisterMultiplexFork(reg)
	return reg
}

func (r *ProcessorRegistryImpl) GetStep(name string) *RegisteredPipelineStep {
	return r.stepRegistry[name]
}

func (r *ProcessorRegistryImpl) GetFork(name string) *RegisteredFork {
	return r.forkRegistry[name]
}

func (r *ProcessorRegistryImpl) GetBatchStep(name string) *RegisteredBatchStep {
	return r.batchRegistry[name]
}

func (r *RegisteredStep) Param(name string, parser ParameterParser, defaultValue interface{}, isRequired bool) *RegisteredStep {
	r.Params.Param(name, parser, defaultValue, isRequired)
	return r
}

func (r *RegisteredStep) Optional(name string, parser ParameterParser, defaultValue interface{}) *RegisteredStep {
	r.Params.Optional(name, parser, defaultValue)
	return r
}

func (r *RegisteredStep) Required(name string, parser ParameterParser) *RegisteredStep {
	r.Params.Required(name, parser)
	return r
}

func (params RegisteredParameters) Param(name string, parser ParameterParser, defaultValue interface{}, isRequired bool) RegisteredParameters {
	if _, ok := params[name]; ok {
		panic(fmt.Sprintf("Parameter %v already registered", name))
	}
	if !isRequired && !parser.CorrectType(defaultValue) {
		panic(fmt.Sprintf("Default value for parameter '%v' (type %v) is of unexpected type %T: %v", name, parser, defaultValue, defaultValue))
	}
	params[name] = RegisteredParameter{
		Name:     name,
		Parser:   parser,
		Default:  defaultValue,
		Required: isRequired,
	}
	return params
}

func (params RegisteredParameters) Optional(name string, parser ParameterParser, defaultValue interface{}) RegisteredParameters {
	return params.Param(name, parser, defaultValue, false)
}

func (params RegisteredParameters) Required(name string, parser ParameterParser) RegisteredParameters {
	return params.Param(name, parser, nil, true)
}

func (params RegisteredParameters) ParsePrimitives(stringParams map[string]string) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(stringParams))
	for key, val := range stringParams {
		param, ok := params[key]
		if !ok {
			return nil, fmt.Errorf("Unexpected parameter %v = %v", key, val)
		}
		var err error
		result[key], err = param.Parser.ParsePrimitive(val)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Check if required parameters are defined and fill defaults for optional parameters
func (params RegisteredParameters) ValidateAndSetDefaults(parsedValues map[string]interface{}) error {
	for key, param := range params {
		if _, paramDefined := parsedValues[key]; !paramDefined {
			if param.Required {
				return fmt.Errorf("Missing required parameter '%v' (type %v)", key, param.Parser)
			} else {
				parsedValues[key] = param.Default
			}
		}
		if val := parsedValues[key]; !param.Parser.CorrectType(val) {
			return fmt.Errorf("Value for parameter '%v' (type %v) is of unexpected type %T: %v", key, param.Parser, val, val)
		}
	}
	return nil
}

func (r *RegisteredParameter) String() string {
	defaultStr := ""
	if r.Default != nil {
		defaultStr = fmt.Sprintf(", default: %v", r.Default)
	}
	return fmt.Sprintf("%v (%v%v)", r.Name, r.Parser, defaultStr)
}

func (r *ProcessorRegistryImpl) RegisterStep(name string, setupPipeline AnalysisFunc, description string) *RegisteredStep {
	if _, ok := r.stepRegistry[name]; ok {
		panic("Analysis already registered: " + name)
	}
	newStep := &RegisteredPipelineStep{
		RegisteredStep: RegisteredStep{
			Name:        name,
			Description: description,
			Params:      make(RegisteredParameters),
		},
		Func: setupPipeline,
	}
	r.stepRegistry[name] = newStep
	return &newStep.RegisteredStep
}

func (r *ProcessorRegistryImpl) RegisterFork(name string, createFork ForkFunc, description string) *RegisteredStep {
	if _, ok := r.forkRegistry[name]; ok {
		panic("Fork already registered: " + name)
	}
	newFork := &RegisteredFork{
		RegisteredStep: RegisteredStep{
			Name:        name,
			Description: description,
			Params:      make(RegisteredParameters),
		},
		Func: createFork,
	}
	r.forkRegistry[name] = newFork
	return &newFork.RegisteredStep
}

func (r *ProcessorRegistryImpl) RegisterBatchStep(name string, createBatchStep BatchStepFunc, description string) *RegisteredStep {
	if _, ok := r.batchRegistry[name]; ok {
		panic("Batch step already registered: " + name)
	}
	newBatchStep := &RegisteredBatchStep{
		RegisteredStep: RegisteredStep{
			Name:        name,
			Description: description,
			Params:      make(RegisteredParameters),
		},
		Func: createBatchStep,
	}
	r.batchRegistry[name] = newBatchStep
	return &newBatchStep.RegisteredStep
}

func RegisterMultiplexFork(builder ProcessorRegistry) {
	builder.RegisterFork(MultiplexForkName, createMultiplexFork,
		"Basic fork forwarding samples to all subpipelines. Subpipeline keys are ignored.")
}

func createMultiplexFork(subpipelines []Subpipeline, _ map[string]interface{}) (fork.Distributor, error) {
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
