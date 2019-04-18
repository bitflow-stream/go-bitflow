package script

import (
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
	testScript := "./in -> batch_transform() -> ./out"
	parser, _ := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "Processing step 'batch_transform' is unknown, but a batch step with that name exists")
}

func TestParseScript_withEnforcedBatchInStream_shouldReturnError_unknownStep(t *testing.T) {
	testScript := "./in -> xxxYYY() -> ./out"
	parser, _ := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "Processing step 'xxxYYY' is unknown")
	assert.NotContains(t, errs[0].Error(), ", but a ") // no additional help, because the step is not known at all
}

func TestParseScript_withStreamTransformInWindow_shouldReturnError(t *testing.T) {
	testScript := "./in -> batch() { normal_transform() -> batch_transform()} -> ./out"
	parser, _ := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "Batch step 'normal_transform' is unknown, but a non-batch step with that name exists")
}

func TestParseScript_withStreamTransformInWindow_shouldReturnError_unknownStep(t *testing.T) {
	testScript := "./in -> batch() { xxxxYYYY() -> batch_transform()} -> ./out"
	parser, _ := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "Batch step 'xxxxYYYY' is unknown")
	assert.NotContains(t, errs[0].Error(), ", but a ") // no additional help, because the step is not known at all
}

func TestParseScript_withWindow_shouldWork(t *testing.T) {
	testScript := "./in -> batch() { batch_transform()-> batch_transform()} -> normal_transform() -> ./out"
	parser, out := createTestParser()

	_, errs := parser.ParseScript(testScript)
	assert.NoError(t, errs.NilOrError())
	assert.Equal(t, "batch_transform", out.calledSteps[0])
	assert.Equal(t, "batch_transform", out.calledSteps[1])
	assert.Equal(t, "normal_transform", out.calledSteps[2])
}

func TestParseScript_withWindowInWindow_shouldReturnError(t *testing.T) {
	testScript := "./in -> batch() {batch_transform() -> batch() { batch_transform()}} -> normal_transform() -> ./out"
	parser, _ := createTestParser()

	_, errs := parser.ParseScript(testScript)

	assert.Equal(t, 4, len(errs))
	assert.Contains(t, errs.Error(), "mismatched input 'batch'")
}

func createTestParser() (BitflowScriptParser, *testOutputCatcher) {
	out := &testOutputCatcher{}
	registry := reg.NewProcessorRegistry()
	registry.Endpoints = *bitflow.NewEndpointFactory()
	registry.RegisterStep("normal_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
			out.calledSteps = append(out.calledSteps, "normal_transform")
			return nil
		}, "a normal transform")

	registry.RegisterStep("error_returning_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
			out.calledSteps = append(out.calledSteps, "error_returning_transform")
			return errors.Errorf("error_returning_transform")
		}, "an error returning")

	registry.RegisterStep("required_param_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]string) error {
			out.calledSteps = append(out.calledSteps, "required_param_transform")
			return nil
		}, "a transform requiring a parameter", reg.RequiredParams("requiredParam"))
	registry.RegisterBatchStep("batch_transform",
		func(params map[string]string) (res bitflow.BatchProcessingStep, err error) {
			out.calledSteps = append(out.calledSteps, "batch_transform")
			return nil, nil
		}, "a batch enforcing transform")
	steps.RegisterNoop(registry)
	return BitflowScriptParser{Registry: registry}, out
}
