package script

import (
	"testing"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/steps"
	"github.com/bugsnag/bugsnag-go/errors"
)

type ParserTestSuite struct {
	golib.AbstractTestSuite
}

func TestParser(t *testing.T) {
	new(ParserTestSuite).Run(t)
}

func (s *ParserTestSuite) Test_withFileInputAndOutput_shouldHaveFileSourceAndFileSink() {
	testScript := "./in -> noop() -> ./out"
	parser, _ := s.createTestParser()

	pipe, errs := parser.ParseScript(testScript)

	s.Len(errs, 0)
	_, isFileSource := pipe.Source.(*bitflow.FileSource)
	s.True(isFileSource)
	_, isFileSink := pipe.Processors[1].(*bitflow.FileSink)
	s.True(isFileSink)
}

func (s *ParserTestSuite) Test_multipleOutputs() {
	testScript := "./in -> normal_transform() -> error_returning_transform -> ./out"
	parser, out := s.createTestParser()
	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 0)
	s.Equal("normal_transform", out.calledSteps[0])
}

func (s *ParserTestSuite) Test_withEnforcedBatchInStream_shouldReturnError() {
	testScript := "./in -> batch_transform() -> ./out"
	parser, _ := s.createTestParser()

	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 1)
	s.Contains(errs[0].Error(), "Processing step 'batch_transform' is unknown, but a batch step with that name exists")
}

func (s *ParserTestSuite) Test_withEnforcedBatchInStream_shouldReturnError_unknownStep() {
	testScript := "./in -> xxxYYY() -> ./out"
	parser, _ := s.createTestParser()

	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 1)
	s.Contains(errs[0].Error(), "Processing step 'xxxYYY' is unknown")
	s.NotContains(errs[0].Error(), ", but a ") // no additional help, because the step is not known at all
}

func (s *ParserTestSuite) Test_withStreamTransformInWindow_shouldReturnError() {
	testScript := "./in -> batch() { normal_transform() -> batch_transform()} -> ./out"
	parser, _ := s.createTestParser()

	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 1)
	s.Contains(errs[0].Error(), "Batch step 'normal_transform' is unknown, but a non-batch step with that name exists")
}

func (s *ParserTestSuite) Test_withStreamTransformInWindow_shouldReturnError_unknownStep() {
	testScript := "./in -> batch() { xxxxYYYY() -> batch_transform()} -> ./out"
	parser, _ := s.createTestParser()

	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 1)
	s.Contains(errs[0].Error(), "Batch step 'xxxxYYYY' is unknown")
	s.NotContains(errs[0].Error(), ", but a ") // no additional help, because the step is not known at all
}

func (s *ParserTestSuite) Test_missingRequiredParameter() {
	testScript := "./in -> required_param_transform() -> ./out"
	parser, _ := s.createTestParser()

	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 1)
	s.Contains(errs[0].Error(), "Missing required parameter 'requiredParam' (type string)")
}

func (s *ParserTestSuite) Test_unknownParameter() {
	testScript := "./in -> required_param_transform(HELLO=1) -> ./out"
	parser, _ := s.createTestParser()

	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 1)
	s.Contains(errs[0].Error(), "Unknown parameter 'HELLO'")
}

func (s *ParserTestSuite) Test_doubleParameter() {
	testScript := "./in -> required_param_transform(requiredParam=1, requiredParam=2) -> ./out"
	parser, _ := s.createTestParser()

	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 1)
	s.Contains(errs[0].Error(), "Parameter 'requiredParam' is redefined")
}

func (s *ParserTestSuite) Test_withWindow_shouldWork() {
	testScript := "./in -> batch() { batch_transform()-> batch_transform()} -> normal_transform() -> ./out"
	parser, out := s.createTestParser()

	_, errs := parser.ParseScript(testScript)
	s.NoError(errs.NilOrError())
	s.Equal("batch_transform", out.calledSteps[0])
	s.Equal("batch_transform", out.calledSteps[1])
	s.Equal("normal_transform", out.calledSteps[2])
}

func (s *ParserTestSuite) Test_withWindowInWindow_shouldReturnError() {
	testScript := "./in -> batch() {batch_transform() -> batch() { batch_transform()}} -> normal_transform() -> ./out"
	parser, _ := s.createTestParser()

	_, errs := parser.ParseScript(testScript)

	s.Len(errs, 3)
	s.Contains(errs.Error(), "mismatched input 'batch'")
}

func (s *ParserTestSuite) Test_listAndMapParams() {
	var (
		normal1       bool
		normal2       int
		optional1     float64
		optional2     string
		list1         []time.Duration
		list2         []int
		list3         []time.Time
		emptyList     []float64
		map1          map[string]string
		map2          map[string]int
		emptyMap      map[string]float64
		optionalList1 []int
		optionalList2 []int
		optionalMap1  map[string]int
		optionalMap2  map[string]int
	)

	parser, _ := s.createTestParser()
	parser.Registry.RegisterStep("special_params",
		func(pipeline *bitflow.SamplePipeline, params map[string]interface{}) error {
			list1 = params["list1"].([]time.Duration)
			list2 = params["list2"].([]int)
			list3 = params["list3"].([]time.Time)
			map1 = params["map1"].(map[string]string)
			map2 = params["map2"].(map[string]int)
			emptyMap = params["emptyMap"].(map[string]float64)
			emptyList = params["emptyList"].([]float64)
			optionalList1 = params["optionalList1"].([]int)
			optionalList2 = params["optionalList2"].([]int)
			optionalMap1 = params["optionalMap1"].(map[string]int)
			optionalMap2 = params["optionalMap2"].(map[string]int)
			normal1 = params["normal1"].(bool)
			normal2 = params["normal2"].(int)
			optional1 = params["optional1"].(float64)
			optional2 = params["optional2"].(string)
			return nil
		}, "step with list and map parameters").
		Required("list1", reg.List(reg.Duration())).
		Required("list2", reg.List(reg.Int())).
		Required("list3", reg.List(reg.Time())).
		Required("map1", reg.Map(reg.String())).
		Required("map2", reg.Map(reg.Int())).
		Required("emptyMap", reg.Map(reg.Float())).
		Required("emptyList", reg.List(reg.Float())).
		Optional("optionalList1", reg.List(reg.Int()), []int{10, 11}).
		Optional("optionalList2", reg.List(reg.Int()), []int{12, 13}).
		Optional("optionalMap1", reg.Map(reg.Int()), map[string]int{"a": 10, "b": 11}).
		Optional("optionalMap2", reg.Map(reg.Int()), map[string]int{"c": 12, "d": 12}).
		Required("normal1", reg.Bool()).
		Required("normal2", reg.Int()).
		Optional("optional1", reg.Float(), 10.222).
		Optional("optional2", reg.String(), "defaultVal2")

	testScript := "special_params(list2=[ 1,'2',3], map1 = { x=y, 1=2, ' '=v } , list1 =    [1s   ,2h,    3m]  " +
		", map2 = { 4=5 }, emptyList=[], emptyMap={}, optionalList2= [50,60],  optionalMap1={ g=40, h=60 }," +
		"normal1= 'true', 'normal2'=  33, optional1=  3.444,   list3=['2100-10-10 10:10:10.123456', '1990-05-06 07:15:06.1'])"
	_, errs := parser.ParseScript(testScript)
	s.NoError(errs.NilOrError())

	s.Equal([]time.Duration{1 * time.Second, 2 * time.Hour, 3 * time.Minute}, list1)
	s.Equal([]int{1, 2, 3}, list2)
	s.Equal([]float64{}, emptyList)
	s.Equal(map[string]string{"x": "y", "1": "2", " ": "v"}, map1)
	s.Equal(map[string]int{"4": 5}, map2)
	s.Equal(map[string]float64{}, emptyMap)
	s.Equal([]int{10, 11}, optionalList1) // default value
	s.Equal([]int{50, 60}, optionalList2)
	s.Equal(map[string]int{"g": 40, "h": 60}, optionalMap1)
	s.Equal(map[string]int{"c": 12, "d": 12}, optionalMap2) // default value
	s.Equal(true, normal1)
	s.Equal(33, normal2)
	s.Equal(3.444, optional1)
	s.Equal("defaultVal2", optional2)

	s.Len(list3, 2)
	format := "2006-01-02 15:04:05.999999"
	s.Equal(time.Date(2100, time.October, 10, 10, 10, 10, 123456*1000, time.Local).Format(format), list3[0].Format(format))
	s.Equal(time.Date(1990, time.May, 6, 7, 15, 06, 100*1000*1000, time.Local).Format(format), list3[1].Format(format))
}

type testOutputCatcher struct {
	calledSteps []string
}

func (s *ParserTestSuite) createTestParser() (BitflowScriptParser, *testOutputCatcher) {
	out := &testOutputCatcher{}
	registry := reg.NewProcessorRegistry(bitflow.NewEndpointFactory())
	registry.RegisterStep("normal_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]interface{}) error {
			out.calledSteps = append(out.calledSteps, "normal_transform")
			return nil
		}, "a normal transform")

	registry.RegisterStep("error_returning_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]interface{}) error {
			out.calledSteps = append(out.calledSteps, "error_returning_transform")
			return errors.Errorf("error_returning_transform")
		}, "an error returning")

	registry.RegisterStep("required_param_transform",
		func(pipeline *bitflow.SamplePipeline, params map[string]interface{}) error {
			out.calledSteps = append(out.calledSteps, "required_param_transform")
			return nil
		}, "a transform requiring a parameter").
		Required("requiredParam", reg.String())

	registry.RegisterBatchStep("batch_transform",
		func(params map[string]interface{}) (res bitflow.BatchProcessingStep, err error) {
			out.calledSteps = append(out.calledSteps, "batch_transform")
			return nil, nil
		}, "a batch enforcing transform")
	steps.RegisterNoop(registry)
	return BitflowScriptParser{Registry: registry}, out
}
