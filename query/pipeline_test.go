package query

/*

type pipeTestSuite struct {
	t *testing.T
	*require.Assertions
}

func TestPipelineGeneration(t *testing.T) {
	suite.Run(t, new(pipeTestSuite))
}

func (suite *pipeTestSuite) T() *testing.T {
	return suite.t
}

func (suite *pipeTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (suite *pipeTestSuite) test(script string, expected *bitflow.SamplePipeline) {
	ast, err := NewParser(bytes.NewReader([]byte(script))).Parse()
	suite.NoError(err)

	var b PipelineBuilder
	pipe, err := b.MakePipeline(ast)
	suite.NoError(err)
	suite.EqualValues(&SamplePipeline{SamplePipeline: *expected}, pipe)
}

func (suite *pipeTestSuite) TestRegularPipeline() {
	suite.test("in -> out", &bitflow.SamplePipeline{
		Source: &bitflow.FileSource{
			Filenames: []string{"in"},
		},
		Sink: &bitflow.FileSink{
			Filename: "out",
		},
	})

	suite.test("a b c -> out", &bitflow.SamplePipeline{
		Source: &bitflow.FileSource{},
	})
}

*/
