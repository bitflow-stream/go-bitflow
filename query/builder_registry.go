package query

import (
	"fmt"
	"strconv"
	"time"

	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/fork"
)

type Subpipeline struct {
	Keys []string

	builder *PipelineBuilder
	pipe    Pipeline
}

func (s *Subpipeline) Build() (*pipeline.SamplePipeline, error) {
	return s.builder.makePipelineTail(s.pipe)
}

type AnalysisFunc func(pipeline *pipeline.SamplePipeline, params map[string]string) error
type ForkFunc func(subpiplines []Subpipeline, params map[string]string) (fork.ForkDistributor, error)

func ParameterError(name string, err error) error {
	return fmt.Errorf("Failed to parse '%v' parameter: %v", name, err)
}

func StrParam(params map[string]string, name string, defaultVal string, hasDefault bool, err *error) string {
	if *err != nil {
		return ""
	}
	strVal, ok := params[name]
	if ok {
		return strVal
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return ""
	}
}

func BoolParam(params map[string]string, name string, defaultVal bool, hasDefault bool, err *error) bool {
	if *err != nil {
		return false
	}
	strVal, ok := params[name]
	if ok {
		parsed, parseErr := strconv.ParseBool(strVal)
		if parseErr != nil {
			*err = ParameterError(name, parseErr)
			return false
		}
		return parsed
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return false
	}
}

func DurationParam(params map[string]string, name string, defaultVal time.Duration, hasDefault bool, err *error) time.Duration {
	if *err != nil {
		return 0
	}
	strVal, ok := params[name]
	if ok {
		parsed, parseErr := time.ParseDuration(strVal)
		if parseErr != nil {
			*err = ParameterError(name, parseErr)
			return 0
		}
		return parsed
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return 0
	}
}

func IntParam(params map[string]string, name string, defaultVal int, hasDefault bool, err *error) int {
	if *err != nil {
		return 0
	}
	strVal, ok := params[name]
	if ok {
		parsed, parseErr := strconv.Atoi(strVal)
		if parseErr != nil {
			*err = ParameterError(name, parseErr)
			return 0
		}
		return parsed
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return 0
	}
}

func FloatParam(params map[string]string, name string, defaultVal float64, hasDefault bool, err *error) float64 {
	if *err != nil {
		return 0
	}
	strVal, ok := params[name]
	if ok {
		parsed, parseErr := strconv.ParseFloat(strVal, 64)
		if parseErr != nil {
			*err = ParameterError(name, parseErr)
		}
		return parsed
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return 0
	}
}

func (b *PipelineBuilder) RegisterAnalysisParamsErr(name string, setupPipeline AnalysisFunc, description string, requiredParams []string, optionalParams ...string) {
	if _, ok := b.analysis_registry[name]; ok {
		panic("Analysis already registered: " + name)
	}
	params := registeredParameters{requiredParams, optionalParams}
	b.analysis_registry[name] = registeredAnalysis{name, setupPipeline, params.makeDescription(description), params}
}

func (b *PipelineBuilder) RegisterAnalysisParams(name string, setupPipeline func(pipeline *pipeline.SamplePipeline, params map[string]string), description string, requiredParams []string, optionalParams ...string) {
	b.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		setupPipeline(pipeline, params)
		return nil
	}, description, requiredParams, optionalParams...)
}

func (b *PipelineBuilder) RegisterAnalysis(name string, setupPipeline func(pipeline *pipeline.SamplePipeline), description string) {
	b.RegisterAnalysisParams(name, func(pipeline *pipeline.SamplePipeline, _ map[string]string) {
		setupPipeline(pipeline)
	}, description, []string{})
}

func (b *PipelineBuilder) RegisterAnalysisErr(name string, setupPipeline func(pipeline *pipeline.SamplePipeline) error, description string) {
	b.RegisterAnalysisParamsErr(name, func(pipeline *pipeline.SamplePipeline, _ map[string]string) error {
		return setupPipeline(pipeline)
	}, description, []string{})
}

func (b *PipelineBuilder) RegisterFork(name string, createFork ForkFunc, description string, requiredParams []string, optionalParams ...string) {
	if _, ok := b.fork_registry[name]; ok {
		panic("Fork already registered: " + name)
	}
	params := registeredParameters{requiredParams, optionalParams}
	b.fork_registry[name] = registeredFork{name, createFork, params.makeDescription(description), params}
}

type registeredAnalysis struct {
	Name        string
	Func        AnalysisFunc
	Description string
	Params      registeredParameters
}

type registeredFork struct {
	Name        string
	Func        ForkFunc
	Description string
	Params      registeredParameters
}

type registeredParameters struct {
	required []string
	optional []string
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
