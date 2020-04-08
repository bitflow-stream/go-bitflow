package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/antongulenko/golib"
	"github.com/stretchr/testify/suite"
)

const scriptContent = "a -> avg() -> b"
const nonexistingFile = "-------SOMETHING-NONEXISTING-------"

type argsTestSuite struct {
	golib.AbstractTestSuite

	tempDir        string
	executableFile string
	normalFile     string
	wrongNameFile  string
}

func TestProcessorRegistry(t *testing.T) {
	suite.Run(t, new(argsTestSuite))
}

func (suite *argsTestSuite) SetupSuite() {
	name, err := ioutil.TempDir("", "bitflow-pipeline-test-main-script")
	suite.tempDir = name
	suite.NoError(err)
	normalFileName := filepath.Join(name, "normal.bf")
	executableFileName := filepath.Join(name, "executable.bf")
	wrongNameFileName := filepath.Join(name, "wrong-name.csv")
	suite.normalFile = normalFileName
	suite.executableFile = executableFileName
	suite.wrongNameFile = wrongNameFileName

	normalFile, err := os.Create(normalFileName)
	suite.NoError(err)
	suite.NoError(normalFile.Chmod(0666))
	_, err = normalFile.WriteString(scriptContent)
	suite.NoError(err)
	suite.NoError(normalFile.Close())

	executableFile, err := os.Create(executableFileName)
	suite.NoError(err)
	suite.NoError(executableFile.Chmod(0777))
	_, err = executableFile.WriteString(scriptContent)
	suite.NoError(err)
	suite.NoError(executableFile.Close())

	wrongNameFile, err := os.Create(wrongNameFileName)
	suite.NoError(err)
	suite.NoError(wrongNameFile.Chmod(0666))
	_, err = wrongNameFile.WriteString(scriptContent)
	suite.NoError(err)
	suite.NoError(wrongNameFile.Close())
}

func (suite *argsTestSuite) TearDownSuite() {
	suite.NoError(os.Remove(suite.executableFile))
	suite.NoError(os.Remove(suite.normalFile))
	suite.NoError(os.Remove(suite.wrongNameFile))
	suite.NoError(os.RemoveAll(suite.tempDir))
}

func (suite *argsTestSuite) TestFixArgs() {
	run := func(in, out []string) {
		in = append([]string{"bitflow-pipeline"}, in...)
		out = append([]string{"bitflow-pipeline"}, out...)
		fix_arguments(&in)
		suite.Equal(out, in)
	}

	// Typical #!/usr/bin/env invocations
	run([]string{suite.executableFile}, []string{"-f", suite.executableFile})
	run([]string{suite.executableFile, "arg1"}, []string{"-f", suite.executableFile, "arg1"})
	run([]string{suite.executableFile, "arg1", "arg2"}, []string{"-f", suite.executableFile, "arg1", "arg2"})

	// Typical #!/absolute/path/bitflow-pipeline invocations
	run([]string{"argA argB 'x y'", suite.executableFile}, []string{"argA", "argB", "x y", "-f", suite.executableFile})
	run([]string{"argA argB 'x y'", suite.executableFile, "arg1"}, []string{"argA", "argB", "x y", "-f", suite.executableFile, "arg1"})
	run([]string{"argA argB 'x y'", suite.executableFile, "arg1", "arg2"}, []string{"argA", "argB", "x y", "-f", suite.executableFile, "arg1", "arg2"})
	run([]string{"argA argB 'x y' -f", suite.executableFile, "arg1", "arg2"}, []string{"argA", "argB", "x y", "-f", suite.executableFile, "arg1", "arg2"})

	// Other invocations
	run([]string{}, []string{})
	run([]string{nonexistingFile}, []string{nonexistingFile})
	run([]string{"argA argB 'x y'", nonexistingFile}, []string{"argA argB 'x y'", nonexistingFile})
	run([]string{"a", "b", suite.executableFile}, []string{"a", "b", suite.executableFile})
	run([]string{"a", "b", suite.executableFile, "x"}, []string{"a", "b", suite.executableFile, "x"})
	run([]string{suite.normalFile, "arg1"}, []string{suite.normalFile, "arg1"})
	run([]string{"argA argB 'x y'", suite.normalFile, "arg1"}, []string{"argA argB 'x y'", suite.normalFile, "arg1"})
	run([]string{"argA argB 'x y'", "-f", suite.executableFile, "arg1", "arg2"}, []string{"argA argB 'x y'", "-f", suite.executableFile, "arg1", "arg2"})
	run([]string{"-f", suite.executableFile, "arg1", "arg2"}, []string{"-f", suite.executableFile, "arg1", "arg2"})
}

func (suite *argsTestSuite) TestGetScript() {
	run := func(args []string, scriptFile string, result string, expectedErr string) {
		script, err := get_script(args, scriptFile)
		if expectedErr == "" {
			suite.NoError(err)
		} else {
			suite.Error(err)
			suite.Contains(err.Error(), expectedErr)
		}
		suite.Equal(result, script)
	}

	// Errors
	run([]string{}, "", "", "Please provide a bitflow pipeline script via -f or directly as parameter.")
	run([]string{"   ", "\t  \t"}, "", "", "Please provide a bitflow pipeline script via -f or directly as parameter.")
	run([]string{"x"}, "x", "", "Please provide a bitflow pipeline script either via -f or as parameter, not both.")
	run([]string{}, nonexistingFile, "", "Error reading bitflow script file "+nonexistingFile)

	// Normal execution
	run([]string{scriptContent}, "", scriptContent, "")
	run([]string{scriptContent, scriptContent}, "", scriptContent+" "+scriptContent, "")
	run([]string{suite.wrongNameFile}, "", suite.wrongNameFile, "")
	run([]string{nonexistingFile}, "", nonexistingFile, "")
	run([]string{nonexistingFile + ".bf"}, "", nonexistingFile+".bf", "")
	run([]string{}, suite.executableFile, scriptContent, "")
	run([]string{}, suite.normalFile, scriptContent, "")
	run([]string{suite.normalFile}, "", scriptContent, "")
	run([]string{suite.executableFile}, "", scriptContent, "")
}
