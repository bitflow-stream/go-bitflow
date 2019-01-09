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
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 21, 199,
	4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7,
	4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4, 13,
	9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 4, 18, 9,
	18, 4, 19, 9, 19, 4, 20, 9, 20, 4, 21, 9, 21, 3, 2, 3, 2, 3, 2, 3, 3, 6,
	3, 47, 10, 3, 13, 3, 14, 3, 48, 3, 3, 5, 3, 52, 10, 3, 3, 4, 3, 4, 5, 4,
	56, 10, 4, 3, 5, 3, 5, 3, 6, 3, 6, 3, 7, 3, 7, 3, 7, 3, 7, 3, 8, 3, 8,
	3, 8, 3, 8, 7, 8, 70, 10, 8, 12, 8, 14, 8, 73, 11, 8, 5, 8, 75, 10, 8,
	3, 8, 3, 8, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9, 5, 9, 84, 10, 9, 3, 9, 3, 9,
	7, 9, 88, 10, 9, 12, 9, 14, 9, 91, 11, 9, 3, 10, 3, 10, 3, 10, 7, 10, 96,
	10, 10, 12, 10, 14, 10, 99, 11, 10, 3, 10, 5, 10, 102, 10, 10, 3, 11, 3,
	11, 3, 11, 3, 11, 3, 11, 5, 11, 109, 10, 11, 3, 12, 3, 12, 3, 12, 5, 12,
	114, 10, 12, 3, 13, 3, 13, 3, 13, 5, 13, 119, 10, 13, 3, 13, 3, 13, 3,
	13, 3, 13, 7, 13, 125, 10, 13, 12, 13, 14, 13, 128, 11, 13, 3, 13, 5, 13,
	131, 10, 13, 3, 13, 3, 13, 3, 14, 6, 14, 136, 10, 14, 13, 14, 14, 14, 137,
	3, 14, 3, 14, 3, 14, 3, 15, 3, 15, 3, 15, 7, 15, 146, 10, 15, 12, 15, 14,
	15, 149, 11, 15, 3, 16, 3, 16, 3, 16, 3, 16, 7, 16, 155, 10, 16, 12, 16,
	14, 16, 158, 11, 16, 3, 16, 5, 16, 161, 10, 16, 3, 16, 3, 16, 3, 17, 3,
	17, 3, 18, 3, 18, 3, 18, 5, 18, 170, 10, 18, 3, 18, 3, 18, 3, 18, 3, 18,
	3, 19, 3, 19, 3, 19, 7, 19, 179, 10, 19, 12, 19, 14, 19, 182, 11, 19, 3,
	20, 3, 20, 3, 20, 3, 20, 7, 20, 188, 10, 20, 12, 20, 14, 20, 191, 11, 20,
	5, 20, 193, 10, 20, 3, 20, 3, 20, 3, 21, 3, 21, 3, 21, 2, 2, 22, 2, 4,
	6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 2,
	4, 3, 2, 14, 17, 3, 2, 14, 16, 2, 203, 2, 42, 3, 2, 2, 2, 4, 46, 3, 2,
	2, 2, 6, 53, 3, 2, 2, 2, 8, 57, 3, 2, 2, 2, 10, 59, 3, 2, 2, 2, 12, 61,
	3, 2, 2, 2, 14, 65, 3, 2, 2, 2, 16, 83, 3, 2, 2, 2, 18, 92, 3, 2, 2, 2,
	20, 108, 3, 2, 2, 2, 22, 110, 3, 2, 2, 2, 24, 115, 3, 2, 2, 2, 26, 135,
	3, 2, 2, 2, 28, 142, 3, 2, 2, 2, 30, 150, 3, 2, 2, 2, 32, 164, 3, 2, 2,
	2, 34, 166, 3, 2, 2, 2, 36, 175, 3, 2, 2, 2, 38, 183, 3, 2, 2, 2, 40, 196,
	3, 2, 2, 2, 42, 43, 5, 18, 10, 2, 43, 44, 7, 2, 2, 3, 44, 3, 3, 2, 2, 2,
	45, 47, 5, 8, 5, 2, 46, 45, 3, 2, 2, 2, 47, 48, 3, 2, 2, 2, 48, 46, 3,
	2, 2, 2, 48, 49, 3, 2, 2, 2, 49, 51, 3, 2, 2, 2, 50, 52, 5, 38, 20, 2,
	51, 50, 3, 2, 2, 2, 51, 52, 3, 2, 2, 2, 52, 5, 3, 2, 2, 2, 53, 55, 5, 8,
	5, 2, 54, 56, 5, 38, 20, 2, 55, 54, 3, 2, 2, 2, 55, 56, 3, 2, 2, 2, 56,
	7, 3, 2, 2, 2, 57, 58, 9, 2, 2, 2, 58, 9, 3, 2, 2, 2, 59, 60, 9, 3, 2,
	2, 60, 11, 3, 2, 2, 2, 61, 62, 5, 8, 5, 2, 62, 63, 7, 3, 2, 2, 63, 64,
	5, 10, 6, 2, 64, 13, 3, 2, 2, 2, 65, 74, 7, 4, 2, 2, 66, 71, 5, 12, 7,
	2, 67, 68, 7, 5, 2, 2, 68, 70, 5, 12, 7, 2, 69, 67, 3, 2, 2, 2, 70, 73,
	3, 2, 2, 2, 71, 69, 3, 2, 2, 2, 71, 72, 3, 2, 2, 2, 72, 75, 3, 2, 2, 2,
	73, 71, 3, 2, 2, 2, 74, 66, 3, 2, 2, 2, 74, 75, 3, 2, 2, 2, 75, 76, 3,
	2, 2, 2, 76, 77, 7, 6, 2, 2, 77, 15, 3, 2, 2, 2, 78, 84, 5, 4, 3, 2, 79,
	80, 7, 7, 2, 2, 80, 81, 5, 18, 10, 2, 81, 82, 7, 8, 2, 2, 82, 84, 3, 2,
	2, 2, 83, 78, 3, 2, 2, 2, 83, 79, 3, 2, 2, 2, 84, 89, 3, 2, 2, 2, 85, 86,
	7, 13, 2, 2, 86, 88, 5, 20, 11, 2, 87, 85, 3, 2, 2, 2, 88, 91, 3, 2, 2,
	2, 89, 87, 3, 2, 2, 2, 89, 90, 3, 2, 2, 2, 90, 17, 3, 2, 2, 2, 91, 89,
	3, 2, 2, 2, 92, 97, 5, 16, 9, 2, 93, 94, 7, 12, 2, 2, 94, 96, 5, 16, 9,
	2, 95, 93, 3, 2, 2, 2, 96, 99, 3, 2, 2, 2, 97, 95, 3, 2, 2, 2, 97, 98,
	3, 2, 2, 2, 98, 101, 3, 2, 2, 2, 99, 97, 3, 2, 2, 2, 100, 102, 7, 12, 2,
	2, 101, 100, 3, 2, 2, 2, 101, 102, 3, 2, 2, 2, 102, 19, 3, 2, 2, 2, 103,
	109, 5, 22, 12, 2, 104, 109, 5, 24, 13, 2, 105, 109, 5, 30, 16, 2, 106,
	109, 5, 34, 18, 2, 107, 109, 5, 6, 4, 2, 108, 103, 3, 2, 2, 2, 108, 104,
	3, 2, 2, 2, 108, 105, 3, 2, 2, 2, 108, 106, 3, 2, 2, 2, 108, 107, 3, 2,
	2, 2, 109, 21, 3, 2, 2, 2, 110, 111, 5, 8, 5, 2, 111, 113, 5, 14, 8, 2,
	112, 114, 5, 38, 20, 2, 113, 112, 3, 2, 2, 2, 113, 114, 3, 2, 2, 2, 114,
	23, 3, 2, 2, 2, 115, 116, 5, 8, 5, 2, 116, 118, 5, 14, 8, 2, 117, 119,
	5, 38, 20, 2, 118, 117, 3, 2, 2, 2, 118, 119, 3, 2, 2, 2, 119, 120, 3,
	2, 2, 2, 120, 121, 7, 7, 2, 2, 121, 126, 5, 26, 14, 2, 122, 123, 7, 12,
	2, 2, 123, 125, 5, 26, 14, 2, 124, 122, 3, 2, 2, 2, 125, 128, 3, 2, 2,
	2, 126, 124, 3, 2, 2, 2, 126, 127, 3, 2, 2, 2, 127, 130, 3, 2, 2, 2, 128,
	126, 3, 2, 2, 2, 129, 131, 7, 12, 2, 2, 130, 129, 3, 2, 2, 2, 130, 131,
	3, 2, 2, 2, 131, 132, 3, 2, 2, 2, 132, 133, 7, 8, 2, 2, 133, 25, 3, 2,
	2, 2, 134, 136, 5, 8, 5, 2, 135, 134, 3, 2, 2, 2, 136, 137, 3, 2, 2, 2,
	137, 135, 3, 2, 2, 2, 137, 138, 3, 2, 2, 2, 138, 139, 3, 2, 2, 2, 139,
	140, 7, 13, 2, 2, 140, 141, 5, 28, 15, 2, 141, 27, 3, 2, 2, 2, 142, 147,
	5, 20, 11, 2, 143, 144, 7, 13, 2, 2, 144, 146, 5, 20, 11, 2, 145, 143,
	3, 2, 2, 2, 146, 149, 3, 2, 2, 2, 147, 145, 3, 2, 2, 2, 147, 148, 3, 2,
	2, 2, 148, 29, 3, 2, 2, 2, 149, 147, 3, 2, 2, 2, 150, 151, 7, 7, 2, 2,
	151, 156, 5, 32, 17, 2, 152, 153, 7, 12, 2, 2, 153, 155, 5, 32, 17, 2,
	154, 152, 3, 2, 2, 2, 155, 158, 3, 2, 2, 2, 156, 154, 3, 2, 2, 2, 156,
	157, 3, 2, 2, 2, 157, 160, 3, 2, 2, 2, 158, 156, 3, 2, 2, 2, 159, 161,
	7, 12, 2, 2, 160, 159, 3, 2, 2, 2, 160, 161, 3, 2, 2, 2, 161, 162, 3, 2,
	2, 2, 162, 163, 7, 8, 2, 2, 163, 31, 3, 2, 2, 2, 164, 165, 5, 28, 15, 2,
	165, 33, 3, 2, 2, 2, 166, 167, 7, 9, 2, 2, 167, 169, 5, 14, 8, 2, 168,
	170, 5, 38, 20, 2, 169, 168, 3, 2, 2, 2, 169, 170, 3, 2, 2, 2, 170, 171,
	3, 2, 2, 2, 171, 172, 7, 7, 2, 2, 172, 173, 5, 36, 19, 2, 173, 174, 7,
	8, 2, 2, 174, 35, 3, 2, 2, 2, 175, 180, 5, 22, 12, 2, 176, 177, 7, 13,
	2, 2, 177, 179, 5, 22, 12, 2, 178, 176, 3, 2, 2, 2, 179, 182, 3, 2, 2,
	2, 180, 178, 3, 2, 2, 2, 180, 181, 3, 2, 2, 2, 181, 37, 3, 2, 2, 2, 182,
	180, 3, 2, 2, 2, 183, 192, 7, 10, 2, 2, 184, 189, 5, 40, 21, 2, 185, 186,
	7, 5, 2, 2, 186, 188, 5, 40, 21, 2, 187, 185, 3, 2, 2, 2, 188, 191, 3,
	2, 2, 2, 189, 187, 3, 2, 2, 2, 189, 190, 3, 2, 2, 2, 190, 193, 3, 2, 2,
	2, 191, 189, 3, 2, 2, 2, 192, 184, 3, 2, 2, 2, 192, 193, 3, 2, 2, 2, 193,
	194, 3, 2, 2, 2, 194, 195, 7, 11, 2, 2, 195, 39, 3, 2, 2, 2, 196, 197,
	5, 12, 7, 2, 197, 41, 3, 2, 2, 2, 24, 48, 51, 55, 71, 74, 83, 89, 97, 101,
	108, 113, 118, 126, 130, 137, 147, 156, 160, 169, 180, 189, 192,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'='", "'('", "','", "')'", "'{'", "'}'", "'window'", "'['", "']'",
	"';'", "'->'", "", "", "", "", "", "", "", "'\t'",
}
var symbolicNames = []string{
	"", "", "", "", "", "", "", "", "", "", "EOP", "NEXT", "STRING", "NUMBER",
	"BOOL", "IDENTIFIER", "COMMENT", "NEWLINE", "WHITESPACE", "TAB",
}

var ruleNames = []string{
	"script", "input", "output", "name", "val", "parameter", "transformParameters",
	"pipeline", "multiInputPipeline", "pipelineElement", "transform", "fork",
	"namedSubPipeline", "subPipeline", "multiplexFork", "multiplexSubPipeline",
	"window", "windowPipeline", "schedulingHints", "schedulingParameter",
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
	BitflowParserEOF        = antlr.TokenEOF
	BitflowParserT__0       = 1
	BitflowParserT__1       = 2
	BitflowParserT__2       = 3
	BitflowParserT__3       = 4
	BitflowParserT__4       = 5
	BitflowParserT__5       = 6
	BitflowParserT__6       = 7
	BitflowParserT__7       = 8
	BitflowParserT__8       = 9
	BitflowParserEOP        = 10
	BitflowParserNEXT       = 11
	BitflowParserSTRING     = 12
	BitflowParserNUMBER     = 13
	BitflowParserBOOL       = 14
	BitflowParserIDENTIFIER = 15
	BitflowParserCOMMENT    = 16
	BitflowParserNEWLINE    = 17
	BitflowParserWHITESPACE = 18
	BitflowParserTAB        = 19
)

// BitflowParser rules.
const (
	BitflowParserRULE_script               = 0
	BitflowParserRULE_input                = 1
	BitflowParserRULE_output               = 2
	BitflowParserRULE_name                 = 3
	BitflowParserRULE_val                  = 4
	BitflowParserRULE_parameter            = 5
	BitflowParserRULE_transformParameters  = 6
	BitflowParserRULE_pipeline             = 7
	BitflowParserRULE_multiInputPipeline   = 8
	BitflowParserRULE_pipelineElement      = 9
	BitflowParserRULE_transform            = 10
	BitflowParserRULE_fork                 = 11
	BitflowParserRULE_namedSubPipeline     = 12
	BitflowParserRULE_subPipeline          = 13
	BitflowParserRULE_multiplexFork        = 14
	BitflowParserRULE_multiplexSubPipeline = 15
	BitflowParserRULE_window               = 16
	BitflowParserRULE_windowPipeline       = 17
	BitflowParserRULE_schedulingHints      = 18
	BitflowParserRULE_schedulingParameter  = 19
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

func (s *ScriptContext) MultiInputPipeline() IMultiInputPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMultiInputPipelineContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMultiInputPipelineContext)
}

func (s *ScriptContext) EOF() antlr.TerminalNode {
	return s.GetToken(BitflowParserEOF, 0)
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
		p.SetState(40)
		p.MultiInputPipeline()
	}
	{
		p.SetState(41)
		p.Match(BitflowParserEOF)
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

func (s *InputContext) AllName() []INameContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*INameContext)(nil)).Elem())
	var tst = make([]INameContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(INameContext)
		}
	}

	return tst
}

func (s *InputContext) Name(i int) INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(INameContext)
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
	p.EnterRule(localctx, 2, BitflowParserRULE_input)
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
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserBOOL)|(1<<BitflowParserIDENTIFIER))) != 0) {
		{
			p.SetState(43)
			p.Name()
		}

		p.SetState(46)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(49)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(48)
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
	p.EnterRule(localctx, 4, BitflowParserRULE_output)
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
		p.SetState(51)
		p.Name()
	}
	p.SetState(53)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(52)
			p.SchedulingHints()
		}

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

func (s *NameContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(BitflowParserIDENTIFIER, 0)
}

func (s *NameContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(BitflowParserNUMBER, 0)
}

func (s *NameContext) BOOL() antlr.TerminalNode {
	return s.GetToken(BitflowParserBOOL, 0)
}

func (s *NameContext) STRING() antlr.TerminalNode {
	return s.GetToken(BitflowParserSTRING, 0)
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
	p.EnterRule(localctx, 6, BitflowParserRULE_name)
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
		p.SetState(55)
		_la = p.GetTokenStream().LA(1)

		if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserBOOL)|(1<<BitflowParserIDENTIFIER))) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

// IValContext is an interface to support dynamic dispatch.
type IValContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsValContext differentiates from other interfaces.
	IsValContext()
}

type ValContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValContext() *ValContext {
	var p = new(ValContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_val
	return p
}

func (*ValContext) IsValContext() {}

func NewValContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValContext {
	var p = new(ValContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_val

	return p
}

func (s *ValContext) GetParser() antlr.Parser { return s.parser }

func (s *ValContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(BitflowParserNUMBER, 0)
}

func (s *ValContext) BOOL() antlr.TerminalNode {
	return s.GetToken(BitflowParserBOOL, 0)
}

func (s *ValContext) STRING() antlr.TerminalNode {
	return s.GetToken(BitflowParserSTRING, 0)
}

func (s *ValContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ValContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterVal(s)
	}
}

func (s *ValContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitVal(s)
	}
}

func (s *ValContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitVal(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Val() (localctx IValContext) {
	localctx = NewValContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, BitflowParserRULE_val)
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
		p.SetState(57)
		_la = p.GetTokenStream().LA(1)

		if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserBOOL))) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
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

func (s *ParameterContext) Name() INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *ParameterContext) Val() IValContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IValContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IValContext)
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
	p.EnterRule(localctx, 10, BitflowParserRULE_parameter)

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
		p.SetState(59)
		p.Name()
	}
	{
		p.SetState(60)
		p.Match(BitflowParserT__0)
	}
	{
		p.SetState(61)
		p.Val()
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
	p.EnterRule(localctx, 12, BitflowParserRULE_transformParameters)
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
		p.SetState(63)
		p.Match(BitflowParserT__1)
	}
	p.SetState(72)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if ((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserBOOL)|(1<<BitflowParserIDENTIFIER))) != 0 {
		{
			p.SetState(64)
			p.Parameter()
		}
		p.SetState(69)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		for _la == BitflowParserT__2 {
			{
				p.SetState(65)
				p.Match(BitflowParserT__2)
			}
			{
				p.SetState(66)
				p.Parameter()
			}

			p.SetState(71)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)
		}

	}
	{
		p.SetState(74)
		p.Match(BitflowParserT__3)
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

func (s *PipelineContext) Input() IInputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IInputContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IInputContext)
}

func (s *PipelineContext) MultiInputPipeline() IMultiInputPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMultiInputPipelineContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMultiInputPipelineContext)
}

func (s *PipelineContext) AllNEXT() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserNEXT)
}

func (s *PipelineContext) NEXT(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserNEXT, i)
}

func (s *PipelineContext) AllPipelineElement() []IPipelineElementContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IPipelineElementContext)(nil)).Elem())
	var tst = make([]IPipelineElementContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IPipelineElementContext)
		}
	}

	return tst
}

func (s *PipelineContext) PipelineElement(i int) IPipelineElementContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineElementContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IPipelineElementContext)
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
	p.EnterRule(localctx, 14, BitflowParserRULE_pipeline)
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
	p.SetState(81)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case BitflowParserSTRING, BitflowParserNUMBER, BitflowParserBOOL, BitflowParserIDENTIFIER:
		{
			p.SetState(76)
			p.Input()
		}

	case BitflowParserT__4:
		{
			p.SetState(77)
			p.Match(BitflowParserT__4)
		}
		{
			p.SetState(78)
			p.MultiInputPipeline()
		}
		{
			p.SetState(79)
			p.Match(BitflowParserT__5)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}
	p.SetState(87)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == BitflowParserNEXT {
		{
			p.SetState(83)
			p.Match(BitflowParserNEXT)
		}
		{
			p.SetState(84)
			p.PipelineElement()
		}

		p.SetState(89)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}

	return localctx
}

// IMultiInputPipelineContext is an interface to support dynamic dispatch.
type IMultiInputPipelineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsMultiInputPipelineContext differentiates from other interfaces.
	IsMultiInputPipelineContext()
}

type MultiInputPipelineContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiInputPipelineContext() *MultiInputPipelineContext {
	var p = new(MultiInputPipelineContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_multiInputPipeline
	return p
}

func (*MultiInputPipelineContext) IsMultiInputPipelineContext() {}

func NewMultiInputPipelineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiInputPipelineContext {
	var p = new(MultiInputPipelineContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_multiInputPipeline

	return p
}

func (s *MultiInputPipelineContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiInputPipelineContext) AllPipeline() []IPipelineContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IPipelineContext)(nil)).Elem())
	var tst = make([]IPipelineContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IPipelineContext)
		}
	}

	return tst
}

func (s *MultiInputPipelineContext) Pipeline(i int) IPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IPipelineContext)
}

func (s *MultiInputPipelineContext) AllEOP() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserEOP)
}

func (s *MultiInputPipelineContext) EOP(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserEOP, i)
}

func (s *MultiInputPipelineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiInputPipelineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiInputPipelineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterMultiInputPipeline(s)
	}
}

func (s *MultiInputPipelineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitMultiInputPipeline(s)
	}
}

func (s *MultiInputPipelineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitMultiInputPipeline(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) MultiInputPipeline() (localctx IMultiInputPipelineContext) {
	localctx = NewMultiInputPipelineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, BitflowParserRULE_multiInputPipeline)
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
		p.SetState(90)
		p.Pipeline()
	}
	p.SetState(95)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 7, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(91)
				p.Match(BitflowParserEOP)
			}
			{
				p.SetState(92)
				p.Pipeline()
			}

		}
		p.SetState(97)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 7, p.GetParserRuleContext())
	}
	p.SetState(99)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOP {
		{
			p.SetState(98)
			p.Match(BitflowParserEOP)
		}

	}

	return localctx
}

// IPipelineElementContext is an interface to support dynamic dispatch.
type IPipelineElementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsPipelineElementContext differentiates from other interfaces.
	IsPipelineElementContext()
}

type PipelineElementContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPipelineElementContext() *PipelineElementContext {
	var p = new(PipelineElementContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_pipelineElement
	return p
}

func (*PipelineElementContext) IsPipelineElementContext() {}

func NewPipelineElementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PipelineElementContext {
	var p = new(PipelineElementContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_pipelineElement

	return p
}

func (s *PipelineElementContext) GetParser() antlr.Parser { return s.parser }

func (s *PipelineElementContext) Transform() ITransformContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITransformContext)
}

func (s *PipelineElementContext) Fork() IForkContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IForkContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IForkContext)
}

func (s *PipelineElementContext) MultiplexFork() IMultiplexForkContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMultiplexForkContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMultiplexForkContext)
}

func (s *PipelineElementContext) Window() IWindowContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IWindowContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IWindowContext)
}

func (s *PipelineElementContext) Output() IOutputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOutputContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOutputContext)
}

func (s *PipelineElementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PipelineElementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PipelineElementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterPipelineElement(s)
	}
}

func (s *PipelineElementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitPipelineElement(s)
	}
}

func (s *PipelineElementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitPipelineElement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) PipelineElement() (localctx IPipelineElementContext) {
	localctx = NewPipelineElementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, BitflowParserRULE_pipelineElement)

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

	p.SetState(106)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 9, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(101)
			p.Transform()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(102)
			p.Fork()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(103)
			p.MultiplexFork()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(104)
			p.Window()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(105)
			p.Output()
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
	p.EnterRule(localctx, 20, BitflowParserRULE_transform)
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
		p.SetState(108)
		p.Name()
	}
	{
		p.SetState(109)
		p.TransformParameters()
	}
	p.SetState(111)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(110)
			p.SchedulingHints()
		}

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

func (s *ForkContext) AllNamedSubPipeline() []INamedSubPipelineContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*INamedSubPipelineContext)(nil)).Elem())
	var tst = make([]INamedSubPipelineContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(INamedSubPipelineContext)
		}
	}

	return tst
}

func (s *ForkContext) NamedSubPipeline(i int) INamedSubPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INamedSubPipelineContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(INamedSubPipelineContext)
}

func (s *ForkContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *ForkContext) AllEOP() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserEOP)
}

func (s *ForkContext) EOP(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserEOP, i)
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
	p.EnterRule(localctx, 22, BitflowParserRULE_fork)
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
		p.SetState(113)
		p.Name()
	}
	{
		p.SetState(114)
		p.TransformParameters()
	}
	p.SetState(116)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(115)
			p.SchedulingHints()
		}

	}
	{
		p.SetState(118)
		p.Match(BitflowParserT__4)
	}
	{
		p.SetState(119)
		p.NamedSubPipeline()
	}
	p.SetState(124)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 12, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(120)
				p.Match(BitflowParserEOP)
			}
			{
				p.SetState(121)
				p.NamedSubPipeline()
			}

		}
		p.SetState(126)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 12, p.GetParserRuleContext())
	}
	p.SetState(128)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOP {
		{
			p.SetState(127)
			p.Match(BitflowParserEOP)
		}

	}
	{
		p.SetState(130)
		p.Match(BitflowParserT__5)
	}

	return localctx
}

// INamedSubPipelineContext is an interface to support dynamic dispatch.
type INamedSubPipelineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsNamedSubPipelineContext differentiates from other interfaces.
	IsNamedSubPipelineContext()
}

type NamedSubPipelineContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNamedSubPipelineContext() *NamedSubPipelineContext {
	var p = new(NamedSubPipelineContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_namedSubPipeline
	return p
}

func (*NamedSubPipelineContext) IsNamedSubPipelineContext() {}

func NewNamedSubPipelineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NamedSubPipelineContext {
	var p = new(NamedSubPipelineContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_namedSubPipeline

	return p
}

func (s *NamedSubPipelineContext) GetParser() antlr.Parser { return s.parser }

func (s *NamedSubPipelineContext) NEXT() antlr.TerminalNode {
	return s.GetToken(BitflowParserNEXT, 0)
}

func (s *NamedSubPipelineContext) SubPipeline() ISubPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISubPipelineContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISubPipelineContext)
}

func (s *NamedSubPipelineContext) AllName() []INameContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*INameContext)(nil)).Elem())
	var tst = make([]INameContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(INameContext)
		}
	}

	return tst
}

func (s *NamedSubPipelineContext) Name(i int) INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *NamedSubPipelineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NamedSubPipelineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NamedSubPipelineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterNamedSubPipeline(s)
	}
}

func (s *NamedSubPipelineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitNamedSubPipeline(s)
	}
}

func (s *NamedSubPipelineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitNamedSubPipeline(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) NamedSubPipeline() (localctx INamedSubPipelineContext) {
	localctx = NewNamedSubPipelineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, BitflowParserRULE_namedSubPipeline)
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
	p.SetState(133)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = (((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserBOOL)|(1<<BitflowParserIDENTIFIER))) != 0) {
		{
			p.SetState(132)
			p.Name()
		}

		p.SetState(135)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(137)
		p.Match(BitflowParserNEXT)
	}
	{
		p.SetState(138)
		p.SubPipeline()
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

func (s *SubPipelineContext) AllPipelineElement() []IPipelineElementContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IPipelineElementContext)(nil)).Elem())
	var tst = make([]IPipelineElementContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IPipelineElementContext)
		}
	}

	return tst
}

func (s *SubPipelineContext) PipelineElement(i int) IPipelineElementContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineElementContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IPipelineElementContext)
}

func (s *SubPipelineContext) AllNEXT() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserNEXT)
}

func (s *SubPipelineContext) NEXT(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserNEXT, i)
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
	p.EnterRule(localctx, 26, BitflowParserRULE_subPipeline)
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
		p.SetState(140)
		p.PipelineElement()
	}
	p.SetState(145)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == BitflowParserNEXT {
		{
			p.SetState(141)
			p.Match(BitflowParserNEXT)
		}
		{
			p.SetState(142)
			p.PipelineElement()
		}

		p.SetState(147)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}

	return localctx
}

// IMultiplexForkContext is an interface to support dynamic dispatch.
type IMultiplexForkContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsMultiplexForkContext differentiates from other interfaces.
	IsMultiplexForkContext()
}

type MultiplexForkContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiplexForkContext() *MultiplexForkContext {
	var p = new(MultiplexForkContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_multiplexFork
	return p
}

func (*MultiplexForkContext) IsMultiplexForkContext() {}

func NewMultiplexForkContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiplexForkContext {
	var p = new(MultiplexForkContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_multiplexFork

	return p
}

func (s *MultiplexForkContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiplexForkContext) AllMultiplexSubPipeline() []IMultiplexSubPipelineContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IMultiplexSubPipelineContext)(nil)).Elem())
	var tst = make([]IMultiplexSubPipelineContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IMultiplexSubPipelineContext)
		}
	}

	return tst
}

func (s *MultiplexForkContext) MultiplexSubPipeline(i int) IMultiplexSubPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMultiplexSubPipelineContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IMultiplexSubPipelineContext)
}

func (s *MultiplexForkContext) AllEOP() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserEOP)
}

func (s *MultiplexForkContext) EOP(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserEOP, i)
}

func (s *MultiplexForkContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiplexForkContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiplexForkContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterMultiplexFork(s)
	}
}

func (s *MultiplexForkContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitMultiplexFork(s)
	}
}

func (s *MultiplexForkContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitMultiplexFork(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) MultiplexFork() (localctx IMultiplexForkContext) {
	localctx = NewMultiplexForkContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, BitflowParserRULE_multiplexFork)
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
		p.SetState(148)
		p.Match(BitflowParserT__4)
	}
	{
		p.SetState(149)
		p.MultiplexSubPipeline()
	}
	p.SetState(154)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 16, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(150)
				p.Match(BitflowParserEOP)
			}
			{
				p.SetState(151)
				p.MultiplexSubPipeline()
			}

		}
		p.SetState(156)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 16, p.GetParserRuleContext())
	}
	p.SetState(158)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOP {
		{
			p.SetState(157)
			p.Match(BitflowParserEOP)
		}

	}
	{
		p.SetState(160)
		p.Match(BitflowParserT__5)
	}

	return localctx
}

// IMultiplexSubPipelineContext is an interface to support dynamic dispatch.
type IMultiplexSubPipelineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsMultiplexSubPipelineContext differentiates from other interfaces.
	IsMultiplexSubPipelineContext()
}

type MultiplexSubPipelineContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMultiplexSubPipelineContext() *MultiplexSubPipelineContext {
	var p = new(MultiplexSubPipelineContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_multiplexSubPipeline
	return p
}

func (*MultiplexSubPipelineContext) IsMultiplexSubPipelineContext() {}

func NewMultiplexSubPipelineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MultiplexSubPipelineContext {
	var p = new(MultiplexSubPipelineContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_multiplexSubPipeline

	return p
}

func (s *MultiplexSubPipelineContext) GetParser() antlr.Parser { return s.parser }

func (s *MultiplexSubPipelineContext) SubPipeline() ISubPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISubPipelineContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISubPipelineContext)
}

func (s *MultiplexSubPipelineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiplexSubPipelineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MultiplexSubPipelineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterMultiplexSubPipeline(s)
	}
}

func (s *MultiplexSubPipelineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitMultiplexSubPipeline(s)
	}
}

func (s *MultiplexSubPipelineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitMultiplexSubPipeline(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) MultiplexSubPipeline() (localctx IMultiplexSubPipelineContext) {
	localctx = NewMultiplexSubPipelineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, BitflowParserRULE_multiplexSubPipeline)

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
		p.SetState(162)
		p.SubPipeline()
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

func (s *WindowContext) TransformParameters() ITransformParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ITransformParametersContext)
}

func (s *WindowContext) WindowPipeline() IWindowPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IWindowPipelineContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IWindowPipelineContext)
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
	p.EnterRule(localctx, 32, BitflowParserRULE_window)
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
		p.SetState(164)
		p.Match(BitflowParserT__6)
	}
	{
		p.SetState(165)
		p.TransformParameters()
	}
	p.SetState(167)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserT__7 {
		{
			p.SetState(166)
			p.SchedulingHints()
		}

	}
	{
		p.SetState(169)
		p.Match(BitflowParserT__4)
	}
	{
		p.SetState(170)
		p.WindowPipeline()
	}
	{
		p.SetState(171)
		p.Match(BitflowParserT__5)
	}

	return localctx
}

// IWindowPipelineContext is an interface to support dynamic dispatch.
type IWindowPipelineContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsWindowPipelineContext differentiates from other interfaces.
	IsWindowPipelineContext()
}

type WindowPipelineContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWindowPipelineContext() *WindowPipelineContext {
	var p = new(WindowPipelineContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_windowPipeline
	return p
}

func (*WindowPipelineContext) IsWindowPipelineContext() {}

func NewWindowPipelineContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WindowPipelineContext {
	var p = new(WindowPipelineContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_windowPipeline

	return p
}

func (s *WindowPipelineContext) GetParser() antlr.Parser { return s.parser }

func (s *WindowPipelineContext) AllTransform() []ITransformContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ITransformContext)(nil)).Elem())
	var tst = make([]ITransformContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ITransformContext)
		}
	}

	return tst
}

func (s *WindowPipelineContext) Transform(i int) ITransformContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ITransformContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ITransformContext)
}

func (s *WindowPipelineContext) AllNEXT() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserNEXT)
}

func (s *WindowPipelineContext) NEXT(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserNEXT, i)
}

func (s *WindowPipelineContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WindowPipelineContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WindowPipelineContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterWindowPipeline(s)
	}
}

func (s *WindowPipelineContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitWindowPipeline(s)
	}
}

func (s *WindowPipelineContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitWindowPipeline(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) WindowPipeline() (localctx IWindowPipelineContext) {
	localctx = NewWindowPipelineContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, BitflowParserRULE_windowPipeline)
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
		p.SetState(173)
		p.Transform()
	}
	p.SetState(178)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == BitflowParserNEXT {
		{
			p.SetState(174)
			p.Match(BitflowParserNEXT)
		}
		{
			p.SetState(175)
			p.Transform()
		}

		p.SetState(180)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
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

func (s *SchedulingHintsContext) AllSchedulingParameter() []ISchedulingParameterContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ISchedulingParameterContext)(nil)).Elem())
	var tst = make([]ISchedulingParameterContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ISchedulingParameterContext)
		}
	}

	return tst
}

func (s *SchedulingHintsContext) SchedulingParameter(i int) ISchedulingParameterContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingParameterContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ISchedulingParameterContext)
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
	p.EnterRule(localctx, 36, BitflowParserRULE_schedulingHints)
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
		p.SetState(181)
		p.Match(BitflowParserT__7)
	}
	p.SetState(190)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if ((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<BitflowParserSTRING)|(1<<BitflowParserNUMBER)|(1<<BitflowParserBOOL)|(1<<BitflowParserIDENTIFIER))) != 0 {
		{
			p.SetState(182)
			p.SchedulingParameter()
		}
		p.SetState(187)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		for _la == BitflowParserT__2 {
			{
				p.SetState(183)
				p.Match(BitflowParserT__2)
			}
			{
				p.SetState(184)
				p.SchedulingParameter()
			}

			p.SetState(189)
			p.GetErrorHandler().Sync(p)
			_la = p.GetTokenStream().LA(1)
		}

	}
	{
		p.SetState(192)
		p.Match(BitflowParserT__8)
	}

	return localctx
}

// ISchedulingParameterContext is an interface to support dynamic dispatch.
type ISchedulingParameterContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsSchedulingParameterContext differentiates from other interfaces.
	IsSchedulingParameterContext()
}

type SchedulingParameterContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySchedulingParameterContext() *SchedulingParameterContext {
	var p = new(SchedulingParameterContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_schedulingParameter
	return p
}

func (*SchedulingParameterContext) IsSchedulingParameterContext() {}

func NewSchedulingParameterContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SchedulingParameterContext {
	var p = new(SchedulingParameterContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_schedulingParameter

	return p
}

func (s *SchedulingParameterContext) GetParser() antlr.Parser { return s.parser }

func (s *SchedulingParameterContext) Parameter() IParameterContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParameterContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IParameterContext)
}

func (s *SchedulingParameterContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SchedulingParameterContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SchedulingParameterContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterSchedulingParameter(s)
	}
}

func (s *SchedulingParameterContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitSchedulingParameter(s)
	}
}

func (s *SchedulingParameterContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitSchedulingParameter(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) SchedulingParameter() (localctx ISchedulingParameterContext) {
	localctx = NewSchedulingParameterContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, BitflowParserRULE_schedulingParameter)

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
		p.SetState(194)
		p.Parameter()
	}

	return localctx
}
