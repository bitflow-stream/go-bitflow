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
	suite.Equal("msg (at [2-4, NEXT] '->')", ParserError{
		Message: "msg",
		Pos:     Token{Start: 2, End: 4, Lit: "->", Type: NEXT},
	}.Error())
}

func (suite *parserTestSuite) TestLexerError() {
	suite.testErr("-X", Pipelines(nil), ParserError{
		Pos:     Token{Type: NEXT, Start: 0, End: 2, Lit: "-X"},
		Message: "Expected '->'",
	})
}

func (suite *parserTestSuite) TestExpectedStep() {
	suite.testErr("   ", Pipelines(nil), ParserError{
		Pos:     Token{Type: EOF, Start: 3, End: 3, Lit: string(eof)},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("a;", Pipelines(nil), ParserError{
		Pos:     Token{Type: EOF, Start: 2, End: 2, Lit: string(eof)},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr(";", Pipelines(nil), ParserError{
		Pos:     Token{Type: SEP, Start: 0, End: 1, Lit: ";"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("a->;", Pipelines(nil), ParserError{
		Pos:     Token{Type: SEP, Start: 3, End: 4, Lit: ";"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("x->{ ; }", Pipelines(nil), ParserError{
		Pos:     Token{Type: SEP, Start: 5, End: 6, Lit: ";"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("x->fork(){}", Pipelines(nil), ParserError{
		Pos:     Token{Type: CLOSE, Start: 10, End: 11, Lit: "}"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("x->fork(){;}", Pipelines(nil), ParserError{
		Pos:     Token{Type: SEP, Start: 10, End: 11, Lit: ";"},
		Message: ExpectedPipelineStepError,
	})
	suite.testErr("x->fork(){->xx}->out", Pipelines(nil), ParserError{
		Pos:     Token{Type: NEXT, Start: 10, End: 12, Lit: "->"},
		Message: ExpectedPipelineStepError,
	})
}

func (suite *parserTestSuite) TestExpectedEOF() {
	suite.testErr("x(a)(b)", Pipelines(nil), ParserError{
		Pos:     Token{Type: PARAM, Start: 4, End: 7, Lit: "(b)"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("a->x{a()}", Pipelines(nil), ParserError{
		Pos:     Token{Type: OPEN, Start: 4, End: 5, Lit: "{"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("x(a)aa", Pipelines(nil), ParserError{
		Pos:     Token{Type: STR, Start: 4, End: 6, Lit: "aa"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("x(a)}", Pipelines(nil), ParserError{
		Pos:     Token{Type: CLOSE, Start: 4, End: 5, Lit: "}"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("x(a){e->e()}{", Pipelines(nil), ParserError{
		Pos:     Token{Type: OPEN, Start: 12, End: 13, Lit: "{"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("x(a){e->e()}}", Pipelines(nil), ParserError{
		Pos:     Token{Type: CLOSE, Start: 12, End: 13, Lit: "}"},
		Message: "Expected 'EOF'",
	})
	suite.testErr("{a->b}()", Pipelines(nil), ParserError{
		Pos:     Token{Type: PARAM, Start: 6, End: 8, Lit: "()"},
		Message: "Expected 'EOF'",
	})
}

func (suite *parserTestSuite) TestExpectedClosingBracket() {
	suite.testErr("{x", Pipelines(nil), ParserError{
		Pos:     Token{Type: EOF, Start: 2, End: 2, Lit: string(eof)},
		Message: "Expected '}'",
	})
	suite.testErr("a->{ x() (d)", Pipelines(nil), ParserError{
		Pos:     Token{Type: PARAM, Start: 9, End: 12, Lit: "(d)"},
		Message: "Expected '}'",
	})
	suite.testErr("a->{ x() aa", Pipelines(nil), ParserError{
		Pos:     Token{Type: STR, Start: 9, End: 11, Lit: "aa"},
		Message: "Expected '}'",
	})
	suite.testErr("a->{ x(){a->v()} {", Pipelines(nil), ParserError{
		Pos:     Token{Type: OPEN, Start: 17, End: 18, Lit: "{"},
		Message: "Expected '}'",
	})
}

func (suite *parserTestSuite) TestValidatePipeline() {
	suite.testErr("a->b->c", Pipelines(nil), ParserError{
		Pos:     Token{Type: STR, Start: 3, End: 4, Lit: "b"},
		Message: "Intermediate pipeline step cannot be an input or output identifier",
	})
	suite.testErr("a->fork(){x()}", Pipelines(nil), ParserError{
		Pos:     Token{Type: STR, Start: 10, End: 11, Lit: "x"},
		Message: "Forked pipeline must start with a pipeline identifier (string)",
	})
	suite.testErr("a->{a->b()}", Pipelines(nil), ParserError{
		Pos:     Token{Type: STR, Start: 4, End: 5, Lit: "a"},
		Message: "Multiplexed pipeline cannot start with an identifier (string)",
	})
	suite.testErr("fork(){a}", Pipelines(nil), ParserError{
		Pos:     Token{Type: STR, Start: 7, End: 8, Lit: "a"},
		Message: "Forked pipeline must have at least one pipeline step",
	})
	suite.testErr("a->b c", Pipelines(nil), ParserError{
		Pos:     Token{Type: STR, Start: 3, End: 4, Lit: "b"},
		Message: "Multiple output edges are not allowed",
	})
}

func (suite *parserTestSuite) TestExamples() {
	suite.test("a",
		Pipelines{Pipeline{Input{
			Name: Token{Type: STR, Start: 0, End: 1, Lit: "a"}}}})
	suite.test("a b c",
		Pipelines{Pipeline{Inputs{
			Input{Name: Token{Type: STR, Start: 0, End: 1, Lit: "a"}},
			Input{Name: Token{Type: STR, Start: 2, End: 3, Lit: "b"}},
			Input{Name: Token{Type: STR, Start: 4, End: 5, Lit: "c"}},
		}}})
	suite.test("  x  (  )  ",
		Pipelines{Pipeline{Step{
			Name:   Token{Type: STR, Start: 2, End: 3, Lit: "x"},
			Params: Token{Type: PARAM, Start: 5, End: 9, Lit: "(  )"}}}})
	suite.test("{a}", Pipelines{Pipeline{Pipelines{Pipeline{Input{
		Name: Token{Type: STR, Start: 1, End: 2, Lit: "a"}}}}}})
	suite.test("{a()}",
		Pipelines{Pipeline{Pipelines{Pipeline{Step{
			Name:   Token{Type: STR, Start: 1, End: 2, Lit: "a"},
			Params: Token{Type: PARAM, Start: 2, End: 4, Lit: "()"}}}}}})
	suite.test("a->b;b->c",
		Pipelines{
			Pipeline{Input{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}}, Output{Name: Token{Type: STR, Lit: "b", Start: 3, End: 4}}},
			Pipeline{Input{Name: Token{Type: STR, Lit: "b", Start: 5, End: 6}}, Output{Name: Token{Type: STR, Lit: "c", Start: 8, End: 9}}}})
	suite.test("a()->b()->c();d()->e()->f()",
		Pipelines{
			Pipeline{Step{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}, Params: Token{Type: PARAM, Lit: "()", Start: 1, End: 3}},
				Step{Name: Token{Type: STR, Lit: "b", Start: 5, End: 6}, Params: Token{Type: PARAM, Lit: "()", Start: 6, End: 8}},
				Step{Name: Token{Type: STR, Lit: "c", Start: 10, End: 11}, Params: Token{Type: PARAM, Lit: "()", Start: 11, End: 13}}},
			Pipeline{Step{Name: Token{Type: STR, Lit: "d", Start: 14, End: 15}, Params: Token{Type: PARAM, Lit: "()", Start: 15, End: 17}},
				Step{Name: Token{Type: STR, Lit: "e", Start: 19, End: 20}, Params: Token{Type: PARAM, Lit: "()", Start: 20, End: 22}},
				Step{Name: Token{Type: STR, Lit: "f", Start: 24, End: 25}, Params: Token{Type: PARAM, Lit: "()", Start: 25, End: 27}}}})
	suite.test("{a->b()}->{a()->b}",
		Pipelines{Pipeline{
			Pipelines{Pipeline{
				Input{Name: Token{Type: STR, Lit: "a", Start: 1, End: 2}},
				Step{Name: Token{Type: STR, Lit: "b", Start: 4, End: 5}, Params: Token{Type: PARAM, Lit: "()", Start: 5, End: 7}}}},
			Pipelines{Pipeline{
				Step{Name: Token{Type: STR, Lit: "a", Start: 11, End: 12}, Params: Token{Type: PARAM, Lit: "()", Start: 12, End: 14}},
				Output{Name: Token{Type: STR, Lit: "b", Start: 16, End: 17}}}}}})
	suite.test("a->{a()->b}->c", Pipelines{Pipeline{
		Input{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}},
		Pipelines{Pipeline{Step{Name: Token{Type: STR, Lit: "a", Start: 4, End: 5}, Params: Token{Type: PARAM, Lit: "()", Start: 5, End: 7}},
			Output{Name: Token{Type: STR, Lit: "b", Start: 9, End: 10}}}},
		Output{Name: Token{Type: STR, Lit: "c", Start: 13, End: 14}}}})
	suite.test("a(){b->x()}->c", Pipelines{Pipeline{
		Fork{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}, Params: Token{Type: PARAM, Lit: "()", Start: 1, End: 3},
			Pipelines: Pipelines{Pipeline{Input{Name: Token{Type: STR, Lit: "b", Start: 4, End: 5}},
				Step{Name: Token{Type: STR, Lit: "x", Start: 7, End: 8}, Params: Token{Type: PARAM, Lit: "()", Start: 8, End: 10}}}}},
		Output{Name: Token{Type: STR, Lit: "c", Start: 13, End: 14}}}})
	suite.test("a->fork(){a x->b}->c", Pipelines{Pipeline{
		Input{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}},
		Fork{Name: Token{Type: STR, Lit: "fork", Start: 3, End: 7}, Params: Token{Type: PARAM, Lit: "()", Start: 7, End: 9},
			Pipelines: Pipelines{Pipeline{
				Inputs{
					Input{Name: Token{Type: STR, Lit: "a", Start: 10, End: 11}},
					Input{Name: Token{Type: STR, Lit: "x", Start: 12, End: 13}}},
				Output{Name: Token{Type: STR, Lit: "b", Start: 15, End: 16}}}}},
		Output{Name: Token{Type: STR, Lit: "c", Start: 19, End: 20}}}})
	suite.test("a->{ { a() } -> x }->b", Pipelines{Pipeline{
		Input{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}},
		Pipelines{Pipeline{Pipelines{Pipeline{
			Step{Name: Token{Type: STR, Lit: "a", Start: 7, End: 8}, Params: Token{Type: PARAM, Lit: "()", Start: 8, End: 10}}}},
			Output{Name: Token{Type: STR, Lit: "x", Start: 16, End: 17}}}},
		Output{Name: Token{Type: STR, Lit: "b", Start: 21, End: 22}}}})
	suite.test("a->fork(){ x -> { a() } -> x }->b", Pipelines{Pipeline{Input{Name: Token{Type: STR, Lit: "a", Start: 0, End: 1}},
		Fork{Name: Token{Type: STR, Lit: "fork", Start: 3, End: 7}, Params: Token{Type: PARAM, Lit: "()", Start: 7, End: 9},
			Pipelines: Pipelines{Pipeline{
				Input{Name: Token{Type: STR, Lit: "x", Start: 11, End: 12}},
				Pipelines{Pipeline{
					Step{Name: Token{Type: STR, Lit: "a", Start: 18, End: 19}, Params: Token{Type: PARAM, Lit: "()", Start: 19, End: 21}}}},
				Output{Name: Token{Type: STR, Lit: "x", Start: 27, End: 28}}}}},
		Output{Name: Token{Type: STR, Lit: "b", Start: 32, End: 33}}}})
}

func (suite *parserTestSuite) TestBigExample() {
	suite.test(
		`{
			"file1" file2 -> avg();
			:111 -> slope(t=2);
			{ host:44; host:55 } -> noop()
		} -> "do stats"(a.ini) -> rr(2){ "1" -> min(); 2 -> noop() } -> { extend() -> outfile; :444 };
		host:6767 -> otherfile`,
		Pipelines{
			Pipeline{
				Pipelines{
					Pipeline{
						Inputs{
							Input{Name: Token{Type: QUOT_STR, Lit: `"file1"`, Start: 5, End: 12}},
							Input{Name: Token{Type: STR, Lit: "file2", Start: 13, End: 18}},
						},
						Step{Name: Token{Type: STR, Lit: "avg", Start: 22, End: 25}, Params: Token{Type: PARAM, Lit: "()", Start: 25, End: 27}},
					},
					Pipeline{
						Input{Name: Token{Type: STR, Lit: ":111", Start: 32, End: 36}},
						Step{Name: Token{Type: STR, Lit: "slope", Start: 40, End: 45}, Params: Token{Type: PARAM, Lit: "(t=2)", Start: 45, End: 50}},
					},
					Pipeline{
						Pipelines{
							Pipeline{
								Input{Name: Token{Type: STR, Lit: "host:44", Start: 57, End: 64}},
							},
							Pipeline{
								Input{Name: Token{Type: STR, Lit: "host:55", Start: 66, End: 73}},
							},
						},
						Step{Name: Token{Type: STR, Lit: "noop", Start: 79, End: 83}, Params: Token{Type: PARAM, Lit: "()", Start: 83, End: 85}},
					},
				},
				Step{Name: Token{Type: QUOT_STR, Lit: `"do stats"`, Start: 93, End: 103}, Params: Token{Type: PARAM, Lit: "(a.ini)", Start: 103, End: 110}},
				Fork{
					Name:   Token{Type: STR, Lit: "rr", Start: 114, End: 116},
					Params: Token{Type: PARAM, Lit: "(2)", Start: 116, End: 119},
					Pipelines: Pipelines{
						Pipeline{
							Input{Name: Token{Type: QUOT_STR, Lit: `"1"`, Start: 121, End: 124}},
							Step{Name: Token{Type: STR, Lit: "min", Start: 128, End: 131}, Params: Token{Type: PARAM, Lit: "()", Start: 131, End: 133}},
						},
						Pipeline{
							Input{Name: Token{Type: STR, Lit: "2", Start: 135, End: 136}},
							Step{Name: Token{Type: STR, Lit: "noop", Start: 140, End: 144}, Params: Token{Type: PARAM, Lit: "()", Start: 144, End: 146}},
						},
					},
				},
				Pipelines{
					Pipeline{
						Step{Name: Token{Type: STR, Lit: "extend", Start: 154, End: 160}, Params: Token{Type: PARAM, Lit: "()", Start: 160, End: 162}},
						Output{Name: Token{Type: STR, Lit: "outfile", Start: 166, End: 173}},
					},
					Pipeline{
						Output{Name: Token{Type: STR, Lit: ":444", Start: 175, End: 179}},
					},
				},
			},
			Pipeline{
				Input{Name: Token{Type: STR, Lit: "host:6767", Start: 185, End: 194}},
				Output{Name: Token{Type: STR, Lit: "otherfile", Start: 198, End: 207}},
			},
		})
}
