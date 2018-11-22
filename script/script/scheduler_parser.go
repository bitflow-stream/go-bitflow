package script

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/antongulenko/golib"
	internal "github.com/bitflow-stream/go-bitflow/script/script/internal"
)

type Index int

type ScriptSchedule struct {
	Index        Index
	Script       string
	Hints        map[string]interface{}
	Predecessors []Index
	Successors   []Index
}

type BitflowScriptScheduleParser interface {
	// ParseScript takes a Bitflow Script as input and returns the parsed SamplePipeline and occurred parsing errors as output
	ParseScript(script string) ([]ScriptSchedule, golib.MultiError)
}

func NewBitflowScriptScheduleParser() BitflowScriptScheduleParser {
	r := new(AntlrBitflowScriptScheduleListener)
	r.resetListener()
	return r
}

// next line to ensure AntlrBitflowScriptScheduleListener satisfies BitflowListener interface
var _ internal.BitflowListener = &AntlrBitflowScriptScheduleListener{}

// AntlrBitflowScriptScheduleListener is a complete listener for a parse tree produced by BitflowParser.
type AntlrBitflowScriptScheduleListener struct {
	currentIndex       int
	script             string
	schedulableScripts []ScriptSchedule
	errors             []error
}

func (s *AntlrBitflowScriptScheduleListener) pushError(err error) {
	s.errors = append(s.errors, err)
}

func (s *AntlrBitflowScriptScheduleListener) resetListener() {
	s.schedulableScripts = make([]ScriptSchedule, 0)
	s.errors = make([]error, 0)
	s.currentIndex = 0
}

// ParseScript takes a Bitflow Script as input and returns the parsed SamplePipeline and occurred parsing errors as output
func (s *AntlrBitflowScriptScheduleListener) ParseScript(script string) (scripts []ScriptSchedule, errors golib.MultiError) {
	defer s.resetListener()
	defer func() {
		if r := recover(); r != nil {
			// TODO what if r is not error
			errors = golib.MultiError{r.(error)}
		}
	}()
	rawScript := strings.Replace(script, "\n", "", -1)
	input := antlr.NewInputStream(rawScript)
	lexer := internal.NewBitflowLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := internal.NewBitflowParser(stream)
	// Todo replace Error listener
	p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	p.BuildParseTrees = true
	tree := p.Script()
	antlr.ParseTreeWalkerDefault.Walk(s, tree)
	return s.schedulableScripts, s.errors
}

// EnterScheduling_hints is called when production scheduling_hints is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterSchedulingHints(ctx *internal.SchedulingHintsContext) {
	switch c := ctx.GetParent().(type) {
	case *internal.TransformContext:
		// TODO
		//replaceWithReference(script)
	case *internal.ForkContext:
		// TODO
	default:
		fmt.Println(reflect.TypeOf(c))
	}
}

func replaceWithReference(script, reference string, ctx *antlr.BaseParserRuleContext) string {
	runes := []rune(script)
	runes = append(append(runes[:ctx.GetStart().GetStart()], []rune(reference)...), runes[:ctx.GetStop().GetStop()]...)
	return string(runes)
}

// ExitScheduling_hints is called when production scheduling_hints is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitSchedulingHints(ctx *internal.SchedulingHintsContext) {
	// ignored
}

// EnterTransform_parameters is called when production transform_parameters is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterTransformParameters(ctx *internal.TransformParametersContext) {
}

// ExitTransform_parameters is called when production transform_parameters is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitTransformParameters(ctx *internal.TransformParametersContext) {
}

// EnterName is called when production name is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterName(ctx *internal.NameContext) {}

// ExitName is called when production name is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitName(ctx *internal.NameContext) {
}

// EnterOutput is called when production output is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterOutput(ctx *internal.OutputContext) {}

// ExitOutput is called when production output is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitOutput(ctx *internal.OutputContext) {
}

// EnterOutput_fork is called when production output_fork is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterOutputFork(ctx *internal.OutputForkContext) {}

// ExitOutput_fork is called when production output_fork is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitOutputFork(ctx *internal.OutputForkContext) {}

// VisitTerminal is called when a terminal node is visited.
func (s *AntlrBitflowScriptScheduleListener) VisitTerminal(node antlr.TerminalNode) {
}

// VisitErrorNode is called when an error node is visited.
func (s *AntlrBitflowScriptScheduleListener) VisitErrorNode(node antlr.ErrorNode) {
}

// EnterEveryRule is called when any rule is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterEveryRule(ctx antlr.ParserRuleContext) {

}

// ExitEveryRule is called when any rule is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitEveryRule(ctx antlr.ParserRuleContext) {
}

// EnterScript is called when production script is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterScript(ctx *internal.ScriptContext) {
}

// ExitScript is called when production script is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitScript(ctx *internal.ScriptContext) {
}

// EnterFork is called when production fork is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterFork(ctx *internal.ForkContext) {
}

// ExitFork is called when production fork is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitFork(ctx *internal.ForkContext) {
}

// EnterInput is called when production input is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterInput(ctx *internal.InputContext) {}

// ExitInput is called when production input is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitInput(ctx *internal.InputContext) {
}

// EnterMultiinput is called when production multiinput is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterMultiinput(ctx *internal.MultiinputContext) {
}

// ExitMultiinput is called when production multiinput is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitMultiinput(ctx *internal.MultiinputContext) {
}

// EnterPipeline is called when production pipeline is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterPipeline(ctx *internal.PipelineContext) {
}

// ExitPipeline is called when production pipeline is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitPipeline(ctx *internal.PipelineContext) {

}

// EnterSub_pipeline is called when production sub_pipeline is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterSubPipeline(ctx *internal.SubPipelineContext) {
}

// ExitSub_pipeline is called when production sub_pipeline is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitSubPipeline(ctx *internal.SubPipelineContext) {
}

// EnterSub_pipeline is called when production sub_pipeline is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterWindowSubPipeline(ctx *internal.WindowSubPipelineContext) {
}

// ExitSub_pipeline is called when production sub_pipeline is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitWindowSubPipeline(ctx *internal.WindowSubPipelineContext) {
}

// EnterTransform is called when production transform is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterTransform(ctx *internal.TransformContext) {
}

// ExitTransform is called when production transform is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitTransform(ctx *internal.TransformContext) {
}

// EnterParameter is called when production parameter is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterParameter(ctx *internal.ParameterContext) {
}

// ExitParameter is called when production parameter is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitPipelineName(ctx *internal.PipelineNameContext) {

}
func (s *AntlrBitflowScriptScheduleListener) EnterPipelineName(ctx *internal.PipelineNameContext) {

}
func (s *AntlrBitflowScriptScheduleListener) ExitParameter(ctx *internal.ParameterContext) {
}

func (s *AntlrBitflowScriptScheduleListener) EnterWindow(ctx *internal.WindowContext) {

}
func (s *AntlrBitflowScriptScheduleListener) ExitWindow(ctx *internal.WindowContext) {
}
