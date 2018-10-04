package script // Bitflow

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	internal "github.com/antongulenko/go-bitflow-pipeline/bitflow-script/script/internal"
	"github.com/antongulenko/golib"
)

type Index int

type HintedSubscript struct {
	Index  Index
	Script string
	Hints  map[string]string
}

type BitflowScriptScheduleParser interface {
	// ParseScript takes a Bitflow Script as input and returns the parsed SamplePipeline and occurred parsing errors as output
	ParseScript(script string) ([]HintedSubscript, golib.MultiError)
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
	currentIndex       Index
	script             string
	rawScript          string
	schedulableScripts []HintedSubscript
	errors             []error
	state              StackMap
}

func (s *AntlrBitflowScriptScheduleListener) pushError(err error) {
	s.errors = append(s.errors, err)
}

func (s *AntlrBitflowScriptScheduleListener) resetListener() {
	s.schedulableScripts = make([]HintedSubscript, 0)
	s.errors = make([]error, 0)
	s.currentIndex = 1
	s.state = StackMap{}
}

// ParseScript takes a Bitflow Script as input and returns the parsed SamplePipeline and occurred parsing errors as output
func (s *AntlrBitflowScriptScheduleListener) ParseScript(script string) (scripts []HintedSubscript, errors golib.MultiError) {
	s.resetListener()
	defer s.resetListener()
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				errors = golib.MultiError{err}
			} else {
				panic(r)
			}
		}
	}()
	rawScript := strings.Replace(script, "\n", "", -1)
	s.rawScript = rawScript
	s.script = rawScript
	input := antlr.NewInputStream(rawScript)
	lexer := internal.NewBitflowLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := internal.NewBitflowParser(stream)
	p.AddErrorListener(antlr.NewConsoleErrorListener())
	p.BuildParseTrees = true
	tree := p.Script()
	antlr.ParseTreeWalkerDefault.Walk(s, tree)
	return sortHintedScripts(s.schedulableScripts), s.errors
}

func sortHintedScripts(subscripts []HintedSubscript) []HintedSubscript{
	res := make([]HintedSubscript,len(subscripts))
	for _,s := range subscripts{
		res[s.Index] = s
	}
	return res
}

/**
	extractHintedSubscript takes a BaseParserRuleContext and replaces it's occurrence in the script with an index.
	The script is updated during parsing, but the antlr.BaseParserRuleContext's properties are static, including start and
	end position. Therefore extractHintedSubscript is keeping track of the offset between
 */
func (s *AntlrBitflowScriptScheduleListener) addHintedSubscript(from, until int, hints map[string]string) Index {
	index := s.currentIndex
	s.currentIndex++
	runes := []rune(s.rawScript)
	subscript := runes[from : until+1]
	s.schedulableScripts = append(s.schedulableScripts, HintedSubscript{
		Hints:  hints,
		Index:  index,
		Script: string(subscript),
	})
	return index
}

/**
	extractHintedSubscript takes a BaseParserRuleContext and replaces it's occurrence in the script with an index.
	The script is updated during parsing, but the antlr.BaseParserRuleContext's properties are static, including start and
	end position. Therefore extractHintedSubscript is keeping track of the offset between
 */
func (s *AntlrBitflowScriptScheduleListener) replaceSubscriptWithReference(from, until int, index Index) {
	runes := []rune(s.script)

	reference := "{{" + strconv.Itoa(int(index)) + "}}"
	currentOffset := len([]rune(s.rawScript)) - len(runes)

	start := from - currentOffset
	stop := until - currentOffset + 1

	// HAVE: head of script -> current token -> tail of script
	// WANT: head of script -> {{reference}} -> tail of script
	head := append([]rune{}, runes[:start]...)      // head of script ->
	headRef := append(head, []rune(reference)...)   // head of script -> {{reference}}
	headRefTail := append(headRef, runes[stop:]...) // head of script -> {{reference}} -> tail of script
	updatedScript := headRefTail

	s.script = string(updatedScript)
}

// EnterScheduling_hints is called when production scheduling_hints is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterSchedulingHints(ctx *internal.SchedulingHintsContext) {
	s.state.Push("scheduling_params", map[string]string{})
}

// ExitScheduling_hints is called when production scheduling_hints is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitSchedulingHints(ctx *internal.SchedulingHintsContext) {
	parentStart, parentStop := extractParentsStartStop(ctx)

	if s.state.Pop("is_propagating") == true {
		// terminate ongoing scheduling-hint-propagation
		propStart := s.state.Pop("propagation_start").(int)
		propHints := s.state.Pop("propagated_hints").(map[string]string)
		propStop := parentStart - 1 - len("->")

		ref := s.addHintedSubscript(propStart, propStop, propHints)
		s.replaceSubscriptWithReference(propStart, propStop, ref)
	}

	hints := s.state.PopOrDefault("scheduling_params", map[string]string{}).(map[string]string)
	if val, ok := hints["propagate-down"]; ok && val == "true" {
		// save state and propagate down these scheduling hints
		s.state.Push("is_propagating", true)
		s.state.Push("propagation_start", parentStart)
		s.state.Push("propagated_hints", hints)
		return
	}

	// apply hints now
	ref := s.addHintedSubscript(parentStart, parentStop, hints)
	s.replaceSubscriptWithReference(parentStart, parentStop, ref)
}

func extractParentsStartStop(ctx *internal.SchedulingHintsContext) (start int, stop int) {
	var baseCtx *antlr.BaseParserRuleContext
	switch c := ctx.GetParent().(type) {
	case *internal.TransformContext:
		baseCtx = c.BaseParserRuleContext
	case *internal.ForkContext:
		baseCtx = c.BaseParserRuleContext
	case *internal.WindowContext:
		baseCtx = c.BaseParserRuleContext
	}
	start = baseCtx.GetStart().GetStart()
	stop = baseCtx.GetStop().GetStop()
	return start, stop
}

// ExitScript is called when production script is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitScript(ctx *internal.ScriptContext) {
	if s.state.Pop("is_propagating") == true {
		// terminate ongoing scheduling-hint-propagation
		propStart := s.state.Pop("propagation_start").(int)
		propHints := s.state.Pop("propagated_hints").(map[string]string)
		propStop := len([]rune(s.rawScript)) - 1

		ref := s.addHintedSubscript(propStart, propStop, propHints)
		s.replaceSubscriptWithReference(propStart, propStop, ref)
	}

	// add the root script with Index 0
	s.schedulableScripts = append(s.schedulableScripts, HintedSubscript{
		Script: s.script,
		Index:  0,
	})
}

// ExitParameter is called when production parameter is exited.
func (s *AntlrBitflowScriptScheduleListener) ExitParameter(ctx *internal.ParameterContext) {
	p := s.state.Peek("scheduling_params")
	if p == nil {
		return
	}
	params := p.(map[string]string)
	if ctx.GetChildCount() != 3 {
		msg := fmt.Sprintf("Invalid parameter. "+
			"Parameters must be given in the form of <key>=<value>, but was %v", ctx.GetText())
		s.pushError(errors.New(msg))
	}
	key := ctx.GetChild(0).(*antlr.TerminalNodeImpl)
	val := ctx.GetChild(2).(*antlr.TerminalNodeImpl)
	params[key.GetText()] = val.GetText()
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

func (s *AntlrBitflowScriptScheduleListener) ExitEveryRule(ctx antlr.ParserRuleContext) {

}

// EnterScript is called when production script is entered.
func (s *AntlrBitflowScriptScheduleListener) EnterScript(ctx *internal.ScriptContext) {
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

func (s *AntlrBitflowScriptScheduleListener) EnterWindow(ctx *internal.WindowContext) {

}

func (s *AntlrBitflowScriptScheduleListener) ExitWindow(ctx *internal.WindowContext) {
}
