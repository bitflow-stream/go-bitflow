package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AnalysisTestSuite struct {
	suite.Suite
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(AnalysisTestSuite))
}

func (suite *AnalysisTestSuite) SetupTest() {

}

func (suite *AnalysisTestSuite) resolve(args ...string) ([]string, error) {
	analyses, err := resolvePipeline(args)
	if err != nil {
		return nil, err
	}
	var pipeline SamplePipeline
	pipeline.setup(analyses)
	return pipeline.print(), nil
}

func (suite *AnalysisTestSuite) TestMissingAnalysis() {
	_, err := suite.resolve("xxXX-missing-XXxx")
	suite.Error(err)
	suite.True(strings.Contains(err.Error(), "xxXX-missing-XXxx"))
}
