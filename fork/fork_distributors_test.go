package fork

import (
	"testing"

	bitflow "github.com/antongulenko/go-bitflow"
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

	test := func(input, output string) {
		res, err := (&TagDistributor{TagTemplate: TagTemplate{Template: input}}).Distribute(s, h)
		suite.Nil(err)
		suite.Len(res, 1)
		suite.Equal(output, res[0].Key)
	}
	test("abc", "abc")
	test("${missing}", "")
	test("a${missing}b", "ab")
	test("a${missing}", "a")
	test("${missing}b", "b")
	test("${tag1}", "val1")
	test("a${tag1}b${tag2}c", "aval1bval2c")
	test("a$x$z", "a$x$z")
	test("${tag", "${tag")
}
