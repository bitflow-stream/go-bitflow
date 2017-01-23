package query

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type lexerTestSuite struct {
	t *testing.T
	*require.Assertions
}

func TestLexer(t *testing.T) {
	suite.Run(t, new(lexerTestSuite))
}

func (suite *lexerTestSuite) T() *testing.T {
	return suite.t
}

func (suite *lexerTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (suite *lexerTestSuite) testErr(input string, tokens []Token, errIndex int, expectedErr error) {
	s := NewScanner(bytes.NewBufferString(input))
	var tok Token
	var err error
	pos := 0
	for i, expectedTok := range tokens {
		tok, err = s.Scan()
		if i == errIndex && expectedErr != nil {
			suite.Error(err)
			suite.Equal(expectedErr, err)
		} else {
			suite.NoError(err)
		}
		expectedTok.Start = pos
		if tok.Lit != string(eof) {
			pos += len(tok.Lit)
		}
		expectedTok.End = pos
		suite.Equal(expectedTok, tok, fmt.Sprintf("Wrong token %v", i))
	}

	// After the end/error, the scanner should keep returning the same thing
	for i := 0; i < 3; i++ {
		tok2, err2 := s.Scan()
		suite.Equal(tok, tok2)
		suite.Equal(err, err2)
	}
}

func (suite *lexerTestSuite) test(input string, tokens []Token) {
	suite.testErr(input, tokens, -1, nil)
}

func (suite *lexerTestSuite) TestEmpty() {
	suite.test("", []Token{
		{Type: EOF, Lit: string(eof)},
	})
}

func (suite *lexerTestSuite) TestWhitespace() {
	ws := "\t  \n\n\t  \n"
	suite.test(ws, []Token{
		{Type: WS, Lit: ws},
		{Type: EOF, Lit: string(eof)},
	})
}

func (suite *lexerTestSuite) TestOperators() {
	suite.test("  ;{ \n;\n }}\t}{{  ;-> {->->}  {->",
		[]Token{
			{Type: WS, Lit: "  "},
			{Type: SEP, Lit: ";"},
			{Type: OPEN, Lit: "{"},
			{Type: WS, Lit: " \n"},
			{Type: SEP, Lit: ";"},
			{Type: WS, Lit: "\n "},
			{Type: CLOSE, Lit: "}"},
			{Type: CLOSE, Lit: "}"},
			{Type: WS, Lit: "\t"},
			{Type: CLOSE, Lit: "}"},
			{Type: OPEN, Lit: "{"},
			{Type: OPEN, Lit: "{"},
			{Type: WS, Lit: "  "},
			{Type: SEP, Lit: ";"},
			{Type: NEXT, Lit: "->"},
			{Type: WS, Lit: " "},
			{Type: OPEN, Lit: "{"},
			{Type: NEXT, Lit: "->"},
			{Type: NEXT, Lit: "->"},
			{Type: CLOSE, Lit: "}"},
			{Type: WS, Lit: "  "},
			{Type: OPEN, Lit: "{"},
			{Type: NEXT, Lit: "->"},
			{Type: EOF, Lit: string(eof)},
		})
}

func (suite *lexerTestSuite) TestStrParam() {
	suite.test("xx \"c c\nv\" a d (a b c) (\"a\"\n \"c\",--v) \"x\"xxx\"tt\" cc(({(}(cc() \"{f{f(ff)\"(g)",
		[]Token{
			{Type: STR, Lit: "xx"},
			{Type: WS, Lit: " "},
			{Type: QUOT_STR, Lit: "\"c c\nv\""},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "a"},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "d"},
			{Type: WS, Lit: " "},
			{Type: PARAM, Lit: "(a b c)"},
			{Type: WS, Lit: " "},
			{Type: PARAM, Lit: "(\"a\"\n \"c\",--v)"},
			{Type: WS, Lit: " "},
			{Type: QUOT_STR, Lit: "\"x\""},
			{Type: STR, Lit: "xxx"},
			{Type: QUOT_STR, Lit: "\"tt\""},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "cc"},
			{Type: PARAM, Lit: "(({(}(cc()"},
			{Type: WS, Lit: " "},
			{Type: QUOT_STR, Lit: "\"{f{f(ff)\""},
			{Type: PARAM, Lit: "(g)"},
			{Type: EOF, Lit: string(eof)},
		})
}

func (suite *lexerTestSuite) TestExample() {
	suite.test(
		`{ "file1" file2 -> avg;:111->slope(t=2);host:44} -> "do stats"(a.ini)->`+
			`fork(rr=2){"1"->min;2->noop}->{extend->outfile;:444} ; host:6767->otherfile`,
		[]Token{
			{Type: OPEN, Lit: "{"},
			{Type: WS, Lit: " "},
			{Type: QUOT_STR, Lit: "\"file1\""},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "file2"},
			{Type: WS, Lit: " "},
			{Type: NEXT, Lit: "->"},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "avg"},
			{Type: SEP, Lit: ";"},
			{Type: STR, Lit: ":111"},
			{Type: NEXT, Lit: "->"},
			{Type: STR, Lit: "slope"},
			{Type: PARAM, Lit: "(t=2)"},
			{Type: SEP, Lit: ";"},
			{Type: STR, Lit: "host:44"},
			{Type: CLOSE, Lit: "}"},
			{Type: WS, Lit: " "},
			{Type: NEXT, Lit: "->"},
			{Type: WS, Lit: " "},
			{Type: QUOT_STR, Lit: "\"do stats\""},
			{Type: PARAM, Lit: "(a.ini)"},
			{Type: NEXT, Lit: "->"},

			{Type: STR, Lit: "fork"},
			{Type: PARAM, Lit: "(rr=2)"},
			{Type: OPEN, Lit: "{"},
			{Type: QUOT_STR, Lit: "\"1\""},
			{Type: NEXT, Lit: "->"},
			{Type: STR, Lit: "min"},
			{Type: SEP, Lit: ";"},
			{Type: STR, Lit: "2"},
			{Type: NEXT, Lit: "->"},
			{Type: STR, Lit: "noop"},
			{Type: CLOSE, Lit: "}"},
			{Type: NEXT, Lit: "->"},
			{Type: OPEN, Lit: "{"},
			{Type: STR, Lit: "extend"},
			{Type: NEXT, Lit: "->"},
			{Type: STR, Lit: "outfile"},
			{Type: SEP, Lit: ";"},
			{Type: STR, Lit: ":444"},
			{Type: CLOSE, Lit: "}"},

			{Type: WS, Lit: " "},
			{Type: SEP, Lit: ";"},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "host:6767"},
			{Type: NEXT, Lit: "->"},
			{Type: STR, Lit: "otherfile"},

			{Type: EOF, Lit: string(eof)},
		})
}

func (suite *lexerTestSuite) TestErrors() {
	suite.testErr("x- (o)", []Token{
		{Type: STR, Lit: "x"},
		{Type: NEXT, Lit: "- "},
		{Type: PARAM, Lit: "(o)"}, // Continue lexing normally after this error
		{Type: EOF, Lit: string(eof)},
	}, 1, ErrorMissingNext)

	suite.testErr("x\"XX", []Token{
		{Type: STR, Lit: "x"},
		{Type: QUOT_STR, Lit: "\"XX"},
		{Type: EOF, Lit: string(eof)},
	}, 1, ErrorMissingQuote)

	suite.testErr("x(XX", []Token{
		{Type: STR, Lit: "x"},
		{Type: PARAM, Lit: "(XX"},
		{Type: EOF, Lit: string(eof)},
	}, 1, ErrorMissingClosingBracket)
}

func (suite *lexerTestSuite) TestContent() {
	suite.Equal("xx", Token{Type: STR, Lit: "xx"}.Content())
	suite.Equal("xx", Token{Type: QUOT_STR, Lit: "\"xx\""}.Content())
	suite.Equal("(xx)", Token{Type: QUOT_STR, Lit: "\"(xx)\""}.Content())
	suite.Equal(" xx ", Token{Type: PARAM, Lit: "( xx )"}.Content())
}
