package fork

import (
	"testing"

	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type distributorsTestSuite struct {
	t *testing.T
	*require.Assertions
}

func TestDistributors(t *testing.T) {
	suite.Run(t, new(distributorsTestSuite))
}

func (suite *distributorsTestSuite) T() *testing.T {
	return suite.t
}

func (suite *distributorsTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (suite *distributorsTestSuite) TestTagTemplateDistributor() {
	s := &bitflow.Sample{Values: []bitflow.Value{1, 2, 3}}
	s.SetTag("tag1", "val1")
	s.SetTag("tag2", "val2")
	h := &bitflow.Header{Fields: []string{"a", "b", "c"}}

	pipeA := new(pipeline.SamplePipeline)
	pipeB := new(pipeline.SamplePipeline)

	test := func(input, output string, expectedPipes ...*pipeline.SamplePipeline) {
		var dist TagDistributor
		dist.Template = input
		dist.Pipelines = map[string]func() ([]*pipeline.SamplePipeline, error){
			"a*": func() ([]*pipeline.SamplePipeline, error) {
				return []*pipeline.SamplePipeline{pipeA}, nil
			},
			"b*": func() ([]*pipeline.SamplePipeline, error) {
				return []*pipeline.SamplePipeline{pipeB}, nil
			},
			"c": func() ([]*pipeline.SamplePipeline, error) {
				return []*pipeline.SamplePipeline{pipeA, pipeB}, nil
			},
		}
		suite.NoError(dist.Init())
		res, err := dist.Distribute(s, h)
		suite.Nil(err)
		suite.Len(res, len(expectedPipes))
		for i := 0; i < len(expectedPipes); i++ {
			suite.Equal(output, res[i].Key)
			suite.Equal(expectedPipes[i], res[i].Pipe)
		}
	}
	test("abc", "abc", pipeA)
	test("${missing}", "")
	test("a${missing}b", "ab", pipeA)
	test("a${missing}", "a", pipeA)
	test("${missing}b", "b", pipeB)
	test("b${tag1}", "bval1", pipeB)
	test("a${tag1}b${tag2}c", "aval1bval2c", pipeA)
	test("a$x$z", "a$x$z", pipeA)
	test("b${tag", "b${tag", pipeB)
	test("c", "c", pipeA, pipeB)
	test("cxx", "")
}
