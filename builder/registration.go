package builder

import (
	"fmt"
	"time"
	"strconv"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/fork"
)

type Subpipeline interface {
	Keys() []string
	Build() (*pipeline.SamplePipeline, error)
}

type AnalysisFunc func(pipeline *pipeline.SamplePipeline, params map[string]string) error

type ForkFunc func(subpiplines []Subpipeline, params map[string]string) (fork.Distributor, error)

// Alias for reducing name changes in code
type PipelineBuilder ProcessorRegistry

type AnalysisRegistration struct {
	AnalysisFunc   AnalysisFunc
	Name           string
	Description    string
	RequiredParams []string
	OptionalParams []string
	ProcessesBatch bool
}

type Options struct {
	RequiredParams []string
	OptionalParams []string
	// SupportBatchProcessing, if true this Processor can be called with batches
	SupportBatchProcessing bool
	// EnforceBatchProcessing, if true this Processor can ONLY be called with batches
	EnforceBatchProcessing bool
}

type Option func(*Options)

func RequiredParams(params ...string) Option {
	return func(opts *Options) {
		opts.RequiredParams = params
	}
}

func OptionalParams(params ...string) Option {
	return func(opts *Options) {
		opts.OptionalParams = params
	}
}

func SupportBatch() Option {
	return func(opts *Options) {
		opts.SupportBatchProcessing = true
	}
}

func EnforceBatch() Option {
	return func(opts *Options) {
		opts.SupportBatchProcessing = true
		opts.EnforceBatchProcessing = true
	}
}

func GetOpts(options []Option) Options {
	opts := Options{}
	for _, o := range options {
		o(&opts)
	}
	return opts
}

type ProcessorRegistry interface {
	RegisterAnalysisParamsErr(name string, setupPipeline AnalysisFunc, description string, options ...Option)

	RegisterAnalysisParams(name string, setupPipeline func(pipeline *pipeline.SamplePipeline, params map[string]string), description string, options ...Option)

	RegisterAnalysis(name string, setupPipeline func(pipeline *pipeline.SamplePipeline), description string, options ...Option)

	RegisterAnalysisErr(name string, setupPipeline func(pipeline *pipeline.SamplePipeline) error, description string, options ...Option)

	RegisterFork(name string, createFork ForkFunc, description string, options ...Option)
}

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

// Proxy to register steps in multiple registries at once.
type MultiRegistryBuilder struct {
	Registries []ProcessorRegistry
}

func (m MultiRegistryBuilder) RegisterAnalysisParamsErr(name string, setupPipeline AnalysisFunc, description string, options ...Option) {
	for _, r := range m.Registries {
		r.RegisterAnalysisParamsErr(name, setupPipeline, description, options...)
	}
}

func (m MultiRegistryBuilder) RegisterAnalysisParams(name string, setupPipeline func(pipeline *pipeline.SamplePipeline, params map[string]string), description string, options ...Option) {
	for _, r := range m.Registries {
		r.RegisterAnalysisParams(name, setupPipeline, description, options...)
	}
}

func (m MultiRegistryBuilder) RegisterAnalysis(name string, setupPipeline func(pipeline *pipeline.SamplePipeline), description string, options ...Option) {
	for _, r := range m.Registries {
		r.RegisterAnalysis(name, setupPipeline, description, options...)
	}
}

func (m MultiRegistryBuilder) RegisterAnalysisErr(name string, setupPipeline func(pipeline *pipeline.SamplePipeline) error, description string, options ...Option) {
	for _, r := range m.Registries {
		r.RegisterAnalysisErr(name, setupPipeline, description, options...)
	}
}

func (m MultiRegistryBuilder) RegisterFork(name string, createFork ForkFunc, description string, options ...Option) {
	for _, r := range m.Registries {
		r.RegisterFork(name, createFork, description, options...)
	}
}
