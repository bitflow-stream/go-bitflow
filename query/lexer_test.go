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
			pos += len(expectedTok.Lit)
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

func (suite *lexerTestSuite) TestStr() {
	suite.test("xx- - a-b \"c =,(){}'`c v\" ` \" =,(){} a '`a\"d\"`c`'=,(){}t'",
		[]Token{
			{Type: STR, Lit: "xx-"},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "-"},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "a-b"},
			{Type: WS, Lit: " "},
			{Type: QUOT_STR, Lit: "\"c =,(){}'`c v\""},
			{Type: WS, Lit: " "},
			{Type: QUOT_STR, Lit: "` \" =,(){} a '`"},
			{Type: STR, Lit: "a"},
			{Type: QUOT_STR, Lit: "\"d\""},
			{Type: QUOT_STR, Lit: "`c`"},
			{Type: QUOT_STR, Lit: "'=,(){}t'"},
			{Type: EOF, Lit: string(eof)},
		})

	suite.test("- -",
		[]Token{
			{Type: STR, Lit: "-"},
			{Type: WS, Lit: " "},
			{Type: STR, Lit: "-"},
			{Type: EOF, Lit: string(eof)},
		})
}

func (suite *lexerTestSuite) TestParam() {
	suite.test(`(a=b)( ("a,a"=='b=(b',  d='ee'))`,
		[]Token{
			{Type: PARAM_OPEN, Lit: "("},
			{Type: STR, Lit: "a"},
			{Type: PARAM_EQ, Lit: "="},
			{Type: STR, Lit: "b"},
			{Type: PARAM_CLOSE, Lit: ")"},
			{Type: PARAM_OPEN, Lit: "("},
			{Type: WS, Lit: " "},
			{Type: PARAM_OPEN, Lit: "("},
			{Type: QUOT_STR, Lit: "\"a,a\""},
			{Type: PARAM_EQ, Lit: "="},
			{Type: PARAM_EQ, Lit: "="},
			{Type: QUOT_STR, Lit: "'b=(b'"},
			{Type: PARAM_SEP, Lit: ","},
			{Type: WS, Lit: "  "},
			{Type: STR, Lit: "d"},
			{Type: PARAM_EQ, Lit: "="},
			{Type: QUOT_STR, Lit: "'ee'"},
			{Type: PARAM_CLOSE, Lit: ")"},
			{Type: PARAM_CLOSE, Lit: ")"},
			{Type: EOF, Lit: string(eof)},
		})
}

func (suite *lexerTestSuite) TestExample() {
	suite.test(
		`{ "file1" file2 -> avg;:111->slope(t=2);host:44} -> "do stats"(file=a.ini)->`+
			`fork(rr=2,'parallel'=true){"1"->min;2->noop}->{extend->outfile;:444} ; host:6767->otherfile`,
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
			{Type: PARAM_OPEN, Lit: "("},
			{Type: STR, Lit: "t"},
			{Type: PARAM_EQ, Lit: "="},
			{Type: STR, Lit: "2"},
			{Type: PARAM_CLOSE, Lit: ")"},

			{Type: SEP, Lit: ";"},
			{Type: STR, Lit: "host:44"},
			{Type: CLOSE, Lit: "}"},
			{Type: WS, Lit: " "},
			{Type: NEXT, Lit: "->"},
			{Type: WS, Lit: " "},

			{Type: QUOT_STR, Lit: "\"do stats\""},
			{Type: PARAM_OPEN, Lit: "("},
			{Type: STR, Lit: "file"},
			{Type: PARAM_EQ, Lit: "="},
			{Type: STR, Lit: "a.ini"},
			{Type: PARAM_CLOSE, Lit: ")"},

			{Type: NEXT, Lit: "->"},

			{Type: STR, Lit: "fork"},
			{Type: PARAM_OPEN, Lit: "("},
			{Type: STR, Lit: "rr"},
			{Type: PARAM_EQ, Lit: "="},
			{Type: STR, Lit: "2"},
			{Type: PARAM_SEP, Lit: ","},
			{Type: QUOT_STR, Lit: "'parallel'"},
			{Type: PARAM_EQ, Lit: "="},
			{Type: STR, Lit: "true"},
			{Type: PARAM_CLOSE, Lit: ")"},

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
	suite.testErr("x\"XX", []Token{
		{Type: STR, Lit: "x"},
		{Type: QUOT_STR, Lit: "\"XX"},
		{Type: EOF, Lit: string(eof)},
	}, 1, fmt.Errorf(ErrorMissingQuote, "\""))

	suite.testErr("x'XX", []Token{
		{Type: STR, Lit: "x"},
		{Type: QUOT_STR, Lit: "'XX"},
		{Type: EOF, Lit: string(eof)},
	}, 1, fmt.Errorf(ErrorMissingQuote, "'"))

	suite.testErr("x`XX", []Token{
		{Type: STR, Lit: "x"},
		{Type: QUOT_STR, Lit: "`XX"},
		{Type: EOF, Lit: string(eof)},
	}, 1, fmt.Errorf(ErrorMissingQuote, "`"))
}

func (suite *lexerTestSuite) TestContent() {
	suite.Equal("xx", Token{Type: STR, Lit: "xx"}.Content())
	suite.Equal("(xx)", Token{Type: QUOT_STR, Lit: "\"(xx)\""}.Content())
	suite.Equal("xx", Token{Type: QUOT_STR, Lit: "\"xx\""}.Content())
	suite.Equal("xx", Token{Type: QUOT_STR, Lit: "'xx'"}.Content())
	suite.Equal("xx", Token{Type: QUOT_STR, Lit: "`xx`"}.Content())
}
