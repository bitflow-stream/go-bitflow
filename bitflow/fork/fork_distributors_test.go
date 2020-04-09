package fork

import (
	"testing"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/stretchr/testify/suite"
)

type ForkDistributorsTestSuite struct {
	golib.AbstractTestSuite
}

func TestForkDistributors(t *testing.T) {
	suite.Run(t, new(ForkDistributorsTestSuite))
}

func (suite *ForkDistributorsTestSuite) TestTagTemplateDistributor() {
	s := &bitflow.Sample{Values: []bitflow.Value{1, 2, 3}}
	s.SetTag("tag1", "val1")
	s.SetTag("tag2", "val2")
	h := &bitflow.Header{Fields: []string{"a", "b", "c"}}

	pipeA := new(bitflow.SamplePipeline)
	pipeB := new(bitflow.SamplePipeline)

	test := func(input, output string, expectedPipes ...*bitflow.SamplePipeline) {
		var dist TagDistributor
		dist.Template = input
		dist.Pipelines = map[string]func() ([]*bitflow.SamplePipeline, error){
			"a*": func() ([]*bitflow.SamplePipeline, error) {
				return []*bitflow.SamplePipeline{pipeA}, nil
			},
			"b*": func() ([]*bitflow.SamplePipeline, error) {
				return []*bitflow.SamplePipeline{pipeB}, nil
			},
			"c": func() ([]*bitflow.SamplePipeline, error) {
				return []*bitflow.SamplePipeline{pipeA, pipeB}, nil
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
