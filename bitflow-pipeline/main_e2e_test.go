package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testData = `time,val
2006-01-02 15:04:05.999999980,1
2006-01-02 15:04:05.999999981,2
2006-01-02 15:04:05.999999982,3
2006-01-02 15:04:05.999999983,4
2006-01-02 15:04:05.999999984,5
2006-01-02 15:04:05.999999985,6
2006-01-02 15:04:05.999999986,7
2006-01-02 15:04:05.999999987,8
2006-01-02 15:04:05.999999988,9
`
	expectedOutput = `time,tags,val
2006-01-02 15:04:05.999999985,,6
2006-01-02 15:04:05.999999984,,5
2006-01-02 15:04:05.999999982,,3
2006-01-02 15:04:05.999999986,,7
2006-01-02 15:04:05.999999988,,9
2006-01-02 15:04:05.99999998,,1
2006-01-02 15:04:05.999999981,,2
`
	testScript = `-> shuffle() -> rr(){1 -> head(num=2);2 -> head(num=5)}->`
)

func TestE2EWithSampleScripts(t *testing.T) {
	suite.Run(t, new(scriptIntegrationTestSuite))
}

type scriptIntegrationTestSuite struct {
	t *testing.T
	*require.Assertions

	sampleDataFile       *os.File
	sampleScriptFile     *os.File
	sampleOutputFileName string
}

func (suite *scriptIntegrationTestSuite) T() *testing.T {
	return suite.t
}

func (suite *scriptIntegrationTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (suite *scriptIntegrationTestSuite) SetupSuite() {
	var err error
	suite.sampleDataFile, err = ioutil.TempFile(os.TempDir(), "sample-data-")
	suite.NoError(err)
	suite.sampleScriptFile, err = ioutil.TempFile(os.TempDir(), "test-script-")
	suite.NoError(err)
	uid, err := uuid.NewV4()
	suite.NoError(err)
	suite.sampleOutputFileName = filepath.Join(os.TempDir(), "test-output-"+uid.String())
	suite.NoError(err)

	suite.sampleDataFile.WriteString(testData)
	suite.sampleScriptFile.WriteString(suite.sampleDataFile.Name() + testScript + suite.sampleOutputFileName)

	suite.NoError(suite.sampleScriptFile.Close())
	suite.NoError(suite.sampleDataFile.Close())
}

func (suite *scriptIntegrationTestSuite) TearDownSuite() {
	suite.NoError(os.Remove(suite.sampleDataFile.Name()))
	suite.NoError(os.Remove(suite.sampleScriptFile.Name()))
	suite.NoError(os.Remove(suite.sampleOutputFileName))
}

func (suite *scriptIntegrationTestSuite) TestScriptExecutionWithFirstImplementation() {
	resultCode := executeMain([]string{"bitflow-pipeline", "-f", suite.sampleScriptFile.Name()})
	suite.Equal(resultCode, 0)
	content, err := ioutil.ReadFile(suite.sampleOutputFileName)
	suite.NoError(err)
	suite.Equal(expectedOutput, string(content))
}

func (suite *scriptIntegrationTestSuite) TestScriptExecutionWithNewImplementation() {
	resultCode := executeMain([]string{"bitflow-pipeline", "-new", "-f", suite.sampleScriptFile.Name()})
	suite.Equal(resultCode, 0)
	content, err := ioutil.ReadFile(suite.sampleOutputFileName)
	suite.NoError(err)
	suite.Equal(expectedOutput, string(content))
}

func executeMain(args []string) int {
	os.Args = args
	c := do_main()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	return c
}
