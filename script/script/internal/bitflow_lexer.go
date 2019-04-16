// Code generated from Bitflow.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 19, 124,
	8, 1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7,
	9, 7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12,
	4, 13, 9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 4,
	18, 9, 18, 3, 2, 3, 2, 3, 3, 3, 3, 3, 4, 3, 4, 3, 5, 3, 5, 3, 5, 3, 6,
	3, 6, 3, 7, 3, 7, 3, 8, 3, 8, 3, 9, 3, 9, 3, 10, 3, 10, 3, 11, 3, 11, 3,
	12, 3, 12, 3, 12, 3, 12, 3, 12, 3, 12, 3, 13, 3, 13, 7, 13, 67, 10, 13,
	12, 13, 14, 13, 70, 11, 13, 3, 13, 3, 13, 3, 13, 7, 13, 75, 10, 13, 12,
	13, 14, 13, 78, 11, 13, 3, 13, 3, 13, 3, 13, 7, 13, 83, 10, 13, 12, 13,
	14, 13, 86, 11, 13, 3, 13, 5, 13, 89, 10, 13, 3, 14, 6, 14, 92, 10, 14,
	13, 14, 14, 14, 93, 3, 15, 3, 15, 7, 15, 98, 10, 15, 12, 15, 14, 15, 101,
	11, 15, 3, 15, 3, 15, 3, 15, 3, 15, 3, 16, 3, 16, 3, 16, 5, 16, 110, 10,
	16, 3, 16, 3, 16, 3, 17, 3, 17, 3, 17, 5, 17, 117, 10, 17, 3, 17, 3, 17,
	3, 18, 3, 18, 3, 18, 3, 18, 5, 68, 76, 84, 2, 19, 3, 3, 5, 4, 7, 5, 9,
	6, 11, 7, 13, 8, 15, 9, 17, 10, 19, 11, 21, 12, 23, 13, 25, 14, 27, 15,
	29, 16, 31, 17, 33, 18, 35, 19, 3, 2, 4, 10, 2, 39, 40, 44, 45, 47, 60,
	65, 65, 67, 92, 94, 94, 97, 97, 99, 124, 4, 2, 12, 12, 15, 15, 2, 132,
	2, 3, 3, 2, 2, 2, 2, 5, 3, 2, 2, 2, 2, 7, 3, 2, 2, 2, 2, 9, 3, 2, 2, 2,
	2, 11, 3, 2, 2, 2, 2, 13, 3, 2, 2, 2, 2, 15, 3, 2, 2, 2, 2, 17, 3, 2, 2,
	2, 2, 19, 3, 2, 2, 2, 2, 21, 3, 2, 2, 2, 2, 23, 3, 2, 2, 2, 2, 25, 3, 2,
	2, 2, 2, 27, 3, 2, 2, 2, 2, 29, 3, 2, 2, 2, 2, 31, 3, 2, 2, 2, 2, 33, 3,
	2, 2, 2, 2, 35, 3, 2, 2, 2, 3, 37, 3, 2, 2, 2, 5, 39, 3, 2, 2, 2, 7, 41,
	3, 2, 2, 2, 9, 43, 3, 2, 2, 2, 11, 46, 3, 2, 2, 2, 13, 48, 3, 2, 2, 2,
	15, 50, 3, 2, 2, 2, 17, 52, 3, 2, 2, 2, 19, 54, 3, 2, 2, 2, 21, 56, 3,
	2, 2, 2, 23, 58, 3, 2, 2, 2, 25, 88, 3, 2, 2, 2, 27, 91, 3, 2, 2, 2, 29,
	95, 3, 2, 2, 2, 31, 109, 3, 2, 2, 2, 33, 116, 3, 2, 2, 2, 35, 120, 3, 2,
	2, 2, 37, 38, 7, 125, 2, 2, 38, 4, 3, 2, 2, 2, 39, 40, 7, 127, 2, 2, 40,
	6, 3, 2, 2, 2, 41, 42, 7, 61, 2, 2, 42, 8, 3, 2, 2, 2, 43, 44, 7, 47, 2,
	2, 44, 45, 7, 64, 2, 2, 45, 10, 3, 2, 2, 2, 46, 47, 7, 42, 2, 2, 47, 12,
	3, 2, 2, 2, 48, 49, 7, 43, 2, 2, 49, 14, 3, 2, 2, 2, 50, 51, 7, 63, 2,
	2, 51, 16, 3, 2, 2, 2, 52, 53, 7, 46, 2, 2, 53, 18, 3, 2, 2, 2, 54, 55,
	7, 93, 2, 2, 55, 20, 3, 2, 2, 2, 56, 57, 7, 95, 2, 2, 57, 22, 3, 2, 2,
	2, 58, 59, 7, 100, 2, 2, 59, 60, 7, 99, 2, 2, 60, 61, 7, 118, 2, 2, 61,
	62, 7, 101, 2, 2, 62, 63, 7, 106, 2, 2, 63, 24, 3, 2, 2, 2, 64, 68, 7,
	36, 2, 2, 65, 67, 11, 2, 2, 2, 66, 65, 3, 2, 2, 2, 67, 70, 3, 2, 2, 2,
	68, 69, 3, 2, 2, 2, 68, 66, 3, 2, 2, 2, 69, 71, 3, 2, 2, 2, 70, 68, 3,
	2, 2, 2, 71, 89, 7, 36, 2, 2, 72, 76, 7, 41, 2, 2, 73, 75, 11, 2, 2, 2,
	74, 73, 3, 2, 2, 2, 75, 78, 3, 2, 2, 2, 76, 77, 3, 2, 2, 2, 76, 74, 3,
	2, 2, 2, 77, 79, 3, 2, 2, 2, 78, 76, 3, 2, 2, 2, 79, 89, 7, 41, 2, 2, 80,
	84, 7, 98, 2, 2, 81, 83, 11, 2, 2, 2, 82, 81, 3, 2, 2, 2, 83, 86, 3, 2,
	2, 2, 84, 85, 3, 2, 2, 2, 84, 82, 3, 2, 2, 2, 85, 87, 3, 2, 2, 2, 86, 84,
	3, 2, 2, 2, 87, 89, 7, 98, 2, 2, 88, 64, 3, 2, 2, 2, 88, 72, 3, 2, 2, 2,
	88, 80, 3, 2, 2, 2, 89, 26, 3, 2, 2, 2, 90, 92, 9, 2, 2, 2, 91, 90, 3,
	2, 2, 2, 92, 93, 3, 2, 2, 2, 93, 91, 3, 2, 2, 2, 93, 94, 3, 2, 2, 2, 94,
	28, 3, 2, 2, 2, 95, 99, 7, 37, 2, 2, 96, 98, 10, 3, 2, 2, 97, 96, 3, 2,
	2, 2, 98, 101, 3, 2, 2, 2, 99, 97, 3, 2, 2, 2, 99, 100, 3, 2, 2, 2, 100,
	102, 3, 2, 2, 2, 101, 99, 3, 2, 2, 2, 102, 103, 5, 31, 16, 2, 103, 104,
	3, 2, 2, 2, 104, 105, 8, 15, 2, 2, 105, 30, 3, 2, 2, 2, 106, 110, 9, 3,
	2, 2, 107, 108, 7, 15, 2, 2, 108, 110, 7, 12, 2, 2, 109, 106, 3, 2, 2,
	2, 109, 107, 3, 2, 2, 2, 110, 111, 3, 2, 2, 2, 111, 112, 8, 16, 2, 2, 112,
	32, 3, 2, 2, 2, 113, 117, 7, 34, 2, 2, 114, 115, 7, 94, 2, 2, 115, 117,
	7, 117, 2, 2, 116, 113, 3, 2, 2, 2, 116, 114, 3, 2, 2, 2, 117, 118, 3,
	2, 2, 2, 118, 119, 8, 17, 2, 2, 119, 34, 3, 2, 2, 2, 120, 121, 7, 11, 2,
	2, 121, 122, 3, 2, 2, 2, 122, 123, 8, 18, 2, 2, 123, 36, 3, 2, 2, 2, 11,
	2, 68, 76, 84, 88, 93, 99, 109, 116, 3, 8, 2, 2,
}

var lexerDeserializer = antlr.NewATNDeserializer(nil)
var lexerAtn = lexerDeserializer.DeserializeFromUInt16(serializedLexerAtn)

var lexerChannelNames = []string{
	"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
}

var lexerModeNames = []string{
	"DEFAULT_MODE",
}

var lexerLiteralNames = []string{
	"", "'{'", "'}'", "';'", "'->'", "'('", "')'", "'='", "','", "'['", "']'",
	"'batch'", "", "", "", "", "", "'\t'",
}

var lexerSymbolicNames = []string{
	"", "OPEN", "CLOSE", "EOP", "NEXT", "OPEN_PARAMS", "CLOSE_PARAMS", "EQ",
	"SEP", "OPEN_HINTS", "CLOSE_HINTS", "WINDOW", "STRING", "IDENTIFIER", "COMMENT",
	"NEWLINE", "WHITESPACE", "TAB",
}

var lexerRuleNames = []string{
	"OPEN", "CLOSE", "EOP", "NEXT", "OPEN_PARAMS", "CLOSE_PARAMS", "EQ", "SEP",
	"OPEN_HINTS", "CLOSE_HINTS", "WINDOW", "STRING", "IDENTIFIER", "COMMENT",
	"NEWLINE", "WHITESPACE", "TAB",
}

type BitflowLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var lexerDecisionToDFA = make([]*antlr.DFA, len(lexerAtn.DecisionToState))

func init() {
	for index, ds := range lexerAtn.DecisionToState {
		lexerDecisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

func NewBitflowLexer(input antlr.CharStream) *BitflowLexer {

	l := new(BitflowLexer)

	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
	l.GrammarFileName = "Bitflow.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// BitflowLexer tokens.
const (
	BitflowLexerOPEN         = 1
	BitflowLexerCLOSE        = 2
	BitflowLexerEOP          = 3
	BitflowLexerNEXT         = 4
	BitflowLexerOPEN_PARAMS  = 5
	BitflowLexerCLOSE_PARAMS = 6
	BitflowLexerEQ           = 7
	BitflowLexerSEP          = 8
	BitflowLexerOPEN_HINTS   = 9
	BitflowLexerCLOSE_HINTS  = 10
	BitflowLexerWINDOW       = 11
	BitflowLexerSTRING       = 12
	BitflowLexerIDENTIFIER   = 13
	BitflowLexerCOMMENT      = 14
	BitflowLexerNEWLINE      = 15
	BitflowLexerWHITESPACE   = 16
	BitflowLexerTAB          = 17
)
