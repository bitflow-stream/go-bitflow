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
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 19, 192,
	4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7, 9, 7,
	4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12, 4, 13,
	9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 4, 18, 9,
	18, 4, 19, 9, 19, 3, 2, 3, 2, 3, 2, 3, 3, 6, 3, 43, 10, 3, 13, 3, 14, 3,
	44, 3, 3, 5, 3, 48, 10, 3, 3, 4, 3, 4, 5, 4, 52, 10, 4, 3, 5, 3, 5, 3,
	6, 3, 6, 3, 6, 3, 6, 3, 7, 3, 7, 3, 7, 7, 7, 63, 10, 7, 12, 7, 14, 7, 66,
	11, 7, 3, 8, 3, 8, 3, 8, 5, 8, 71, 10, 8, 5, 8, 73, 10, 8, 3, 8, 3, 8,
	3, 9, 3, 9, 3, 9, 7, 9, 80, 10, 9, 12, 9, 14, 9, 83, 11, 9, 3, 9, 5, 9,
	86, 10, 9, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10, 5, 10, 94, 10, 10,
	3, 10, 3, 10, 7, 10, 98, 10, 10, 12, 10, 14, 10, 101, 11, 10, 3, 11, 3,
	11, 3, 11, 5, 11, 106, 10, 11, 3, 12, 3, 12, 3, 12, 5, 12, 111, 10, 12,
	3, 13, 3, 13, 3, 13, 5, 13, 116, 10, 13, 3, 14, 3, 14, 3, 14, 5, 14, 121,
	10, 14, 3, 14, 3, 14, 3, 14, 3, 14, 7, 14, 127, 10, 14, 12, 14, 14, 14,
	130, 11, 14, 3, 14, 5, 14, 133, 10, 14, 3, 14, 3, 14, 3, 15, 6, 15, 138,
	10, 15, 13, 15, 14, 15, 139, 3, 15, 3, 15, 3, 15, 3, 16, 3, 16, 3, 16,
	7, 16, 148, 10, 16, 12, 16, 14, 16, 151, 11, 16, 3, 17, 3, 17, 3, 17, 3,
	17, 7, 17, 157, 10, 17, 12, 17, 14, 17, 160, 11, 17, 3, 17, 5, 17, 163,
	10, 17, 3, 17, 3, 17, 3, 18, 3, 18, 3, 18, 5, 18, 170, 10, 18, 3, 18, 3,
	18, 3, 18, 3, 18, 7, 18, 176, 10, 18, 12, 18, 14, 18, 179, 11, 18, 3, 18,
	3, 18, 3, 19, 3, 19, 3, 19, 5, 19, 186, 10, 19, 5, 19, 188, 10, 19, 3,
	19, 3, 19, 3, 19, 2, 2, 20, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24,
	26, 28, 30, 32, 34, 36, 2, 3, 3, 2, 14, 15, 2, 200, 2, 38, 3, 2, 2, 2,
	4, 42, 3, 2, 2, 2, 6, 49, 3, 2, 2, 2, 8, 53, 3, 2, 2, 2, 10, 55, 3, 2,
	2, 2, 12, 59, 3, 2, 2, 2, 14, 67, 3, 2, 2, 2, 16, 76, 3, 2, 2, 2, 18, 93,
	3, 2, 2, 2, 20, 105, 3, 2, 2, 2, 22, 110, 3, 2, 2, 2, 24, 112, 3, 2, 2,
	2, 26, 117, 3, 2, 2, 2, 28, 137, 3, 2, 2, 2, 30, 144, 3, 2, 2, 2, 32, 152,
	3, 2, 2, 2, 34, 166, 3, 2, 2, 2, 36, 182, 3, 2, 2, 2, 38, 39, 5, 16, 9,
	2, 39, 40, 7, 2, 2, 3, 40, 3, 3, 2, 2, 2, 41, 43, 5, 8, 5, 2, 42, 41, 3,
	2, 2, 2, 43, 44, 3, 2, 2, 2, 44, 42, 3, 2, 2, 2, 44, 45, 3, 2, 2, 2, 45,
	47, 3, 2, 2, 2, 46, 48, 5, 36, 19, 2, 47, 46, 3, 2, 2, 2, 47, 48, 3, 2,
	2, 2, 48, 5, 3, 2, 2, 2, 49, 51, 5, 8, 5, 2, 50, 52, 5, 36, 19, 2, 51,
	50, 3, 2, 2, 2, 51, 52, 3, 2, 2, 2, 52, 7, 3, 2, 2, 2, 53, 54, 9, 2, 2,
	2, 54, 9, 3, 2, 2, 2, 55, 56, 5, 8, 5, 2, 56, 57, 7, 9, 2, 2, 57, 58, 5,
	8, 5, 2, 58, 11, 3, 2, 2, 2, 59, 64, 5, 10, 6, 2, 60, 61, 7, 10, 2, 2,
	61, 63, 5, 10, 6, 2, 62, 60, 3, 2, 2, 2, 63, 66, 3, 2, 2, 2, 64, 62, 3,
	2, 2, 2, 64, 65, 3, 2, 2, 2, 65, 13, 3, 2, 2, 2, 66, 64, 3, 2, 2, 2, 67,
	72, 7, 7, 2, 2, 68, 70, 5, 12, 7, 2, 69, 71, 7, 10, 2, 2, 70, 69, 3, 2,
	2, 2, 70, 71, 3, 2, 2, 2, 71, 73, 3, 2, 2, 2, 72, 68, 3, 2, 2, 2, 72, 73,
	3, 2, 2, 2, 73, 74, 3, 2, 2, 2, 74, 75, 7, 8, 2, 2, 75, 15, 3, 2, 2, 2,
	76, 81, 5, 18, 10, 2, 77, 78, 7, 5, 2, 2, 78, 80, 5, 18, 10, 2, 79, 77,
	3, 2, 2, 2, 80, 83, 3, 2, 2, 2, 81, 79, 3, 2, 2, 2, 81, 82, 3, 2, 2, 2,
	82, 85, 3, 2, 2, 2, 83, 81, 3, 2, 2, 2, 84, 86, 7, 5, 2, 2, 85, 84, 3,
	2, 2, 2, 85, 86, 3, 2, 2, 2, 86, 17, 3, 2, 2, 2, 87, 94, 5, 4, 3, 2, 88,
	94, 5, 20, 11, 2, 89, 90, 7, 3, 2, 2, 90, 91, 5, 16, 9, 2, 91, 92, 7, 4,
	2, 2, 92, 94, 3, 2, 2, 2, 93, 87, 3, 2, 2, 2, 93, 88, 3, 2, 2, 2, 93, 89,
	3, 2, 2, 2, 94, 99, 3, 2, 2, 2, 95, 96, 7, 6, 2, 2, 96, 98, 5, 22, 12,
	2, 97, 95, 3, 2, 2, 2, 98, 101, 3, 2, 2, 2, 99, 97, 3, 2, 2, 2, 99, 100,
	3, 2, 2, 2, 100, 19, 3, 2, 2, 2, 101, 99, 3, 2, 2, 2, 102, 106, 5, 24,
	13, 2, 103, 106, 5, 26, 14, 2, 104, 106, 5, 34, 18, 2, 105, 102, 3, 2,
	2, 2, 105, 103, 3, 2, 2, 2, 105, 104, 3, 2, 2, 2, 106, 21, 3, 2, 2, 2,
	107, 111, 5, 20, 11, 2, 108, 111, 5, 32, 17, 2, 109, 111, 5, 6, 4, 2, 110,
	107, 3, 2, 2, 2, 110, 108, 3, 2, 2, 2, 110, 109, 3, 2, 2, 2, 111, 23, 3,
	2, 2, 2, 112, 113, 5, 8, 5, 2, 113, 115, 5, 14, 8, 2, 114, 116, 5, 36,
	19, 2, 115, 114, 3, 2, 2, 2, 115, 116, 3, 2, 2, 2, 116, 25, 3, 2, 2, 2,
	117, 118, 5, 8, 5, 2, 118, 120, 5, 14, 8, 2, 119, 121, 5, 36, 19, 2, 120,
	119, 3, 2, 2, 2, 120, 121, 3, 2, 2, 2, 121, 122, 3, 2, 2, 2, 122, 123,
	7, 3, 2, 2, 123, 128, 5, 28, 15, 2, 124, 125, 7, 5, 2, 2, 125, 127, 5,
	28, 15, 2, 126, 124, 3, 2, 2, 2, 127, 130, 3, 2, 2, 2, 128, 126, 3, 2,
	2, 2, 128, 129, 3, 2, 2, 2, 129, 132, 3, 2, 2, 2, 130, 128, 3, 2, 2, 2,
	131, 133, 7, 5, 2, 2, 132, 131, 3, 2, 2, 2, 132, 133, 3, 2, 2, 2, 133,
	134, 3, 2, 2, 2, 134, 135, 7, 4, 2, 2, 135, 27, 3, 2, 2, 2, 136, 138, 5,
	8, 5, 2, 137, 136, 3, 2, 2, 2, 138, 139, 3, 2, 2, 2, 139, 137, 3, 2, 2,
	2, 139, 140, 3, 2, 2, 2, 140, 141, 3, 2, 2, 2, 141, 142, 7, 6, 2, 2, 142,
	143, 5, 30, 16, 2, 143, 29, 3, 2, 2, 2, 144, 149, 5, 22, 12, 2, 145, 146,
	7, 6, 2, 2, 146, 148, 5, 22, 12, 2, 147, 145, 3, 2, 2, 2, 148, 151, 3,
	2, 2, 2, 149, 147, 3, 2, 2, 2, 149, 150, 3, 2, 2, 2, 150, 31, 3, 2, 2,
	2, 151, 149, 3, 2, 2, 2, 152, 153, 7, 3, 2, 2, 153, 158, 5, 30, 16, 2,
	154, 155, 7, 5, 2, 2, 155, 157, 5, 30, 16, 2, 156, 154, 3, 2, 2, 2, 157,
	160, 3, 2, 2, 2, 158, 156, 3, 2, 2, 2, 158, 159, 3, 2, 2, 2, 159, 162,
	3, 2, 2, 2, 160, 158, 3, 2, 2, 2, 161, 163, 7, 5, 2, 2, 162, 161, 3, 2,
	2, 2, 162, 163, 3, 2, 2, 2, 163, 164, 3, 2, 2, 2, 164, 165, 7, 4, 2, 2,
	165, 33, 3, 2, 2, 2, 166, 167, 7, 13, 2, 2, 167, 169, 5, 14, 8, 2, 168,
	170, 5, 36, 19, 2, 169, 168, 3, 2, 2, 2, 169, 170, 3, 2, 2, 2, 170, 171,
	3, 2, 2, 2, 171, 172, 7, 3, 2, 2, 172, 177, 5, 24, 13, 2, 173, 174, 7,
	6, 2, 2, 174, 176, 5, 24, 13, 2, 175, 173, 3, 2, 2, 2, 176, 179, 3, 2,
	2, 2, 177, 175, 3, 2, 2, 2, 177, 178, 3, 2, 2, 2, 178, 180, 3, 2, 2, 2,
	179, 177, 3, 2, 2, 2, 180, 181, 7, 4, 2, 2, 181, 35, 3, 2, 2, 2, 182, 187,
	7, 11, 2, 2, 183, 185, 5, 12, 7, 2, 184, 186, 7, 10, 2, 2, 185, 184, 3,
	2, 2, 2, 185, 186, 3, 2, 2, 2, 186, 188, 3, 2, 2, 2, 187, 183, 3, 2, 2,
	2, 187, 188, 3, 2, 2, 2, 188, 189, 3, 2, 2, 2, 189, 190, 7, 12, 2, 2, 190,
	37, 3, 2, 2, 2, 26, 44, 47, 51, 64, 70, 72, 81, 85, 93, 99, 105, 110, 115,
	120, 128, 132, 139, 149, 158, 162, 169, 177, 185, 187,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'{'", "'}'", "';'", "'->'", "'('", "')'", "'='", "','", "'['", "']'",
	"'batch'", "", "", "", "", "", "'\t'",
}
var symbolicNames = []string{
	"", "OPEN", "CLOSE", "EOP", "NEXT", "OPEN_PARAMS", "CLOSE_PARAMS", "EQ",
	"SEP", "OPEN_HINTS", "CLOSE_HINTS", "WINDOW", "STRING", "IDENTIFIER", "COMMENT",
	"NEWLINE", "WHITESPACE", "TAB",
}

var ruleNames = []string{
	"script", "dataInput", "dataOutput", "name", "parameter", "parameterList",
	"parameters", "pipelines", "pipeline", "pipelineElement", "pipelineTailElement",
	"processingStep", "fork", "namedSubPipeline", "subPipeline", "multiplexFork",
	"window", "schedulingHints",
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
	BitflowParserEOF          = antlr.TokenEOF
	BitflowParserOPEN         = 1
	BitflowParserCLOSE        = 2
	BitflowParserEOP          = 3
	BitflowParserNEXT         = 4
	BitflowParserOPEN_PARAMS  = 5
	BitflowParserCLOSE_PARAMS = 6
	BitflowParserEQ           = 7
	BitflowParserSEP          = 8
	BitflowParserOPEN_HINTS   = 9
	BitflowParserCLOSE_HINTS  = 10
	BitflowParserWINDOW       = 11
	BitflowParserSTRING       = 12
	BitflowParserIDENTIFIER   = 13
	BitflowParserCOMMENT      = 14
	BitflowParserNEWLINE      = 15
	BitflowParserWHITESPACE   = 16
	BitflowParserTAB          = 17
)

// BitflowParser rules.
const (
	BitflowParserRULE_script              = 0
	BitflowParserRULE_dataInput           = 1
	BitflowParserRULE_dataOutput          = 2
	BitflowParserRULE_name                = 3
	BitflowParserRULE_parameter           = 4
	BitflowParserRULE_parameterList       = 5
	BitflowParserRULE_parameters          = 6
	BitflowParserRULE_pipelines           = 7
	BitflowParserRULE_pipeline            = 8
	BitflowParserRULE_pipelineElement     = 9
	BitflowParserRULE_pipelineTailElement = 10
	BitflowParserRULE_processingStep      = 11
	BitflowParserRULE_fork                = 12
	BitflowParserRULE_namedSubPipeline    = 13
	BitflowParserRULE_subPipeline         = 14
	BitflowParserRULE_multiplexFork       = 15
	BitflowParserRULE_window              = 16
	BitflowParserRULE_schedulingHints     = 17
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

func (s *ScriptContext) Pipelines() IPipelinesContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelinesContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IPipelinesContext)
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
		p.SetState(36)
		p.Pipelines()
	}
	{
		p.SetState(37)
		p.Match(BitflowParserEOF)
	}

	return localctx
}

// IDataInputContext is an interface to support dynamic dispatch.
type IDataInputContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsDataInputContext differentiates from other interfaces.
	IsDataInputContext()
}

type DataInputContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDataInputContext() *DataInputContext {
	var p = new(DataInputContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_dataInput
	return p
}

func (*DataInputContext) IsDataInputContext() {}

func NewDataInputContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DataInputContext {
	var p = new(DataInputContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_dataInput

	return p
}

func (s *DataInputContext) GetParser() antlr.Parser { return s.parser }

func (s *DataInputContext) AllName() []INameContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*INameContext)(nil)).Elem())
	var tst = make([]INameContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(INameContext)
		}
	}

	return tst
}

func (s *DataInputContext) Name(i int) INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *DataInputContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *DataInputContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DataInputContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DataInputContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterDataInput(s)
	}
}

func (s *DataInputContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitDataInput(s)
	}
}

func (s *DataInputContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitDataInput(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) DataInput() (localctx IDataInputContext) {
	localctx = NewDataInputContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, BitflowParserRULE_dataInput)
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
	p.SetState(40)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == BitflowParserSTRING || _la == BitflowParserIDENTIFIER {
		{
			p.SetState(39)
			p.Name()
		}

		p.SetState(42)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(45)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserOPEN_HINTS {
		{
			p.SetState(44)
			p.SchedulingHints()
		}

	}

	return localctx
}

// IDataOutputContext is an interface to support dynamic dispatch.
type IDataOutputContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsDataOutputContext differentiates from other interfaces.
	IsDataOutputContext()
}

type DataOutputContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDataOutputContext() *DataOutputContext {
	var p = new(DataOutputContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_dataOutput
	return p
}

func (*DataOutputContext) IsDataOutputContext() {}

func NewDataOutputContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DataOutputContext {
	var p = new(DataOutputContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_dataOutput

	return p
}

func (s *DataOutputContext) GetParser() antlr.Parser { return s.parser }

func (s *DataOutputContext) Name() INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *DataOutputContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *DataOutputContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DataOutputContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DataOutputContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterDataOutput(s)
	}
}

func (s *DataOutputContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitDataOutput(s)
	}
}

func (s *DataOutputContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitDataOutput(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) DataOutput() (localctx IDataOutputContext) {
	localctx = NewDataOutputContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, BitflowParserRULE_dataOutput)
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
		p.SetState(47)
		p.Name()
	}
	p.SetState(49)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserOPEN_HINTS {
		{
			p.SetState(48)
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
		p.SetState(51)
		_la = p.GetTokenStream().LA(1)

		if !(_la == BitflowParserSTRING || _la == BitflowParserIDENTIFIER) {
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

func (s *ParameterContext) AllName() []INameContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*INameContext)(nil)).Elem())
	var tst = make([]INameContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(INameContext)
		}
	}

	return tst
}

func (s *ParameterContext) Name(i int) INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *ParameterContext) EQ() antlr.TerminalNode {
	return s.GetToken(BitflowParserEQ, 0)
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
	p.EnterRule(localctx, 8, BitflowParserRULE_parameter)

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
		p.SetState(53)
		p.Name()
	}
	{
		p.SetState(54)
		p.Match(BitflowParserEQ)
	}
	{
		p.SetState(55)
		p.Name()
	}

	return localctx
}

// IParameterListContext is an interface to support dynamic dispatch.
type IParameterListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsParameterListContext differentiates from other interfaces.
	IsParameterListContext()
}

type ParameterListContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyParameterListContext() *ParameterListContext {
	var p = new(ParameterListContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_parameterList
	return p
}

func (*ParameterListContext) IsParameterListContext() {}

func NewParameterListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ParameterListContext {
	var p = new(ParameterListContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_parameterList

	return p
}

func (s *ParameterListContext) GetParser() antlr.Parser { return s.parser }

func (s *ParameterListContext) AllParameter() []IParameterContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IParameterContext)(nil)).Elem())
	var tst = make([]IParameterContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IParameterContext)
		}
	}

	return tst
}

func (s *ParameterListContext) Parameter(i int) IParameterContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParameterContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IParameterContext)
}

func (s *ParameterListContext) AllSEP() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserSEP)
}

func (s *ParameterListContext) SEP(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserSEP, i)
}

func (s *ParameterListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParameterListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ParameterListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterParameterList(s)
	}
}

func (s *ParameterListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitParameterList(s)
	}
}

func (s *ParameterListContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitParameterList(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) ParameterList() (localctx IParameterListContext) {
	localctx = NewParameterListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, BitflowParserRULE_parameterList)

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
		p.SetState(57)
		p.Parameter()
	}
	p.SetState(62)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(58)
				p.Match(BitflowParserSEP)
			}
			{
				p.SetState(59)
				p.Parameter()
			}

		}
		p.SetState(64)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 3, p.GetParserRuleContext())
	}

	return localctx
}

// IParametersContext is an interface to support dynamic dispatch.
type IParametersContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsParametersContext differentiates from other interfaces.
	IsParametersContext()
}

type ParametersContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyParametersContext() *ParametersContext {
	var p = new(ParametersContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_parameters
	return p
}

func (*ParametersContext) IsParametersContext() {}

func NewParametersContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ParametersContext {
	var p = new(ParametersContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_parameters

	return p
}

func (s *ParametersContext) GetParser() antlr.Parser { return s.parser }

func (s *ParametersContext) OPEN_PARAMS() antlr.TerminalNode {
	return s.GetToken(BitflowParserOPEN_PARAMS, 0)
}

func (s *ParametersContext) CLOSE_PARAMS() antlr.TerminalNode {
	return s.GetToken(BitflowParserCLOSE_PARAMS, 0)
}

func (s *ParametersContext) ParameterList() IParameterListContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParameterListContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IParameterListContext)
}

func (s *ParametersContext) SEP() antlr.TerminalNode {
	return s.GetToken(BitflowParserSEP, 0)
}

func (s *ParametersContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParametersContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ParametersContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterParameters(s)
	}
}

func (s *ParametersContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitParameters(s)
	}
}

func (s *ParametersContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitParameters(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Parameters() (localctx IParametersContext) {
	localctx = NewParametersContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, BitflowParserRULE_parameters)
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
		p.SetState(65)
		p.Match(BitflowParserOPEN_PARAMS)
	}
	p.SetState(70)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserSTRING || _la == BitflowParserIDENTIFIER {
		{
			p.SetState(66)
			p.ParameterList()
		}
		p.SetState(68)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		if _la == BitflowParserSEP {
			{
				p.SetState(67)
				p.Match(BitflowParserSEP)
			}

		}

	}
	{
		p.SetState(72)
		p.Match(BitflowParserCLOSE_PARAMS)
	}

	return localctx
}

// IPipelinesContext is an interface to support dynamic dispatch.
type IPipelinesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsPipelinesContext differentiates from other interfaces.
	IsPipelinesContext()
}

type PipelinesContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPipelinesContext() *PipelinesContext {
	var p = new(PipelinesContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_pipelines
	return p
}

func (*PipelinesContext) IsPipelinesContext() {}

func NewPipelinesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PipelinesContext {
	var p = new(PipelinesContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_pipelines

	return p
}

func (s *PipelinesContext) GetParser() antlr.Parser { return s.parser }

func (s *PipelinesContext) AllPipeline() []IPipelineContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IPipelineContext)(nil)).Elem())
	var tst = make([]IPipelineContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IPipelineContext)
		}
	}

	return tst
}

func (s *PipelinesContext) Pipeline(i int) IPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IPipelineContext)
}

func (s *PipelinesContext) AllEOP() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserEOP)
}

func (s *PipelinesContext) EOP(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserEOP, i)
}

func (s *PipelinesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PipelinesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PipelinesContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterPipelines(s)
	}
}

func (s *PipelinesContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitPipelines(s)
	}
}

func (s *PipelinesContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitPipelines(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) Pipelines() (localctx IPipelinesContext) {
	localctx = NewPipelinesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, BitflowParserRULE_pipelines)
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
		p.SetState(74)
		p.Pipeline()
	}
	p.SetState(79)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 6, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(75)
				p.Match(BitflowParserEOP)
			}
			{
				p.SetState(76)
				p.Pipeline()
			}

		}
		p.SetState(81)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 6, p.GetParserRuleContext())
	}
	p.SetState(83)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOP {
		{
			p.SetState(82)
			p.Match(BitflowParserEOP)
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

func (s *PipelineContext) DataInput() IDataInputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDataInputContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDataInputContext)
}

func (s *PipelineContext) PipelineElement() IPipelineElementContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineElementContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IPipelineElementContext)
}

func (s *PipelineContext) OPEN() antlr.TerminalNode {
	return s.GetToken(BitflowParserOPEN, 0)
}

func (s *PipelineContext) Pipelines() IPipelinesContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelinesContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IPipelinesContext)
}

func (s *PipelineContext) CLOSE() antlr.TerminalNode {
	return s.GetToken(BitflowParserCLOSE, 0)
}

func (s *PipelineContext) AllNEXT() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserNEXT)
}

func (s *PipelineContext) NEXT(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserNEXT, i)
}

func (s *PipelineContext) AllPipelineTailElement() []IPipelineTailElementContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IPipelineTailElementContext)(nil)).Elem())
	var tst = make([]IPipelineTailElementContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IPipelineTailElementContext)
		}
	}

	return tst
}

func (s *PipelineContext) PipelineTailElement(i int) IPipelineTailElementContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineTailElementContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IPipelineTailElementContext)
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
	p.EnterRule(localctx, 16, BitflowParserRULE_pipeline)
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
	p.SetState(91)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 8, p.GetParserRuleContext()) {
	case 1:
		{
			p.SetState(85)
			p.DataInput()
		}

	case 2:
		{
			p.SetState(86)
			p.PipelineElement()
		}

	case 3:
		{
			p.SetState(87)
			p.Match(BitflowParserOPEN)
		}
		{
			p.SetState(88)
			p.Pipelines()
		}
		{
			p.SetState(89)
			p.Match(BitflowParserCLOSE)
		}

	}
	p.SetState(97)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == BitflowParserNEXT {
		{
			p.SetState(93)
			p.Match(BitflowParserNEXT)
		}
		{
			p.SetState(94)
			p.PipelineTailElement()
		}

		p.SetState(99)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
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

func (s *PipelineElementContext) ProcessingStep() IProcessingStepContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IProcessingStepContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IProcessingStepContext)
}

func (s *PipelineElementContext) Fork() IForkContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IForkContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IForkContext)
}

func (s *PipelineElementContext) Window() IWindowContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IWindowContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IWindowContext)
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

	p.SetState(103)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 10, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(100)
			p.ProcessingStep()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(101)
			p.Fork()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(102)
			p.Window()
		}

	}

	return localctx
}

// IPipelineTailElementContext is an interface to support dynamic dispatch.
type IPipelineTailElementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsPipelineTailElementContext differentiates from other interfaces.
	IsPipelineTailElementContext()
}

type PipelineTailElementContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPipelineTailElementContext() *PipelineTailElementContext {
	var p = new(PipelineTailElementContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_pipelineTailElement
	return p
}

func (*PipelineTailElementContext) IsPipelineTailElementContext() {}

func NewPipelineTailElementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PipelineTailElementContext {
	var p = new(PipelineTailElementContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_pipelineTailElement

	return p
}

func (s *PipelineTailElementContext) GetParser() antlr.Parser { return s.parser }

func (s *PipelineTailElementContext) PipelineElement() IPipelineElementContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineElementContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IPipelineElementContext)
}

func (s *PipelineTailElementContext) MultiplexFork() IMultiplexForkContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IMultiplexForkContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IMultiplexForkContext)
}

func (s *PipelineTailElementContext) DataOutput() IDataOutputContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IDataOutputContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IDataOutputContext)
}

func (s *PipelineTailElementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PipelineTailElementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PipelineTailElementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterPipelineTailElement(s)
	}
}

func (s *PipelineTailElementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitPipelineTailElement(s)
	}
}

func (s *PipelineTailElementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitPipelineTailElement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) PipelineTailElement() (localctx IPipelineTailElementContext) {
	localctx = NewPipelineTailElementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, BitflowParserRULE_pipelineTailElement)

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

	p.SetState(108)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 11, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(105)
			p.PipelineElement()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(106)
			p.MultiplexFork()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(107)
			p.DataOutput()
		}

	}

	return localctx
}

// IProcessingStepContext is an interface to support dynamic dispatch.
type IProcessingStepContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsProcessingStepContext differentiates from other interfaces.
	IsProcessingStepContext()
}

type ProcessingStepContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyProcessingStepContext() *ProcessingStepContext {
	var p = new(ProcessingStepContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = BitflowParserRULE_processingStep
	return p
}

func (*ProcessingStepContext) IsProcessingStepContext() {}

func NewProcessingStepContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ProcessingStepContext {
	var p = new(ProcessingStepContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = BitflowParserRULE_processingStep

	return p
}

func (s *ProcessingStepContext) GetParser() antlr.Parser { return s.parser }

func (s *ProcessingStepContext) Name() INameContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*INameContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(INameContext)
}

func (s *ProcessingStepContext) Parameters() IParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IParametersContext)
}

func (s *ProcessingStepContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *ProcessingStepContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ProcessingStepContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ProcessingStepContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.EnterProcessingStep(s)
	}
}

func (s *ProcessingStepContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(BitflowListener); ok {
		listenerT.ExitProcessingStep(s)
	}
}

func (s *ProcessingStepContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case BitflowVisitor:
		return t.VisitProcessingStep(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *BitflowParser) ProcessingStep() (localctx IProcessingStepContext) {
	localctx = NewProcessingStepContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, BitflowParserRULE_processingStep)
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
		p.SetState(110)
		p.Name()
	}
	{
		p.SetState(111)
		p.Parameters()
	}
	p.SetState(113)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserOPEN_HINTS {
		{
			p.SetState(112)
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

func (s *ForkContext) Parameters() IParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IParametersContext)
}

func (s *ForkContext) OPEN() antlr.TerminalNode {
	return s.GetToken(BitflowParserOPEN, 0)
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

func (s *ForkContext) CLOSE() antlr.TerminalNode {
	return s.GetToken(BitflowParserCLOSE, 0)
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
	p.EnterRule(localctx, 24, BitflowParserRULE_fork)
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
		p.SetState(115)
		p.Name()
	}
	{
		p.SetState(116)
		p.Parameters()
	}
	p.SetState(118)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserOPEN_HINTS {
		{
			p.SetState(117)
			p.SchedulingHints()
		}

	}
	{
		p.SetState(120)
		p.Match(BitflowParserOPEN)
	}
	{
		p.SetState(121)
		p.NamedSubPipeline()
	}
	p.SetState(126)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 14, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(122)
				p.Match(BitflowParserEOP)
			}
			{
				p.SetState(123)
				p.NamedSubPipeline()
			}

		}
		p.SetState(128)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 14, p.GetParserRuleContext())
	}
	p.SetState(130)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOP {
		{
			p.SetState(129)
			p.Match(BitflowParserEOP)
		}

	}
	{
		p.SetState(132)
		p.Match(BitflowParserCLOSE)
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
	p.EnterRule(localctx, 26, BitflowParserRULE_namedSubPipeline)
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
	p.SetState(135)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == BitflowParserSTRING || _la == BitflowParserIDENTIFIER {
		{
			p.SetState(134)
			p.Name()
		}

		p.SetState(137)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(139)
		p.Match(BitflowParserNEXT)
	}
	{
		p.SetState(140)
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

func (s *SubPipelineContext) AllPipelineTailElement() []IPipelineTailElementContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IPipelineTailElementContext)(nil)).Elem())
	var tst = make([]IPipelineTailElementContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IPipelineTailElementContext)
		}
	}

	return tst
}

func (s *SubPipelineContext) PipelineTailElement(i int) IPipelineTailElementContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IPipelineTailElementContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IPipelineTailElementContext)
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
	p.EnterRule(localctx, 28, BitflowParserRULE_subPipeline)
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
		p.SetState(142)
		p.PipelineTailElement()
	}
	p.SetState(147)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == BitflowParserNEXT {
		{
			p.SetState(143)
			p.Match(BitflowParserNEXT)
		}
		{
			p.SetState(144)
			p.PipelineTailElement()
		}

		p.SetState(149)
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

func (s *MultiplexForkContext) OPEN() antlr.TerminalNode {
	return s.GetToken(BitflowParserOPEN, 0)
}

func (s *MultiplexForkContext) AllSubPipeline() []ISubPipelineContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*ISubPipelineContext)(nil)).Elem())
	var tst = make([]ISubPipelineContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(ISubPipelineContext)
		}
	}

	return tst
}

func (s *MultiplexForkContext) SubPipeline(i int) ISubPipelineContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISubPipelineContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(ISubPipelineContext)
}

func (s *MultiplexForkContext) CLOSE() antlr.TerminalNode {
	return s.GetToken(BitflowParserCLOSE, 0)
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
	p.EnterRule(localctx, 30, BitflowParserRULE_multiplexFork)
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
		p.SetState(150)
		p.Match(BitflowParserOPEN)
	}
	{
		p.SetState(151)
		p.SubPipeline()
	}
	p.SetState(156)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 18, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(152)
				p.Match(BitflowParserEOP)
			}
			{
				p.SetState(153)
				p.SubPipeline()
			}

		}
		p.SetState(158)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 18, p.GetParserRuleContext())
	}
	p.SetState(160)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserEOP {
		{
			p.SetState(159)
			p.Match(BitflowParserEOP)
		}

	}
	{
		p.SetState(162)
		p.Match(BitflowParserCLOSE)
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

func (s *WindowContext) WINDOW() antlr.TerminalNode {
	return s.GetToken(BitflowParserWINDOW, 0)
}

func (s *WindowContext) Parameters() IParametersContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParametersContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IParametersContext)
}

func (s *WindowContext) OPEN() antlr.TerminalNode {
	return s.GetToken(BitflowParserOPEN, 0)
}

func (s *WindowContext) AllProcessingStep() []IProcessingStepContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IProcessingStepContext)(nil)).Elem())
	var tst = make([]IProcessingStepContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IProcessingStepContext)
		}
	}

	return tst
}

func (s *WindowContext) ProcessingStep(i int) IProcessingStepContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IProcessingStepContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IProcessingStepContext)
}

func (s *WindowContext) CLOSE() antlr.TerminalNode {
	return s.GetToken(BitflowParserCLOSE, 0)
}

func (s *WindowContext) SchedulingHints() ISchedulingHintsContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ISchedulingHintsContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ISchedulingHintsContext)
}

func (s *WindowContext) AllNEXT() []antlr.TerminalNode {
	return s.GetTokens(BitflowParserNEXT)
}

func (s *WindowContext) NEXT(i int) antlr.TerminalNode {
	return s.GetToken(BitflowParserNEXT, i)
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
		p.Match(BitflowParserWINDOW)
	}
	{
		p.SetState(165)
		p.Parameters()
	}
	p.SetState(167)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserOPEN_HINTS {
		{
			p.SetState(166)
			p.SchedulingHints()
		}

	}
	{
		p.SetState(169)
		p.Match(BitflowParserOPEN)
	}
	{
		p.SetState(170)
		p.ProcessingStep()
	}
	p.SetState(175)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == BitflowParserNEXT {
		{
			p.SetState(171)
			p.Match(BitflowParserNEXT)
		}
		{
			p.SetState(172)
			p.ProcessingStep()
		}

		p.SetState(177)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(178)
		p.Match(BitflowParserCLOSE)
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

func (s *SchedulingHintsContext) OPEN_HINTS() antlr.TerminalNode {
	return s.GetToken(BitflowParserOPEN_HINTS, 0)
}

func (s *SchedulingHintsContext) CLOSE_HINTS() antlr.TerminalNode {
	return s.GetToken(BitflowParserCLOSE_HINTS, 0)
}

func (s *SchedulingHintsContext) ParameterList() IParameterListContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IParameterListContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IParameterListContext)
}

func (s *SchedulingHintsContext) SEP() antlr.TerminalNode {
	return s.GetToken(BitflowParserSEP, 0)
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
	p.EnterRule(localctx, 34, BitflowParserRULE_schedulingHints)
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
		p.SetState(180)
		p.Match(BitflowParserOPEN_HINTS)
	}
	p.SetState(185)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	if _la == BitflowParserSTRING || _la == BitflowParserIDENTIFIER {
		{
			p.SetState(181)
			p.ParameterList()
		}
		p.SetState(183)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)

		if _la == BitflowParserSEP {
			{
				p.SetState(182)
				p.Match(BitflowParserSEP)
			}

		}

	}
	{
		p.SetState(187)
		p.Match(BitflowParserCLOSE_HINTS)
	}

	return localctx
}
