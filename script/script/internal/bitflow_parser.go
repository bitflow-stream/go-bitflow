// Code generated from Bitflow.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Bitflow

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 23, 222,
	4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7,
	4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4, 13,
	9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 3, 2, 6,
	2, 36, 10, 2, 13, 2, 14, 2, 37, 3, 3, 5, 3, 41, 10, 3, 3, 3, 5, 3, 44,
	10, 3, 3, 3, 5, 3, 47, 10, 3, 3, 3, 3, 3, 3, 3, 3, 3, 6, 3, 53, 10, 3,
	13, 3, 14, 3, 54, 3, 3, 3, 3, 3, 4, 5, 4, 60, 10, 4, 3, 4, 5, 4, 63, 10,
	4, 3, 4, 5, 4, 66, 10, 4, 3, 4, 3, 4, 6, 4, 70, 10, 4, 13, 4, 14, 4, 71,
	3, 4, 3, 4, 3, 5, 3, 5, 5, 5, 78, 10, 5, 3, 5, 5, 5, 81, 10, 5, 3, 5, 3,
	5, 3, 5, 3, 5, 3, 6, 3, 6, 3, 6, 7, 6, 90, 10, 6, 12, 6, 14, 6, 93, 11,
	6, 3, 6, 5, 6, 96, 10, 6, 3, 7, 3, 7, 5, 7, 100, 10, 7, 3, 7, 5, 7, 103,
	10, 7, 3, 8, 3, 8, 5, 8, 107, 10, 8, 3, 8, 5, 8, 110, 10, 8, 3, 9, 3, 9,
	5, 9, 114, 10, 9, 3, 9, 5, 9, 117, 10, 9, 3, 10, 5, 10, 120, 10, 10, 3,
	10, 3, 10, 3, 10, 5, 10, 125, 10, 10, 3, 10, 3, 10, 3, 10, 5, 10, 130,
	10, 10, 3, 10, 3, 10, 3, 10, 3, 10, 5, 10, 136, 10, 10, 7, 10, 138, 10,
	10, 12, 10, 14, 10, 141, 11, 10, 3, 10, 5, 10, 144, 10, 10, 3, 11, 3, 11,
	3, 11, 5, 11, 149, 10, 11, 3, 11, 3, 11, 3, 11, 3, 11, 5, 11, 155, 10,
	11, 7, 11, 157, 10, 11, 12, 11, 14, 11, 160, 11, 11, 3, 11, 5, 11, 163,
	10, 11, 3, 12, 3, 12, 5, 12, 167, 10, 12, 3, 12, 3, 12, 3, 12, 3, 12, 5,
	12, 173, 10, 12, 7, 12, 175, 10, 12, 12, 12, 14, 12, 178, 11, 12, 3, 12,
	3, 12, 3, 12, 5, 12, 183, 10, 12, 3, 12, 5, 12, 186, 10, 12, 3, 13, 3,
	13, 3, 13, 3, 13, 3, 14, 3, 14, 3, 14, 3, 14, 7, 14, 196, 10, 14, 12, 14,
	14, 14, 199, 11, 14, 5, 14, 201, 10, 14, 3, 14, 3, 14, 3, 15, 3, 15, 3,
	16, 3, 16, 3, 17, 3, 17, 3, 17, 3, 17, 7, 17, 213, 10, 17, 12, 17, 14,
	17, 216, 11, 17, 5, 17, 218, 10, 17, 3, 17, 3, 17, 3, 17, 2, 2, 18, 2,
	4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 2, 5, 3, 3, 13,
	13, 3, 2, 16, 17, 3, 2, 16, 18, 2, 248, 2, 35, 3, 2, 2, 2, 4, 40, 3, 2,
	2, 2, 6, 59, 3, 2, 2, 2, 8, 75, 3, 2, 2, 2, 10, 86, 3, 2, 2, 2, 12, 97,
	3, 2, 2, 2, 14, 104, 3, 2, 2, 2, 16, 111, 3, 2, 2, 2, 18, 124, 3, 2, 2,
	2, 20, 148, 3, 2, 2, 2, 22, 166, 3, 2, 2, 2, 24, 187, 3, 2, 2, 2, 26, 191,
	3, 2, 2, 2, 28, 204, 3, 2, 2, 2, 30, 206, 3, 2, 2, 2, 32, 208, 3, 2, 2,
	2, 34, 36, 5, 22, 12, 2, 35, 34, 3, 2, 2, 2, 36, 37, 3, 2, 2, 2, 37, 35,
	3, 2, 2, 2, 37, 38, 3, 2, 2, 2, 38, 3, 3, 2, 2, 2, 39, 41, 5, 28, 15, 2,
	40, 39, 3, 2, 2, 2, 40, 41, 3, 2, 2, 2, 41, 43, 3, 2, 2, 2, 42, 44, 5,
	26, 14, 2, 43, 42, 3, 2, 2, 2, 43, 44, 3, 2, 2, 2, 44, 46, 3, 2, 2, 2,
	45, 47, 5, 32, 17, 2, 46, 45, 3, 2, 2, 2, 46, 47, 3, 2, 2, 2, 47, 48, 3,
	2, 2, 2, 48, 52, 7, 3, 2, 2, 49, 50, 5, 14, 8, 2, 50, 51, 7, 13, 2, 2,
	51, 53, 3, 2, 2, 2, 52, 49, 3, 2, 2, 2, 53, 54, 3, 2, 2, 2, 54, 52, 3,
	2, 2, 2, 54, 55, 3, 2, 2, 2, 55, 56, 3, 2, 2, 2, 56, 57, 7, 4, 2, 2, 57,
	5, 3, 2, 2, 2, 58, 60, 5, 28, 15, 2, 59, 58, 3, 2, 2, 2, 59, 60, 3, 2,
	2, 2, 60, 62, 3, 2, 2, 2, 61, 63, 5, 26, 14, 2, 62, 61, 3, 2, 2, 2, 62,
	63, 3, 2, 2, 2, 63, 65, 3, 2, 2, 2, 64, 66, 5, 32, 17, 2, 65, 64, 3, 2,
	2, 2, 65, 66, 3, 2, 2, 2, 66, 67, 3, 2, 2, 2, 67, 69, 7, 3, 2, 2, 68, 70,
	5, 18, 10, 2, 69, 68, 3, 2, 2, 2, 70, 71, 3, 2, 2, 2, 71, 69, 3, 2, 2,
	2, 71, 72, 3, 2, 2, 2, 72, 73, 3, 2, 2, 2, 73, 74, 7, 4, 2, 2, 74, 7, 3,
	2, 2, 2, 75, 77, 7, 5, 2, 2, 76, 78, 5, 26, 14, 2, 77, 76, 3, 2, 2, 2,
	77, 78, 3, 2, 2, 2, 78, 80, 3, 2, 2, 2, 79, 81, 5, 32, 17, 2, 80, 79, 3,
	2, 2, 2, 80, 81, 3, 2, 2, 2, 81, 82, 3, 2, 2, 2, 82, 83, 7, 3, 2, 2, 83,
	84, 5, 20, 11, 2, 84, 85, 7, 4, 2, 2, 85, 9, 3, 2, 2, 2, 86, 91, 5, 12,
	7, 2, 87, 88, 7, 13, 2, 2, 88, 90, 5, 12, 7, 2, 89, 87, 3, 2, 2, 2, 90,
	93, 3, 2, 2, 2, 91, 89, 3, 2, 2, 2, 91, 92, 3, 2, 2, 2, 92, 95, 3, 2, 2,
	2, 93, 91, 3, 2, 2, 2, 94, 96, 7, 13, 2, 2, 95, 94, 3, 2, 2, 2, 95, 96,
	3, 2, 2, 2, 96, 11, 3, 2, 2, 2, 97, 99, 5, 28, 15, 2, 98, 100, 5, 26, 14,
	2, 99, 98, 3, 2, 2, 2, 99, 100, 3, 2, 2, 2, 100, 102, 3, 2, 2, 2, 101,
	103, 5, 32, 17, 2, 102, 101, 3, 2, 2, 2, 102, 103, 3, 2, 2, 2, 103, 13,
	3, 2, 2, 2, 104, 106, 5, 28, 15, 2, 105, 107, 5, 26, 14, 2, 106, 105, 3,
	2, 2, 2, 106, 107, 3, 2, 2, 2, 107, 109, 3, 2, 2, 2, 108, 110, 5, 32, 17,
	2, 109, 108, 3, 2, 2, 2, 109, 110, 3, 2, 2, 2, 110, 15, 3, 2, 2, 2, 111,
	113, 5, 28, 15, 2, 112, 114, 5, 26, 14, 2, 113, 112, 3, 2, 2, 2, 113, 114,
	3, 2, 2, 2, 114, 116, 3, 2, 2, 2, 115, 117, 5, 32, 17, 2, 116, 115, 3,
	2, 2, 2, 116, 117, 3, 2, 2, 2, 117, 17, 3, 2, 2, 2, 118, 120, 7, 12, 2,
	2, 119, 118, 3, 2, 2, 2, 119, 120, 3, 2, 2, 2, 120, 121, 3, 2, 2, 2, 121,
	122, 5, 30, 16, 2, 122, 123, 7, 14, 2, 2, 123, 125, 3, 2, 2, 2, 124, 119,
	3, 2, 2, 2, 124, 125, 3, 2, 2, 2, 125, 129, 3, 2, 2, 2, 126, 130, 5, 16,
	9, 2, 127, 130, 5, 6, 4, 2, 128, 130, 5, 8, 5, 2, 129, 126, 3, 2, 2, 2,
	129, 127, 3, 2, 2, 2, 129, 128, 3, 2, 2, 2, 130, 139, 3, 2, 2, 2, 131,
	135, 7, 14, 2, 2, 132, 136, 5, 16, 9, 2, 133, 136, 5, 6, 4, 2, 134, 136,
	5, 8, 5, 2, 135, 132, 3, 2, 2, 2, 135, 133, 3, 2, 2, 2, 135, 134, 3, 2,
	2, 2, 136, 138, 3, 2, 2, 2, 137, 131, 3, 2, 2, 2, 138, 141, 3, 2, 2, 2,
	139, 137, 3, 2, 2, 2, 139, 140, 3, 2, 2, 2, 140, 143, 3, 2, 2, 2, 141,
	139, 3, 2, 2, 2, 142, 144, 9, 2, 2, 2, 143, 142, 3, 2, 2, 2, 143, 144,
	3, 2, 2, 2, 144, 19, 3, 2, 2, 2, 145, 149, 5, 16, 9, 2, 146, 149, 5, 6,
	4, 2, 147, 149, 5, 8, 5, 2, 148, 145, 3, 2, 2, 2, 148, 146, 3, 2, 2, 2,
	148, 147, 3, 2, 2, 2, 149, 158, 3, 2, 2, 2, 150, 154, 7, 14, 2, 2, 151,
	155, 5, 16, 9, 2, 152, 155, 5, 6, 4, 2, 153, 155, 5, 8, 5, 2, 154, 151,
	3, 2, 2, 2, 154, 152, 3, 2, 2, 2, 154, 153, 3, 2, 2, 2, 155, 157, 3, 2,
	2, 2, 156, 150, 3, 2, 2, 2, 157, 160, 3, 2, 2, 2, 158, 156, 3, 2, 2, 2,
	158, 159, 3, 2, 2, 2, 159, 162, 3, 2, 2, 2, 160, 158, 3, 2, 2, 2, 161,
	163, 9, 2, 2, 2, 162, 161, 3, 2, 2, 2, 162, 163, 3, 2, 2, 2, 163, 21, 3,
	2, 2, 2, 164, 167, 5, 10, 6, 2, 165, 167, 5, 12, 7, 2, 166, 164, 3, 2,
	2, 2, 166, 165, 3, 2, 2, 2, 167, 176, 3, 2, 2, 2, 168, 172, 7, 14, 2, 2,
	169, 173, 5, 16, 9, 2, 170, 173, 5, 6, 4, 2, 171, 173, 5, 8, 5, 2, 172,
	169, 3, 2, 2, 2, 172, 170, 3, 2, 2, 2, 172, 171, 3, 2, 2, 2, 173, 175,
	3, 2, 2, 2, 174, 168, 3, 2, 2, 2, 175, 178, 3, 2, 2, 2, 176, 174, 3, 2,
	2, 2, 176, 177, 3, 2, 2, 2, 177, 179, 3, 2, 2, 2, 178, 176, 3, 2, 2, 2,
	179, 182, 7, 14, 2, 2, 180, 183, 5, 14, 8, 2, 181, 183, 5, 4, 3, 2, 182,
	180, 3, 2, 2, 2, 182, 181, 3, 2, 2, 2, 183, 185, 3, 2, 2, 2, 184, 186,
	9, 2, 2, 2, 185, 184, 3, 2, 2, 2, 185, 186, 3, 2, 2, 2, 186, 23, 3, 2,
	2, 2, 187, 188, 7, 18, 2, 2, 188, 189, 7, 6, 2, 2, 189, 190, 9, 3, 2, 2,
	190, 25, 3, 2, 2, 2, 191, 200, 7, 7, 2, 2, 192, 197, 5, 24, 13, 2, 193,
	194, 7, 8, 2, 2, 194, 196, 5, 24, 13, 2, 195, 193, 3, 2, 2, 2, 196, 199,
	3, 2, 2, 2, 197, 195, 3, 2, 2, 2, 197, 198, 3, 2, 2, 2, 198, 201, 3, 2,
	2, 2, 199, 197, 3, 2, 2, 2, 200, 192, 3, 2, 2, 2, 200, 201, 3, 2, 2, 2,
	201, 202, 3, 2, 2, 2, 202, 203, 7, 9, 2, 2, 203, 27, 3, 2, 2, 2, 204, 205,
	9, 4, 2, 2, 205, 29, 3, 2, 2, 2, 206, 207, 9, 4, 2, 2, 207, 31, 3, 2, 2,
	2, 208, 217, 7, 10, 2, 2, 209, 214, 5, 24, 13, 2, 210, 211, 7, 8, 2, 2,
	211, 213, 5, 24, 13, 2, 212, 210, 3, 2, 2, 2, 213, 216, 3, 2, 2, 2, 214,
	212, 3, 2, 2, 2, 214, 215, 3, 2, 2, 2, 215, 218, 3, 2, 2, 2, 216, 214,
	3, 2, 2, 2, 217, 209, 3, 2, 2, 2, 217, 218, 3, 2, 2, 2, 218, 219, 3, 2,
	2, 2, 219, 220, 7, 11, 2, 2, 220, 33, 3, 2, 2, 2, 40, 37, 40, 43, 46, 54,
	59, 62, 65, 71, 77, 80, 91, 95, 99, 102, 106, 109, 113, 116, 119, 124,
	129, 135, 139, 143, 148, 154, 158, 162, 166, 172, 176, 182, 185, 197, 200,
	214, 217,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'{'", "'}'", "'window'", "'='", "'('", "','", "')'", "'['", "']'",
	"", "';'", "'->'", "", "", "", "", "", "", "", "", "'\t'",
}
var symbolicNames = []string{
	"", "", "", "", "", "", "", "", "", "", "CASE", "EOP", "PIPE", "PIPE_NAME",
	"STRING", "NUMBER", "NAME", "COMMENT", "MULTILINE_COMMENT", "NEWLINE",
	"WHITESPACE", "TAB",
}

var ruleNames = []string{
	"script", "outputFork", "fork", "window", "multiinput", "input", "output",
	"transform", "subPipeline", "windowSubPipeline", "pipeline", "parameter",
	"transformParameters", "name", "pipelineName", "schedulingHints",
}
var decisionToDFA = make([]*antlr.DFA, len(deserializedATN.DecisionToState))

func init() {
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

type BitflowParser struct {
	*antlr.BaseParser
}

func NewBitflowParser(input antlr.TokenStream) *BitflowParser {
	this := new(BitflowParser)

	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "Bitflow.g4"

	return this
}

// BitflowParser tokens.
const (
	BitflowParserEOF               = antlr.TokenEOF
	BitflowParserT__0              = 1
	BitflowParserT__1              = 2
	BitflowParserT__2              = 3
	BitflowParserT__3              = 4
	BitflowParserT__4              = 5
	BitflowParserT__5              = 6
	BitflowParserT__6              = 7
	BitflowParserT__7              = 8
	BitflowParserT__8              = 9
	BitflowParserCASE              = 10
	BitflowParserEOP               = 11
	BitflowParserPIPE              = 12
	BitflowParserPIPE_NAME         = 13
	BitflowParserSTRING            = 14
	BitflowParserNUMBER            = 15
	BitflowParserNAME              = 16
	BitflowParserCOMMENT           = 17
	BitflowParserMULTILINE_COMMENT = 18
	BitflowParserNEWLINE           = 19
	BitflowParserWHITESPACE        = 20
	BitflowParserTAB               = 21
)

// BitflowParser rules.
const (
	BitflowParserRULE_script              = 0
	BitflowParserRULE_outputFork          = 1
	BitflowParserRULE_fork                = 2
	BitflowParserRULE_window              = 3
	BitflowParserRULE_multiinput          = 4
	BitflowParserRULE_input               = 5
	BitflowParserRULE_output              = 6
	BitflowParserRULE_transform           = 7
	BitflowParserRULE_subPipeline         = 8
	BitflowParserRULE_windowSubPipeline   = 9
	BitflowParserRULE_pipeline            = 10
	BitflowParserRULE_parameter           = 11
	BitflowParserRULE_transformParameters = 12
	BitflowParserRULE_name                = 13
	BitflowParserRULE_pipelineName        = 14
	BitflowParserRULE_schedulingHints     = 15
)

// IScriptContext is an interface to support dynamic dispatch.
type IScriptContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsScriptContext differentiates from other interfaces.
	IsScriptContext()
}

type ScriptContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyScriptContext() *ScriptContext {
	var p = new(ScriptContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_script
	return p
}

func (*ScriptContext) IsScriptContext() {}

func NewScriptContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ScriptContext {
	var p = new(ScriptContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_script

	return p
}

func (s *ScriptContext) GetParser() antlr.Parser { return s.parser }

func (s *ScriptContext) AllPipeline() []IPipelineContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IPipelineContext)(nil)).Elem())
	var tst = make([]IPipelineContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IPipelineContext)
		}
	}

	return tst
}

func (s *ScriptContext) Pipeline(i int) IPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IPipelineContext)
}

func (s *ScriptContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ScriptContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ScriptContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterScript(s)
	}
}

func (s *ScriptContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitScript(s)
	}
}

func (s *ScriptContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitScript(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Script() (localctx IScriptContext) {
	localctx = NewScriptContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, BitflowParserRULE_script)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(33)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserNAME))) != 0) {
		{
			p.SetState(32)
			p.Pipeline()
		}

		p.SetState(35)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}

	return localctx
}

// IOutputForkContext is an interface to support dynamic dispatch.
type IOutputForkContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsOutputForkContext differentiates from other interfaces.
	IsOutputForkContext()
}

type OutputForkContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOutputForkContext() *OutputForkContext {
	var p = new(OutputForkContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_outputFork
	return p
}

func (*OutputForkContext) IsOutputForkContext() {}

func NewOutputForkContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OutputForkContext {
	var p = new(OutputForkContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_outputFork

	return p
}

func (s *OutputForkContext) GetParser() antlr.Parser { return s.parser }

func (s *OutputForkContext) Name() INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *OutputForkContext) TransformParameters() ITransformParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITransformParametersContext)
}

func (s *OutputForkContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *OutputForkContext) AllOutput() []IOutputContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IOutputContext)(nil)).Elem())
	var tst = make([]IOutputContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IOutputContext)
		}
	}

	return tst
}

func (s *OutputForkContext) Output(i int) IOutputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOutputContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IOutputContext)
}

func (s *OutputForkContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OutputForkContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OutputForkContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterOutputFork(s)
	}
}

func (s *OutputForkContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitOutputFork(s)
	}
}

func (s *OutputForkContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitOutputFork(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) OutputFork() (localctx IOutputForkContext) {
	localctx = NewOutputForkContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, BitflowParserRULE_outputFork)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(38)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if ((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserNAME))) != 0 {
		{
			p.SetState(37)
			p.Name()
		}

	}
	p.SetState(41)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__4 {
		{
			p.SetState(40)
			p.TransformParameters()
		}

	}
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(43)
			p.SchedulingHints()
		}

	}
	{
		p.SetState(46)
		p.Match(BitflowParserT__0)
	}
	p.SetState(50)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserNAME))) != 0) {
		{
			p.SetState(47)
			p.Output()
		}
		{
			p.SetState(48)
			p.Match(BitflowParserEOP)
		}

		p.SetState(52)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(54)
		p.Match(BitflowParserT__1)
	}

	return localctx
}

// IForkContext is an interface to support dynamic dispatch.
type IForkContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsForkContext differentiates from other interfaces.
	IsForkContext()
}

type ForkContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyForkContext() *ForkContext {
	var p = new(ForkContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_fork
	return p
}

func (*ForkContext) IsForkContext() {}

func NewForkContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ForkContext {
	var p = new(ForkContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_fork

	return p
}

func (s *ForkContext) GetParser() antlr.Parser { return s.parser }

func (s *ForkContext) Name() INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *ForkContext) TransformParameters() ITransformParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITransformParametersContext)
}

func (s *ForkContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *ForkContext) AllSubPipeline() []ISubPipelineContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ISubPipelineContext)(nil)).Elem())
	var tst = make([]ISubPipelineContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ISubPipelineContext)
		}
	}

	return tst
}

func (s *ForkContext) SubPipeline(i int) ISubPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISubPipelineContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ISubPipelineContext)
}

func (s *ForkContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ForkContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ForkContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterFork(s)
	}
}

func (s *ForkContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitFork(s)
	}
}

func (s *ForkContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitFork(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Fork() (localctx IForkContext) {
	localctx = NewForkContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, BitflowParserRULE_fork)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if ((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserNAME))) != 0 {
		{
			p.SetState(56)
			p.Name()
		}

	}
	p.SetState(60)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__4 {
		{
			p.SetState(59)
			p.TransformParameters()
		}

	}
	p.SetState(63)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(62)
			p.SchedulingHints()
		}

	}
	{
		p.SetState(65)
		p.Match(BitflowParserT__0)
	}
	p.SetState(67)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserT__0)|(1<<BitflowParserT__2)|(1<<BitflowParserT__4)|(1<<BitflowParserT__7)|(1<<BitflowParserCASE)|(1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserNAME))) != 0) {
		{
			p.SetState(66)
			p.SubPipeline()
		}

		p.SetState(69)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(71)
		p.Match(BitflowParserT__1)
	}

	return localctx
}

// IWindowContext is an interface to support dynamic dispatch.
type IWindowContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsWindowContext differentiates from other interfaces.
	IsWindowContext()
}

type WindowContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWindowContext() *WindowContext {
	var p = new(WindowContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_window
	return p
}

func (*WindowContext) IsWindowContext() {}

func NewWindowContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WindowContext {
	var p = new(WindowContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_window

	return p
}

func (s *WindowContext) GetParser() antlr.Parser { return s.parser }

func (s *WindowContext) WindowSubPipeline() IWindowSubPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IWindowSubPipelineContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IWindowSubPipelineContext)
}

func (s *WindowContext) TransformParameters() ITransformParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITransformParametersContext)
}

func (s *WindowContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *WindowContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WindowContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WindowContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterWindow(s)
	}
}

func (s *WindowContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitWindow(s)
	}
}

func (s *WindowContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitWindow(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Window() (localctx IWindowContext) {
	localctx = NewWindowContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, BitflowParserRULE_window)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(73)
		p.Match(BitflowParserT__2)
	}
	p.SetState(75)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__4 {
		{
			p.SetState(74)
			p.TransformParameters()
		}

	}
	p.SetState(78)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(77)
			p.SchedulingHints()
		}

	}
	{
		p.SetState(80)
		p.Match(BitflowParserT__0)
	}
	{
		p.SetState(81)
		p.WindowSubPipeline()
	}
	{
		p.SetState(82)
		p.Match(BitflowParserT__1)
	}

	return localctx
}

// IMultiinputContext is an interface to support dynamic dispatch.
type IMultiinputContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsMultiinputContext differentiates from other interfaces.
	IsMultiinputContext()
}

type MultiinputContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiinputContext() *MultiinputContext {
	var p = new(MultiinputContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_multiinput
	return p
}

func (*MultiinputContext) IsMultiinputContext() {}

func NewMultiinputContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiinputContext {
	var p = new(MultiinputContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_multiinput

	return p
}

func (s *MultiinputContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiinputContext) AllInput() []IInputContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IInputContext)(nil)).Elem())
	var tst = make([]IInputContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IInputContext)
		}
	}

	return tst
}

func (s *MultiinputContext) Input(i int) IInputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IInputContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IInputContext)
}

func (s *MultiinputContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiinputContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiinputContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterMultiinput(s)
	}
}

func (s *MultiinputContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitMultiinput(s)
	}
}

func (s *MultiinputContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitMultiinput(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Multiinput() (localctx IMultiinputContext) {
	localctx = NewMultiinputContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, BitflowParserRULE_multiinput)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(84)
		p.Input()
	}
	p.SetState(89)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 11, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(85)
				p.Match(BitflowParserEOP)
			}
			{
				p.SetState(86)
				p.Input()
			}

		}
		p.SetState(91)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 11, p.GetParserRuleContext())
	}
	p.SetState(93)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOP {
		{
			p.SetState(92)
			p.Match(BitflowParserEOP)
		}

	}

	return localctx
}

// IInputContext is an interface to support dynamic dispatch.
type IInputContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsInputContext differentiates from other interfaces.
	IsInputContext()
}

type InputContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInputContext() *InputContext {
	var p = new(InputContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_input
	return p
}

func (*InputContext) IsInputContext() {}

func NewInputContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InputContext {
	var p = new(InputContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_input

	return p
}

func (s *InputContext) GetParser() antlr.Parser { return s.parser }

func (s *InputContext) Name() INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *InputContext) TransformParameters() ITransformParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITransformParametersContext)
}

func (s *InputContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *InputContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InputContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InputContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterInput(s)
	}
}

func (s *InputContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitInput(s)
	}
}

func (s *InputContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitInput(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Input() (localctx IInputContext) {
	localctx = NewInputContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, BitflowParserRULE_input)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(95)
		p.Name()
	}
	p.SetState(97)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__4 {
		{
			p.SetState(96)
			p.TransformParameters()
		}

	}
	p.SetState(100)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(99)
			p.SchedulingHints()
		}

	}

	return localctx
}

// IOutputContext is an interface to support dynamic dispatch.
type IOutputContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsOutputContext differentiates from other interfaces.
	IsOutputContext()
}

type OutputContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOutputContext() *OutputContext {
	var p = new(OutputContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_output
	return p
}

func (*OutputContext) IsOutputContext() {}

func NewOutputContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OutputContext {
	var p = new(OutputContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_output

	return p
}

func (s *OutputContext) GetParser() antlr.Parser { return s.parser }

func (s *OutputContext) Name() INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *OutputContext) TransformParameters() ITransformParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITransformParametersContext)
}

func (s *OutputContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *OutputContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OutputContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OutputContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterOutput(s)
	}
}

func (s *OutputContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitOutput(s)
	}
}

func (s *OutputContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitOutput(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Output() (localctx IOutputContext) {
	localctx = NewOutputContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, BitflowParserRULE_output)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(102)
		p.Name()
	}
	p.SetState(104)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__4 {
		{
			p.SetState(103)
			p.TransformParameters()
		}

	}
	p.SetState(107)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(106)
			p.SchedulingHints()
		}

	}

	return localctx
}

// ITransformContext is an interface to support dynamic dispatch.
type ITransformContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsTransformContext differentiates from other interfaces.
	IsTransformContext()
}

type TransformContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTransformContext() *TransformContext {
	var p = new(TransformContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_transform
	return p
}

func (*TransformContext) IsTransformContext() {}

func NewTransformContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TransformContext {
	var p = new(TransformContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_transform

	return p
}

func (s *TransformContext) GetParser() antlr.Parser { return s.parser }

func (s *TransformContext) Name() INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *TransformContext) TransformParameters() ITransformParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITransformParametersContext)
}

func (s *TransformContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *TransformContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TransformContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TransformContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterTransform(s)
	}
}

func (s *TransformContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitTransform(s)
	}
}

func (s *TransformContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitTransform(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Transform() (localctx ITransformContext) {
	localctx = NewTransformContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, BitflowParserRULE_transform)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(109)
		p.Name()
	}
	p.SetState(111)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 17, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(110)
			p.TransformParameters()
		}

	}
	p.SetState(114)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 18, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(113)
			p.SchedulingHints()
		}

	}

	return localctx
}

// ISubPipelineContext is an interface to support dynamic dispatch.
type ISubPipelineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSubPipelineContext differentiates from other interfaces.
	IsSubPipelineContext()
}

type SubPipelineContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySubPipelineContext() *SubPipelineContext {
	var p = new(SubPipelineContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_subPipeline
	return p
}

func (*SubPipelineContext) IsSubPipelineContext() {}

func NewSubPipelineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SubPipelineContext {
	var p = new(SubPipelineContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_subPipeline

	return p
}

func (s *SubPipelineContext) GetParser() antlr.Parser { return s.parser }

func (s *SubPipelineContext) AllTransform() []ITransformContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ITransformContext)(nil)).Elem())
	var tst = make([]ITransformContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ITransformContext)
		}
	}

	return tst
}

func (s *SubPipelineContext) Transform(i int) ITransformContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ITransformContext)
}

func (s *SubPipelineContext) AllFork() []IForkContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IForkContext)(nil)).Elem())
	var tst = make([]IForkContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IForkContext)
		}
	}

	return tst
}

func (s *SubPipelineContext) Fork(i int) IForkContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IForkContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IForkContext)
}

func (s *SubPipelineContext) AllWindow() []IWindowContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IWindowContext)(nil)).Elem())
	var tst = make([]IWindowContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IWindowContext)
		}
	}

	return tst
}

func (s *SubPipelineContext) Window(i int) IWindowContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IWindowContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IWindowContext)
}

func (s *SubPipelineContext) PipelineName() IPipelineNameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineNameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IPipelineNameContext)
}

func (s *SubPipelineContext) AllPIPE() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserPIPE)
}

func (s *SubPipelineContext) PIPE(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserPIPE, i)
}

func (s *SubPipelineContext) EOP() antlr.TerminalNode {
	return s.GetToken(BitflowParserEOP, 0)
}

func (s *SubPipelineContext) EOF() antlr.TerminalNode {
	return s.GetToken(BitflowParserEOF, 0)
}

func (s *SubPipelineContext) CASE() antlr.TerminalNode {
	return s.GetToken(BitflowParserCASE, 0)
}

func (s *SubPipelineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SubPipelineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SubPipelineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterSubPipeline(s)
	}
}

func (s *SubPipelineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitSubPipeline(s)
	}
}

func (s *SubPipelineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitSubPipeline(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) SubPipeline() (localctx ISubPipelineContext) {
	localctx = NewSubPipelineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, BitflowParserRULE_subPipeline)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(122)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 20, p.GetParserRuleContext()) == 1 {
		p.SetState(117)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		if _la == BitflowParserCASE {
			{
				p.SetState(116)
				p.Match(BitflowParserCASE)
			}

		}
		{
			p.SetState(119)
			p.PipelineName()
		}
		{
			p.SetState(120)
			p.Match(BitflowParserPIPE)
		}

	}
	p.SetState(127)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 21, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(124)
			p.Transform()
		}

	case 2:
		{
			p.SetState(125)
			p.Fork()
		}

	case 3:
		{
			p.SetState(126)
			p.Window()
		}

	}
	p.SetState(137)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == BitflowParserPIPE {
		{
			p.SetState(129)
			p.Match(BitflowParserPIPE)
		}
		p.SetState(133)
		p.GetErrorHandler().Sync(p)
		switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 22, p.GetParserRuleContext()) {
		case 1:
			{
				p.SetState(130)
				p.Transform()
			}

		case 2:
			{
				p.SetState(131)
				p.Fork()
			}

		case 3:
			{
				p.SetState(132)
				p.Window()
			}

		}

		p.SetState(139)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(141)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOF || _la == BitflowParserEOP {
		{
			p.SetState(140)
			_la = p.GetTokenStream().LA(1)

			if !(_la == BitflowParserEOF || _la == BitflowParserEOP) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	}

	return localctx
}

// IWindowSubPipelineContext is an interface to support dynamic dispatch.
type IWindowSubPipelineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsWindowSubPipelineContext differentiates from other interfaces.
	IsWindowSubPipelineContext()
}

type WindowSubPipelineContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWindowSubPipelineContext() *WindowSubPipelineContext {
	var p = new(WindowSubPipelineContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_windowSubPipeline
	return p
}

func (*WindowSubPipelineContext) IsWindowSubPipelineContext() {}

func NewWindowSubPipelineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WindowSubPipelineContext {
	var p = new(WindowSubPipelineContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_windowSubPipeline

	return p
}

func (s *WindowSubPipelineContext) GetParser() antlr.Parser { return s.parser }

func (s *WindowSubPipelineContext) AllTransform() []ITransformContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ITransformContext)(nil)).Elem())
	var tst = make([]ITransformContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ITransformContext)
		}
	}

	return tst
}

func (s *WindowSubPipelineContext) Transform(i int) ITransformContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ITransformContext)
}

func (s *WindowSubPipelineContext) AllFork() []IForkContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IForkContext)(nil)).Elem())
	var tst = make([]IForkContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IForkContext)
		}
	}

	return tst
}

func (s *WindowSubPipelineContext) Fork(i int) IForkContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IForkContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IForkContext)
}

func (s *WindowSubPipelineContext) AllWindow() []IWindowContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IWindowContext)(nil)).Elem())
	var tst = make([]IWindowContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IWindowContext)
		}
	}

	return tst
}

func (s *WindowSubPipelineContext) Window(i int) IWindowContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IWindowContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IWindowContext)
}

func (s *WindowSubPipelineContext) AllPIPE() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserPIPE)
}

func (s *WindowSubPipelineContext) PIPE(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserPIPE, i)
}

func (s *WindowSubPipelineContext) EOP() antlr.TerminalNode {
	return s.GetToken(BitflowParserEOP, 0)
}

func (s *WindowSubPipelineContext) EOF() antlr.TerminalNode {
	return s.GetToken(BitflowParserEOF, 0)
}

func (s *WindowSubPipelineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WindowSubPipelineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WindowSubPipelineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterWindowSubPipeline(s)
	}
}

func (s *WindowSubPipelineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitWindowSubPipeline(s)
	}
}

func (s *WindowSubPipelineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitWindowSubPipeline(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) WindowSubPipeline() (localctx IWindowSubPipelineContext) {
	localctx = NewWindowSubPipelineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, BitflowParserRULE_windowSubPipeline)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	p.SetState(146)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 25, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(143)
			p.Transform()
		}

	case 2:
		{
			p.SetState(144)
			p.Fork()
		}

	case 3:
		{
			p.SetState(145)
			p.Window()
		}

	}
	p.SetState(156)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == BitflowParserPIPE {
		{
			p.SetState(148)
			p.Match(BitflowParserPIPE)
		}
		p.SetState(152)
		p.GetErrorHandler().Sync(p)
		switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 26, p.GetParserRuleContext()) {
		case 1:
			{
				p.SetState(149)
				p.Transform()
			}

		case 2:
			{
				p.SetState(150)
				p.Fork()
			}

		case 3:
			{
				p.SetState(151)
				p.Window()
			}

		}

		p.SetState(158)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(160)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOF || _la == BitflowParserEOP {
		{
			p.SetState(159)
			_la = p.GetTokenStream().LA(1)

			if !(_la == BitflowParserEOF || _la == BitflowParserEOP) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	}

	return localctx
}

// IPipelineContext is an interface to support dynamic dispatch.
type IPipelineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsPipelineContext differentiates from other interfaces.
	IsPipelineContext()
}

type PipelineContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPipelineContext() *PipelineContext {
	var p = new(PipelineContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_pipeline
	return p
}

func (*PipelineContext) IsPipelineContext() {}

func NewPipelineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PipelineContext {
	var p = new(PipelineContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_pipeline

	return p
}

func (s *PipelineContext) GetParser() antlr.Parser { return s.parser }

func (s *PipelineContext) AllPIPE() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserPIPE)
}

func (s *PipelineContext) PIPE(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserPIPE, i)
}

func (s *PipelineContext) Multiinput() IMultiinputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMultiinputContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMultiinputContext)
}

func (s *PipelineContext) Input() IInputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IInputContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IInputContext)
}

func (s *PipelineContext) Output() IOutputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOutputContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOutputContext)
}

func (s *PipelineContext) OutputFork() IOutputForkContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOutputForkContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOutputForkContext)
}

func (s *PipelineContext) EOP() antlr.TerminalNode {
	return s.GetToken(BitflowParserEOP, 0)
}

func (s *PipelineContext) EOF() antlr.TerminalNode {
	return s.GetToken(BitflowParserEOF, 0)
}

func (s *PipelineContext) AllTransform() []ITransformContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ITransformContext)(nil)).Elem())
	var tst = make([]ITransformContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ITransformContext)
		}
	}

	return tst
}

func (s *PipelineContext) Transform(i int) ITransformContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ITransformContext)
}

func (s *PipelineContext) AllFork() []IForkContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IForkContext)(nil)).Elem())
	var tst = make([]IForkContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IForkContext)
		}
	}

	return tst
}

func (s *PipelineContext) Fork(i int) IForkContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IForkContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IForkContext)
}

func (s *PipelineContext) AllWindow() []IWindowContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IWindowContext)(nil)).Elem())
	var tst = make([]IWindowContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IWindowContext)
		}
	}

	return tst
}

func (s *PipelineContext) Window(i int) IWindowContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IWindowContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IWindowContext)
}

func (s *PipelineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PipelineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PipelineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterPipeline(s)
	}
}

func (s *PipelineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitPipeline(s)
	}
}

func (s *PipelineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitPipeline(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Pipeline() (localctx IPipelineContext) {
	localctx = NewPipelineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, BitflowParserRULE_pipeline)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(164)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 29, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(162)
			p.Multiinput()
		}

	case 2:
		{
			p.SetState(163)
			p.Input()
		}

	}
	p.SetState(174)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 31, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(166)
				p.Match(BitflowParserPIPE)
			}
			p.SetState(170)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 30, p.GetParserRuleContext()) {
			case 1:
				{
					p.SetState(167)
					p.Transform()
				}

			case 2:
				{
					p.SetState(168)
					p.Fork()
				}

			case 3:
				{
					p.SetState(169)
					p.Window()
				}

			}

		}
		p.SetState(176)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 31, p.GetParserRuleContext())
	}
	{
		p.SetState(177)
		p.Match(BitflowParserPIPE)
	}
	p.SetState(180)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 32, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(178)
			p.Output()
		}

	case 2:
		{
			p.SetState(179)
			p.OutputFork()
		}

	}
	p.SetState(183)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 33, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(182)
			_la = p.GetTokenStream().LA(1)

			if !(_la == BitflowParserEOF || _la == BitflowParserEOP) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	}

	return localctx
}

// IParameterContext is an interface to support dynamic dispatch.
type IParameterContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsParameterContext differentiates from other interfaces.
	IsParameterContext()
}

type ParameterContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyParameterContext() *ParameterContext {
	var p = new(ParameterContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_parameter
	return p
}

func (*ParameterContext) IsParameterContext() {}

func NewParameterContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ParameterContext {
	var p = new(ParameterContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_parameter

	return p
}

func (s *ParameterContext) GetParser() antlr.Parser { return s.parser }

func (s *ParameterContext) NAME() antlr.TerminalNode {
	return s.GetToken(BitflowParserNAME, 0)
}

func (s *ParameterContext) STRING() antlr.TerminalNode {
	return s.GetToken(BitflowParserSTRING, 0)
}

func (s *ParameterContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(BitflowParserNUMBER, 0)
}

func (s *ParameterContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParameterContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ParameterContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterParameter(s)
	}
}

func (s *ParameterContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitParameter(s)
	}
}

func (s *ParameterContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitParameter(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Parameter() (localctx IParameterContext) {
	localctx = NewParameterContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, BitflowParserRULE_parameter)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(185)
		p.Match(BitflowParserNAME)
	}
	{
		p.SetState(186)
		p.Match(BitflowParserT__3)
	}
	{
		p.SetState(187)
		_la = p.GetTokenStream().LA(1)

		if !(_la == BitflowParserSTRING || _la == BitflowParserNUMBER) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// ITransformParametersContext is an interface to support dynamic dispatch.
type ITransformParametersContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsTransformParametersContext differentiates from other interfaces.
	IsTransformParametersContext()
}

type TransformParametersContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTransformParametersContext() *TransformParametersContext {
	var p = new(TransformParametersContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_transformParameters
	return p
}

func (*TransformParametersContext) IsTransformParametersContext() {}

func NewTransformParametersContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TransformParametersContext {
	var p = new(TransformParametersContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_transformParameters

	return p
}

func (s *TransformParametersContext) GetParser() antlr.Parser { return s.parser }

func (s *TransformParametersContext) AllParameter() []IParameterContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IParameterContext)(nil)).Elem())
	var tst = make([]IParameterContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IParameterContext)
		}
	}

	return tst
}

func (s *TransformParametersContext) Parameter(i int) IParameterContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParameterContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IParameterContext)
}

func (s *TransformParametersContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TransformParametersContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TransformParametersContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterTransformParameters(s)
	}
}

func (s *TransformParametersContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitTransformParameters(s)
	}
}

func (s *TransformParametersContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitTransformParameters(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) TransformParameters() (localctx ITransformParametersContext) {
	localctx = NewTransformParametersContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, BitflowParserRULE_transformParameters)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(189)
		p.Match(BitflowParserT__4)
	}
	p.SetState(198)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserNAME {
		{
			p.SetState(190)
			p.Parameter()
		}
		p.SetState(195)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		for _la == BitflowParserT__5 {
			{
				p.SetState(191)
				p.Match(BitflowParserT__5)
			}
			{
				p.SetState(192)
				p.Parameter()
			}

			p.SetState(197)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)
		}

	}
	{
		p.SetState(200)
		p.Match(BitflowParserT__6)
	}

	return localctx
}

// INameContext is an interface to support dynamic dispatch.
type INameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsNameContext differentiates from other interfaces.
	IsNameContext()
}

type NameContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNameContext() *NameContext {
	var p = new(NameContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_name
	return p
}

func (*NameContext) IsNameContext() {}

func NewNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NameContext {
	var p = new(NameContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_name

	return p
}

func (s *NameContext) GetParser() antlr.Parser { return s.parser }

func (s *NameContext) STRING() antlr.TerminalNode {
	return s.GetToken(BitflowParserSTRING, 0)
}

func (s *NameContext) NAME() antlr.TerminalNode {
	return s.GetToken(BitflowParserNAME, 0)
}

func (s *NameContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(BitflowParserNUMBER, 0)
}

func (s *NameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterName(s)
	}
}

func (s *NameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitName(s)
	}
}

func (s *NameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitName(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Name() (localctx INameContext) {
	localctx = NewNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, BitflowParserRULE_name)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(202)
		_la = p.GetTokenStream().LA(1)

		if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserNAME))) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// IPipelineNameContext is an interface to support dynamic dispatch.
type IPipelineNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsPipelineNameContext differentiates from other interfaces.
	IsPipelineNameContext()
}

type PipelineNameContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPipelineNameContext() *PipelineNameContext {
	var p = new(PipelineNameContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_pipelineName
	return p
}

func (*PipelineNameContext) IsPipelineNameContext() {}

func NewPipelineNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PipelineNameContext {
	var p = new(PipelineNameContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_pipelineName

	return p
}

func (s *PipelineNameContext) GetParser() antlr.Parser { return s.parser }

func (s *PipelineNameContext) STRING() antlr.TerminalNode {
	return s.GetToken(BitflowParserSTRING, 0)
}

func (s *PipelineNameContext) NAME() antlr.TerminalNode {
	return s.GetToken(BitflowParserNAME, 0)
}

func (s *PipelineNameContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(BitflowParserNUMBER, 0)
}

func (s *PipelineNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PipelineNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PipelineNameContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterPipelineName(s)
	}
}

func (s *PipelineNameContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitPipelineName(s)
	}
}

func (s *PipelineNameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitPipelineName(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) PipelineName() (localctx IPipelineNameContext) {
	localctx = NewPipelineNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, BitflowParserRULE_pipelineName)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(204)
		_la = p.GetTokenStream().LA(1)

		if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserNAME))) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// ISchedulingHintsContext is an interface to support dynamic dispatch.
type ISchedulingHintsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSchedulingHintsContext differentiates from other interfaces.
	IsSchedulingHintsContext()
}

type SchedulingHintsContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySchedulingHintsContext() *SchedulingHintsContext {
	var p = new(SchedulingHintsContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_schedulingHints
	return p
}

func (*SchedulingHintsContext) IsSchedulingHintsContext() {}

func NewSchedulingHintsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SchedulingHintsContext {
	var p = new(SchedulingHintsContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_schedulingHints

	return p
}

func (s *SchedulingHintsContext) GetParser() antlr.Parser { return s.parser }

func (s *SchedulingHintsContext) AllParameter() []IParameterContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IParameterContext)(nil)).Elem())
	var tst = make([]IParameterContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IParameterContext)
		}
	}

	return tst
}

func (s *SchedulingHintsContext) Parameter(i int) IParameterContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParameterContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IParameterContext)
}

func (s *SchedulingHintsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SchedulingHintsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SchedulingHintsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterSchedulingHints(s)
	}
}

func (s *SchedulingHintsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitSchedulingHints(s)
	}
}

func (s *SchedulingHintsContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitSchedulingHints(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) SchedulingHints() (localctx ISchedulingHintsContext) {
	localctx = NewSchedulingHintsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, BitflowParserRULE_schedulingHints)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(206)
		p.Match(BitflowParserT__7)
	}
	p.SetState(215)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserNAME {
		{
			p.SetState(207)
			p.Parameter()
		}
		p.SetState(212)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		for _la == BitflowParserT__5 {
			{
				p.SetState(208)
				p.Match(BitflowParserT__5)
			}
			{
				p.SetState(209)
				p.Parameter()
			}

			p.SetState(214)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)
		}

	}
	{
		p.SetState(217)
		p.Match(BitflowParserT__8)
	}

	return localctx
}
