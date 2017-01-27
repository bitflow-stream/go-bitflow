package query

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type parserTestSuite struct {
	t *testing.T
	*require.Assertions
}

func TestParser(t *testing.T) {
	suite.Run(t, new(parserTestSuite))
}

func (suite *parserTestSuite) T() *testing.T {
	return suite.t
}

func (suite *parserTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (suite *parserTestSuite) testErr(code string, expectedRes interface{}, expectedErr error) {
	p := NewParser(bytes.NewBufferString(code))
	res, err := p.Parse()
	if expectedErr != nil {
		suite.Error(err)
		suite.Equal(expectedErr, err)
	} else {
		suite.NoError(err)
	}
	suite.Equal(expectedRes, res)
}

func (suite *parserTestSuite) test(code string, expectedRes interface{}) {
	suite.testErr(code, expectedRes, nil)
}

func (suite *parserTestSuite) TestErrorString() {
	suite.Equal("msg (at [2-4, ->] '->')", ParserError{
		Message: "msg",
		Pos:     Token{Start: 2, End: 4, Lit: "->", Type: NEXT},
	}.Error())
}

func (suite *parserTestSuite) TestLexerError() {
	suite.testErr("-X", Pipeline(nil), ParserError{
		Pos:     Token{Type: NEXT, Start: 0, End: 2, Lit: "-X"},
		Message: "Expected '->'",
	})
}

func (suite *parserTestSuite) TestExpectedStep() {
	suite.testErr("   ", Pipeline(nil), ParserError{
		Pos:     Token{Type: EOF, Start: 3, End: 3, Lit: string(eof)},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("a;", Pipeline(nil), ParserError{
		Pos:     Token{Type: EOF, Start: 2, End: 2, Lit: string(eof)},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr(";", Pipeline(nil), ParserError{
		Pos:     Token{Type: SEP, Start: 0, End: 1, Lit: ";"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("a->;", Pipeline(nil), ParserError{
		Pos:     Token{Type: SEP, Start: 3, End: 4, Lit: ";"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("x->{ ; }", Pipeline(nil), ParserError{
		Pos:     Token{Type: SEP, Start: 5, End: 6, Lit: ";"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("x->fork(){}", Pipeline(nil), ParserError{
		Pos:     Token{Type: CLOSE, Start: 10, End: 11, Lit: "}"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("x->fork(){;}", Pipeline(nil), ParserError{
		Pos:     Token{Type: SEP, Start: 10, End: 11, Lit: ";"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("x->fork(){->xx}->out", Pipeline(nil), ParserError{
		Pos:     Token{Type: NEXT, Start: 10, End: 12, Lit: "->"},
		Message: ExpectedPipelineStepError,
	})
}

func (suite *parserTestSuite) TestExpectedEOF() {
	suite.testErr("x(a=b)(b=c)", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_OPEN, Start: 6, End: 7, Lit: "("},
		Message: "Expected 'EOF'",
	})
	suite.testErr("a->x{a()}", Pipeline(nil), ParserError{
		Pos:     Token{Type: OPEN, Start: 4, End: 5, Lit: "{"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("x(a=b)aa", Pipeline(nil), ParserError{
		Pos:     Token{Type: STR, Start: 6, End: 8, Lit: "aa"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("x(a=b)}", Pipeline(nil), ParserError{
		Pos:     Token{Type: CLOSE, Start: 6, End: 7, Lit: "}"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("x(a=b){e->e()}{", Pipeline(nil), ParserError{
		Pos:     Token{Type: OPEN, Start: 14, End: 15, Lit: "{"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("x(a=b){e->e()}}", Pipeline(nil), ParserError{
		Pos:     Token{Type: CLOSE, Start: 14, End: 15, Lit: "}"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("{a->b}()", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_OPEN, Start: 6, End: 7, Lit: "("},
		Message: "Expected 'EOF'",
	})
}

func (suite *parserTestSuite) TestExpectedClosingBracket() {
	suite.testErr("{x", Pipeline(nil), ParserError{
		Pos:     Token{Type: EOF, Start: 2, End: 2, Lit: string(eof)},
		Message: "Expected '}'",
	})
	suite.testErr("a->{ x() (d)", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_OPEN, Start: 9, End: 10, Lit: "("},
		Message: "Expected '}'",
	})
	suite.testErr("a->{ x() aa", Pipeline(nil), ParserError{
		Pos:     Token{Type: STR, Start: 9, End: 11, Lit: "aa"},
		Message: "Expected '}'",
	})
	suite.testErr("a->{ x(){a->v()} {", Pipeline(nil), ParserError{
		Pos:     Token{Type: OPEN, Start: 17, End: 18, Lit: "{"},
		Message: "Expected '}'",
	})
}

func (suite *parserTestSuite) TestValidatePipeline() {
	suite.testErr("a->b->c", Pipeline(nil), ParserError{
		Pos:     Token{Type: STR, Start: 3, End: 4, Lit: "b"},
		Message: "Intermediate pipeline step cannot be an input or output identifier",
	})
	suite.testErr("a->fork(){x()}", Pipeline(nil), ParserError{
		Pos:     Token{Type: STR, Start: 10, End: 11, Lit: "x"},
		Message: "Forked pipeline must start with a pipeline identifier (string)",
	})
	suite.testErr("a->{a->b()}", Pipeline(nil), ParserError{
		Pos:     Token{Type: STR, Start: 4, End: 5, Lit: "a"},
		Message: "Multiplexed pipeline cannot start with an identifier (string)",
	})
	suite.testErr("fork(){a}", Pipeline(nil), ParserError{
		Pos:     Token{Type: STR, Start: 7, End: 8, Lit: "a"},
		Message: "Forked pipeline must have at least one pipeline step",
	})
	suite.testErr("a->b c", Pipeline(nil), ParserError{
		Pos:     Token{Type: STR, Start: 3, End: 4, Lit: "b"},
		Message: "Multiple sequential outputs are not allowed",
	})
}

func (suite *parserTestSuite) TestParamErrors() {
	suite.testErr("a(((", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_OPEN, Start: 2, End: 3, Lit: "("},
		Message: "Expected 'parameter name (string)'",
	})
	suite.testErr("a(a=b,,)", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_SEP, Start: 6, End: 7, Lit: ","},
		Message: "Expected 'parameter name (string)'",
	})

	suite.testErr("a(a=,)", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_SEP, Start: 4, End: 5, Lit: ","},
		Message: "Expected 'parameter value (string)'",
	})
	suite.testErr("a(a=b,x=)", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_CLOSE, Start: 8, End: 9, Lit: ")"},
		Message: "Expected 'parameter value (string)'",
	})

	suite.testErr("a('a',)", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_SEP, Start: 5, End: 6, Lit: ","},
		Message: "Expected '='",
	})
	suite.testErr("a(a=b,'x')", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_CLOSE, Start: 9, End: 10, Lit: ")"},
		Message: "Expected '='",
	})

	suite.testErr("a(x=f{", Pipeline(nil), ParserError{
		Pos:     Token{Type: OPEN, Start: 5, End: 6, Lit: "{"},
		Message: "Expected ',' or ')'",
	})
	suite.testErr("a(x=f,a=b=", Pipeline(nil), ParserError{
		Pos:     Token{Type: PARAM_EQ, Start: 9, End: 10, Lit: "="},
		Message: "Expected ',' or ')'",
	})
}

func (suite *parserTestSuite) TestExamples() {
	suite.test("a",
		Pipeline{Input{{Type: STR, Start: 0, End: 1, Lit: "a"}}})
	suite.test("a(x  = y,  f='g')",
		Pipeline{Step{
			Name: Token{Type: 3, Lit: "a", Start: 0, End: 1},
			Params: map[Token]Token{
				Token{Type: 3, Lit: "f", Start: 11, End: 12}: Token{Type: 4, Lit: "'g'", Start: 13, End: 16},
				Token{Type: 3, Lit: "x", Start: 2, End: 3}:   Token{Type: 3, Lit: "y", Start: 7, End: 8}}}})
	suite.test("a b c",
		Pipeline{Input{
			{Type: STR, Start: 0, End: 1, Lit: "a"},
			{Type: STR, Start: 2, End: 3, Lit: "b"},
			{Type: STR, Start: 4, End: 5, Lit: "c"},
		}})
	suite.test("  x  (  )  ",
		Pipeline{Step{
			Name: Token{Type: STR, Start: 2, End: 3, Lit: "x"}}})
	suite.test("{a}", Pipeline{Pipelines{Pipeline{
		Input{{Type: STR, Start: 1, End: 2, Lit: "a"}}}}})
	suite.test("{a()}",
		Pipeline{Pipelines{Pipeline{Step{
			Name: Token{Type: STR, Start: 1, End: 2, Lit: "a"}}}}})
	suite.test("a->b;b->c",
		Pipeline{Pipelines{
			Pipeline{Input{{Type: STR, Lit: "a", Start: 0, End: 1}}, Output{Type: STR, Lit: "b", Start: 3, End: 4}},
			Pipeline{Input{{Type: STR, Lit: "b", Start: 5, End: 6}}, Output{Type: STR, Lit: "c", Start: 8, End: 9}}}})
	suite.test("a()->b()->c();d()->e()->f()",
		Pipeline{Pipelines{
			Pipeline{Step{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}},
				Step{Name: Token{Type: STR, Lit: "b", Start: 5, End: 6}},
				Step{Name: Token{Type: STR, Lit: "c", Start: 10, End: 11}}},
			Pipeline{Step{Name: Token{Type: STR, Lit: "d", Start: 14, End: 15}},
				Step{Name: Token{Type: STR, Lit: "e", Start: 19, End: 20}},
				Step{Name: Token{Type: STR, Lit: "f", Start: 24, End: 25}}}}})
	suite.test("{a->b()}->{a()->b}",
		Pipeline{
			Pipelines{Pipeline{
				Input{{Type: STR, Lit: "a", Start: 1, End: 2}},
				Step{Name: Token{Type: STR, Lit: "b", Start: 4, End: 5}}}},
			Pipelines{Pipeline{
				Step{Name: Token{Type: STR, Lit: "a", Start: 11, End: 12}},
				Output{Type: STR, Lit: "b", Start: 16, End: 17}}}})
	suite.test("a->{a()->b}->c", Pipeline{
		Input{{Type: STR, Lit: "a", Start: 0, End: 1}},
		Pipelines{Pipeline{Step{Name: Token{Type: STR, Lit: "a", Start: 4, End: 5}},
			Output{Type: STR, Lit: "b", Start: 9, End: 10}}},
		Output{Type: STR, Lit: "c", Start: 13, End: 14}})
	suite.test("a(){b->x()}->c", Pipeline{
		Fork{Step: Step{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}},
			Pipelines: Pipelines{Pipeline{Input{{Type: STR, Lit: "b", Start: 4, End: 5}},
				Step{Name: Token{Type: STR, Lit: "x", Start: 7, End: 8}}}}},
		Output{Type: STR, Lit: "c", Start: 13, End: 14}})
	suite.test("a->fork(){a x->b}->c", Pipeline{
		Input{{Type: STR, Lit: "a", Start: 0, End: 1}},
		Fork{Step: Step{Name: Token{Type: STR, Lit: "fork", Start: 3, End: 7}},
			Pipelines: Pipelines{Pipeline{
				Input{
					{Type: STR, Lit: "a", Start: 10, End: 11},
					{Type: STR, Lit: "x", Start: 12, End: 13}},
				Output{Type: STR, Lit: "b", Start: 15, End: 16}}}},
		Output{Type: STR, Lit: "c", Start: 19, End: 20}})
	suite.test("a->{ { a() } -> x }->b", Pipeline{
		Input{{Type: STR, Lit: "a", Start: 0, End: 1}},
		Pipelines{Pipeline{
			Pipelines{Pipeline{
				Step{Name: Token{Type: STR, Lit: "a", Start: 7, End: 8}}}},
			Output{Type: STR, Lit: "x", Start: 16, End: 17}}},
		Output{Type: STR, Lit: "b", Start: 21, End: 22}})
	suite.test("a->fork(){ x -> { a() } -> x }->b", Pipeline{Input{{Type: STR, Lit: "a", Start: 0, End: 1}},
		Fork{Step: Step{Name: Token{Type: STR, Lit: "fork", Start: 3, End: 7}},
			Pipelines: Pipelines{Pipeline{
				Input{{Type: STR, Lit: "x", Start: 11, End: 12}},
				Pipelines{Pipeline{
					Step{Name: Token{Type: STR, Lit: "a", Start: 18, End: 19}}}},
				Output{Type: STR, Lit: "x", Start: 27, End: 28}}}},
		Output{Type: STR, Lit: "b", Start: 32, End: 33}})
}

func (suite *parserTestSuite) TestBigExample() {
	suite.test(
		`{
			"file1" file2 -> avg();
			:111 -> slope(t=2);
			{ host:44; host:55 } -> noop()
		} -> "do stats"(a=ini) -> rrx(){ "1" -> min(); 2 -> noop() } -> { extend() -> outfile; :444 };
		host:6767 -> otherfile`,
		Pipeline{
			Pipelines{
				Pipeline{
					Pipelines{
						Pipeline{
							Input{
								{Type: QUOT_STR, Lit: `"file1"`, Start: 5, End: 12},
								{Type: STR, Lit: "file2", Start: 13, End: 18},
							},
							Step{Name: Token{Type: STR, Lit: "avg", Start: 22, End: 25}},
						},
						Pipeline{
							Input{{Type: STR, Lit: ":111", Start: 32, End: 36}},
							Step{Name: Token{Type: STR, Lit: "slope", Start: 40, End: 45},
								Params: map[Token]Token{Token{Type: STR, Lit: "t", Start: 46, End: 47}: Token{Type: STR, Lit: "2", Start: 48, End: 49}}},
						},
						Pipeline{
							Pipelines{
								Pipeline{
									Input{{Type: STR, Lit: "host:44", Start: 57, End: 64}},
								},
								Pipeline{
									Input{{Type: STR, Lit: "host:55", Start: 66, End: 73}},
								},
							},
							Step{Name: Token{Type: STR, Lit: "noop", Start: 79, End: 83}},
						},
					},
					Step{Name: Token{Type: QUOT_STR, Lit: `"do stats"`, Start: 93, End: 103},
						Params: map[Token]Token{{Type: STR, Lit: "a", Start: 104, End: 105}: {Type: STR, Lit: "ini", Start: 106, End: 109}}},
					Fork{
						Step: Step{Name: Token{Type: STR, Lit: "rrx", Start: 114, End: 117}},
						Pipelines: Pipelines{
							Pipeline{
								Input{{Type: QUOT_STR, Lit: `"1"`, Start: 121, End: 124}},
								Step{Name: Token{Type: STR, Lit: "min", Start: 128, End: 131}},
							},
							Pipeline{
								Input{{Type: STR, Lit: "2", Start: 135, End: 136}},
								Step{Name: Token{Type: STR, Lit: "noop", Start: 140, End: 144}},
							},
						},
					},
					Pipelines{
						Pipeline{
							Step{Name: Token{Type: STR, Lit: "extend", Start: 154, End: 160}},
							Output{Type: STR, Lit: "outfile", Start: 166, End: 173},
						},
						Pipeline{
							Output{Type: STR, Lit: ":444", Start: 175, End: 179},
						},
					},
				},
				Pipeline{
					Input{{Type: STR, Lit: "host:6767", Start: 185, End: 194}},
					Output{Type: STR, Lit: "otherfile", Start: 198, End: 207},
				}}})
}
