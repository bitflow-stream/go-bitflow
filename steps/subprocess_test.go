package steps

import (
	"testing"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/stretchr/testify/suite"
)

type SubProcessTestSuite struct {
	golib.AbstractTestSuite
}

func TestSubProcess(t *testing.T) {
	suite.Run(t, new(SubProcessTestSuite))
}

func (s *SubProcessTestSuite) TestExternalExecutableRegistration() {
	test := func(description string, registrationError string, shortName string, executable string, params map[string]interface{}, args ...string) {
		s.SubTest(description, func() {
			endpoints := bitflow.NewEndpointFactory()
			registry := reg.NewProcessorRegistry(endpoints)

			err := RegisterExecutable(registry, description)
			if registrationError == "" {
				s.NoError(err)
			} else {
				s.Error(err)
				s.Contains(err.Error(), registrationError)
			}

			step := registry.GetStep(shortName)
			if registrationError != "" {
				s.Nil(step)
				return
			}

			s.Equal(step.Name, shortName)
			s.Contains(step.Description, executable)

			pipeline := new(bitflow.SamplePipeline)
			s.NoError(step.Params.ValidateAndSetDefaults(params))
			s.NoError(step.Func(pipeline, params))

			s.Len(pipeline.Processors, 1)
			runningStep := pipeline.Processors[0]
			s.NotNil(runningStep)
			s.IsType(new(SubProcessRunner), runningStep)
			s.Equal(executable, runningStep.(*SubProcessRunner).Cmd)
			s.Equal(args, runningStep.(*SubProcessRunner).Args)
		})
	}

	// Wrong registration format
	test("", "Wrong format for external executable (have 1 part(s))", "", "", nil)
	test("a", "Wrong format for external executable (have 1 part(s))", "", "", nil)
	test("a;b", "Wrong format for external executable (have 2 part(s))", "", "", nil)
	test("a;b;c;d", "Wrong format for external executable (have 4 part(s))", "", "", nil)

	// Correct step creation

	// No initial args
	test("name;exe;", "", "name", "exe",
		map[string]interface{}{"step": "anything",},
		"-step", "anything", "-args", "")

	// One initial arg
	test("name;exe;args", "", "name", "exe",
		map[string]interface{}{"step": "anything",},
		"args", "-step", "anything", "-args", "")

	// Multiple initial args
	test("name;exe;arg1 arg2 arg3", "", "name", "exe",
		map[string]interface{}{"step": "anything"},
		"arg1", "arg2", "arg3", "-step", "anything", "-args", "")

	// No initial args with step args and exe-args
	test("name;exe;", "", "name", "exe",
		map[string]interface{}{"step": "anything", "args": map[string]string{"a": "b", "c": "d"}, "exe-args": []string{"extra", "arg"}},
		"extra", "arg", "-step", "anything", "-args", "a=b, c=d")

	// Multiple initial args with step args and exe-args
	test("name;exe;arg1 arg2 arg3", "", "name", "exe",
		map[string]interface{}{"step": "anything", "args": map[string]string{"c": "d", "a": "b"}, "exe-args": []string{"extra", "arg"}},
		"arg1", "arg2", "arg3", "extra", "arg", "-step", "anything", "-args", "a=b, c=d")
}
