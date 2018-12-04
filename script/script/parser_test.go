package script

import (
	"strings"
	"testing"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/steps"
	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/stretchr/testify/assert"
)

type testOutputCatcher struct {
	calledSteps []string
}

func TestParseScript_withFileInputAndOutput_shouldHaveFileSourceAndFileSink(t *testing.T) {
	testScript := "./in -> noop() -> ./out"
	parser, _ := createTestParser()

	pipe, errs := parser.ParseScript(testScript)

	assert.Len(t, errs, 0)
	_, isFileSource := pipe.Source.(*bitflow.FileSource)
	assert.True(t, isFileSource)
	_, isFileSink := pipe.Processors[1].(*bitflow.FileSink)
	assert.True(t, isFileSink)
}

func TestParseScript_multipleOutputs(t *testing.T) {
	testScript := "./in -> normal_transform() -> error_returning_transform -> ./out"
	parser, out := createTestParser()
	_, errs := parser.ParseScript(testScript)

	assert.Len(t, errs, 0)
	assert.Equal(t, "normal_transform", out.calledSteps[0])
}

func TestParseScript_withEnforcedBatchInStream_shouldReturnError(t *testing.T) {
	testScript := "./in -> batch_enforcing_transform() -> ./out"
	parser, _ := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Len(t, errs, 1)
	assert.True(t, strings.Contains(errs[0].Error(), "Processor used outside window, but does not support stream processing"))
}

// TODO add test
func __TestParseScript_withStreamTransformInWindow_shouldReturnError(t *testing.T) {
	testScript := "./in -> window { normal_transform() -> batch_supporting_transform()} -> ./out"
	parser, _ := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Len(t, errs, 1)
	assert.True(t, strings.Contains(errs[0].Error(), "Processor used in window, but does not support batch processing."))
}

// TODO add test
func __TestParseScript_withWindow_shouldWork(t *testing.T) {
	testScript := "./in -> window { batch_enforcing_transform()-> batch_supporting_transform()} -> normal_transform() -> ./out"
	parser, out := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Equal(t, nil, errs.NilOrError())
	assert.Equal(t, "batch_enforcing_transform", out.calledSteps[0])
	assert.Equal(t, "batch_supporting_transform", out.calledSteps[1])
	assert.Equal(t, "normal_transform", out.calledSteps[2])
}

// TODO add test
func __TestParseScript_withWindowInWindow_shouldReturnError(t *testing.T) {
	testScript := "./in -> window {batch_supporting_transform() -> window { batch_supporting_transform()}} -> normal_transform() -> ./out"
	parser, _ := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Equal(t, 1, len(errs))
	assert.True(t, strings.Contains(errs.Error(), "Window inside Window is not allowed."))
}

func createTestParser() (BitflowScriptParser, *testOutputCatcher) {
	out := &testOutputCatcher{}
	registry := reg.NewProcessorRegistry()
	registry.Endpoints = *bitflow.NewEndpointFactory()
	registry.RegisterAnalysisParamsErr("normal_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
			out.calledSteps = append(out.calledSteps, "normal_transform")
			return nil
		}, "a normal transform")

	registry.RegisterAnalysisParamsErr("error_returning_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
			out.calledSteps = append(out.calledSteps, "error_returning_transform")
			return errors.Errorf("error_returning_transform")
		}, "an error returning")

	registry.RegisterAnalysisParamsErr("required_param_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
			out.calledSteps = append(out.calledSteps, "required_param_transform")
			return nil
		}, "a transform requiring a parameter", reg.RequiredParams("requiredParam"))

	registry.RegisterAnalysisParamsErr("batch_supporting_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
			out.calledSteps = append(out.calledSteps, "batch_supporting_transform")
			return nil
		}, "a batch supporting transform", reg.SupportBatch())

	registry.RegisterAnalysisParamsErr("batch_enforcing_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
			out.calledSteps = append(out.calledSteps, "batch_enforcing_transform")
			return nil
		}, "a batch enforcing transform", reg.EnforceBatch())
	steps.RegisterNoop(registry)
	return BitflowScriptParser{Registry: registry}, out
}
