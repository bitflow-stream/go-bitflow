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
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 19, 125,
	8, 1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7,
	9, 7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12,
	4, 13, 9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 4,
	18, 9, 18, 3, 2, 3, 2, 3, 3, 3, 3, 3, 4, 3, 4, 3, 5, 3, 5, 3, 5, 3, 6,
	3, 6, 3, 7, 3, 7, 3, 8, 3, 8, 3, 9, 3, 9, 3, 10, 3, 10, 3, 11, 3, 11, 3,
	12, 3, 12, 3, 12, 3, 12, 3, 12, 3, 12, 3, 12, 3, 13, 3, 13, 7, 13, 68,
	10, 13, 12, 13, 14, 13, 71, 11, 13, 3, 13, 3, 13, 3, 13, 7, 13, 76, 10,
	13, 12, 13, 14, 13, 79, 11, 13, 3, 13, 3, 13, 3, 13, 7, 13, 84, 10, 13,
	12, 13, 14, 13, 87, 11, 13, 3, 13, 5, 13, 90, 10, 13, 3, 14, 6, 14, 93,
	10, 14, 13, 14, 14, 14, 94, 3, 15, 3, 15, 7, 15, 99, 10, 15, 12, 15, 14,
	15, 102, 11, 15, 3, 15, 3, 15, 3, 15, 3, 15, 3, 16, 3, 16, 3, 16, 5, 16,
	111, 10, 16, 3, 16, 3, 16, 3, 17, 3, 17, 3, 17, 5, 17, 118, 10, 17, 3,
	17, 3, 17, 3, 18, 3, 18, 3, 18, 3, 18, 5, 69, 77, 85, 2, 19, 3, 3, 5, 4,
	7, 5, 9, 6, 11, 7, 13, 8, 15, 9, 17, 10, 19, 11, 21, 12, 23, 13, 25, 14,
	27, 15, 29, 16, 31, 17, 33, 18, 35, 19, 3, 2, 4, 7, 2, 47, 60, 67, 92,
	94, 94, 97, 97, 99, 124, 4, 2, 12, 12, 15, 15, 2, 133, 2, 3, 3, 2, 2, 2,
	2, 5, 3, 2, 2, 2, 2, 7, 3, 2, 2, 2, 2, 9, 3, 2, 2, 2, 2, 11, 3, 2, 2, 2,
	2, 13, 3, 2, 2, 2, 2, 15, 3, 2, 2, 2, 2, 17, 3, 2, 2, 2, 2, 19, 3, 2, 2,
	2, 2, 21, 3, 2, 2, 2, 2, 23, 3, 2, 2, 2, 2, 25, 3, 2, 2, 2, 2, 27, 3, 2,
	2, 2, 2, 29, 3, 2, 2, 2, 2, 31, 3, 2, 2, 2, 2, 33, 3, 2, 2, 2, 2, 35, 3,
	2, 2, 2, 3, 37, 3, 2, 2, 2, 5, 39, 3, 2, 2, 2, 7, 41, 3, 2, 2, 2, 9, 43,
	3, 2, 2, 2, 11, 46, 3, 2, 2, 2, 13, 48, 3, 2, 2, 2, 15, 50, 3, 2, 2, 2,
	17, 52, 3, 2, 2, 2, 19, 54, 3, 2, 2, 2, 21, 56, 3, 2, 2, 2, 23, 58, 3,
	2, 2, 2, 25, 89, 3, 2, 2, 2, 27, 92, 3, 2, 2, 2, 29, 96, 3, 2, 2, 2, 31,
	110, 3, 2, 2, 2, 33, 117, 3, 2, 2, 2, 35, 121, 3, 2, 2, 2, 37, 38, 7, 125,
	2, 2, 38, 4, 3, 2, 2, 2, 39, 40, 7, 127, 2, 2, 40, 6, 3, 2, 2, 2, 41, 42,
	7, 61, 2, 2, 42, 8, 3, 2, 2, 2, 43, 44, 7, 47, 2, 2, 44, 45, 7, 64, 2,
	2, 45, 10, 3, 2, 2, 2, 46, 47, 7, 42, 2, 2, 47, 12, 3, 2, 2, 2, 48, 49,
	7, 43, 2, 2, 49, 14, 3, 2, 2, 2, 50, 51, 7, 63, 2, 2, 51, 16, 3, 2, 2,
	2, 52, 53, 7, 46, 2, 2, 53, 18, 3, 2, 2, 2, 54, 55, 7, 93, 2, 2, 55, 20,
	3, 2, 2, 2, 56, 57, 7, 95, 2, 2, 57, 22, 3, 2, 2, 2, 58, 59, 7, 121, 2,
	2, 59, 60, 7, 107, 2, 2, 60, 61, 7, 112, 2, 2, 61, 62, 7, 102, 2, 2, 62,
	63, 7, 113, 2, 2, 63, 64, 7, 121, 2, 2, 64, 24, 3, 2, 2, 2, 65, 69, 7,
	36, 2, 2, 66, 68, 11, 2, 2, 2, 67, 66, 3, 2, 2, 2, 68, 71, 3, 2, 2, 2,
	69, 70, 3, 2, 2, 2, 69, 67, 3, 2, 2, 2, 70, 72, 3, 2, 2, 2, 71, 69, 3,
	2, 2, 2, 72, 90, 7, 36, 2, 2, 73, 77, 7, 41, 2, 2, 74, 76, 11, 2, 2, 2,
	75, 74, 3, 2, 2, 2, 76, 79, 3, 2, 2, 2, 77, 78, 3, 2, 2, 2, 77, 75, 3,
	2, 2, 2, 78, 80, 3, 2, 2, 2, 79, 77, 3, 2, 2, 2, 80, 90, 7, 41, 2, 2, 81,
	85, 7, 98, 2, 2, 82, 84, 11, 2, 2, 2, 83, 82, 3, 2, 2, 2, 84, 87, 3, 2,
	2, 2, 85, 86, 3, 2, 2, 2, 85, 83, 3, 2, 2, 2, 86, 88, 3, 2, 2, 2, 87, 85,
	3, 2, 2, 2, 88, 90, 7, 98, 2, 2, 89, 65, 3, 2, 2, 2, 89, 73, 3, 2, 2, 2,
	89, 81, 3, 2, 2, 2, 90, 26, 3, 2, 2, 2, 91, 93, 9, 2, 2, 2, 92, 91, 3,
	2, 2, 2, 93, 94, 3, 2, 2, 2, 94, 92, 3, 2, 2, 2, 94, 95, 3, 2, 2, 2, 95,
	28, 3, 2, 2, 2, 96, 100, 7, 37, 2, 2, 97, 99, 10, 3, 2, 2, 98, 97, 3, 2,
	2, 2, 99, 102, 3, 2, 2, 2, 100, 98, 3, 2, 2, 2, 100, 101, 3, 2, 2, 2, 101,
	103, 3, 2, 2, 2, 102, 100, 3, 2, 2, 2, 103, 104, 5, 31, 16, 2, 104, 105,
	3, 2, 2, 2, 105, 106, 8, 15, 2, 2, 106, 30, 3, 2, 2, 2, 107, 111, 9, 3,
	2, 2, 108, 109, 7, 15, 2, 2, 109, 111, 7, 12, 2, 2, 110, 107, 3, 2, 2,
	2, 110, 108, 3, 2, 2, 2, 111, 112, 3, 2, 2, 2, 112, 113, 8, 16, 2, 2, 113,
	32, 3, 2, 2, 2, 114, 118, 7, 34, 2, 2, 115, 116, 7, 94, 2, 2, 116, 118,
	7, 117, 2, 2, 117, 114, 3, 2, 2, 2, 117, 115, 3, 2, 2, 2, 118, 119, 3,
	2, 2, 2, 119, 120, 8, 17, 2, 2, 120, 34, 3, 2, 2, 2, 121, 122, 7, 11, 2,
	2, 122, 123, 3, 2, 2, 2, 123, 124, 8, 18, 2, 2, 124, 36, 3, 2, 2, 2, 11,
	2, 69, 77, 85, 89, 94, 100, 110, 117, 3, 8, 2, 2,
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
	"'window'", "", "", "", "", "", "'\t'",
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
