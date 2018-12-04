package script

import (
	"errors"
	"fmt"
	"log"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/bitflow/fork"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	internal "github.com/bitflow-stream/go-bitflow/script/script/internal"
)

type BitflowScriptParser struct {
	Registry      reg.ProcessorRegistry
	RecoverPanics bool
}

func (s *BitflowScriptParser) ParseScript(script string) (*bitflow.SamplePipeline, golib.MultiError) {
	listener := &AntlrBitflowListener{registry: s.Registry}
	runParser(script, listener, s.RecoverPanics)
	pipe := listener.pipes.PopSingle().(*bitflow.SamplePipeline)
	listener.assertCleanState()
	return pipe, listener.MultiError
}

type parsedSubpipeline struct {
	keys []string
	pipe *bitflow.SamplePipeline
}

func (s *parsedSubpipeline) Build() (*bitflow.SamplePipeline, error) {

	// TODO refactor so that every call to Build() actually returns a new pipeline instance, including new SampleProcessor instances

	return s.pipe, nil
}

func (s *parsedSubpipeline) Keys() []string {
	return s.keys
}

type ParserError struct {
	Pos     antlr.ParserRuleContext
	Message string
}

func (e ParserError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "Unknown parser error"
	}
	return formatParserError(e.Pos.GetStart().GetLine(), e.Pos.GetStart().GetColumn(), e.Pos.GetText(), msg)
}

func formatParserError(line, col int, text, msg string) string {
	if text != "" {
		text = " '" + text + "'"
	}
	return fmt.Sprintf("Line %v:%v%v: %v", line, col, text, msg)
}

type antlrErrorListener struct {
	antlr.DefaultErrorListener
	golib.MultiError
}

func (el *antlrErrorListener) pushError(err error) {
	el.Add(err)
}

func (el *antlrErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	if msg == "" {
		msg = e.GetMessage()
	}
	el.Add(errors.New(formatParserError(line, column, "", msg)))
}

type parserListener interface {
	antlr.ParseTreeListener
	antlr.ErrorListener
	pushError(err error)
}

func runParser(script string, listener parserListener, recoverPanics bool) {
	if recoverPanics {
		defer func() {
			if r := recover(); r != nil {
				listener.pushError(fmt.Errorf("Recovered panic: %v", r))
			}
		}()
	}
	input := antlr.NewInputStream(script)
	lexer := internal.NewBitflowLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := internal.NewBitflowParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(listener)
	p.BuildParseTrees = true
	tree := p.Script()
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)
}

// AntlrBitflowListener is a complete listener for a parse tree produced by BitflowParser.
type AntlrBitflowListener struct {
	internal.BaseBitflowListener
	antlrErrorListener
	registry reg.ProcessorRegistry

	pipes          GenericStack
	multiplexPipes GenericStack
	forkPipes      GenericStack
	inWindow       bool
	params         map[string]string
}

func (s *AntlrBitflowListener) assertCleanState() {
	if len(s.pipes) > 0 || len(s.multiplexPipes) > 0 || len(s.forkPipes) > 0 || s.inWindow || s.params != nil {
		s.pushError(fmt.Errorf("Unclean listener state after parsing: %v pipes, %v multiplexPipes, %v forkPipes, inWindow %v, params %v",
			len(s.pipes), len(s.multiplexPipes), len(s.forkPipes), s.inWindow, s.params))
	}
}

func (s *AntlrBitflowListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	if msg == "" {
		msg = e.GetMessage()
	}
	s.pushError(errors.New(formatParserError(line, column, "", msg)))
}

type stringContext interface {
	STRING() antlr.TerminalNode
	GetText() string
}

func unwrapString(ctx stringContext) string {
	str := ctx.STRING()
	if str != nil {
		strText := str.GetText()
		return strText[1 : len(strText)-1]
	}
	return ctx.GetText()
}

func (s *AntlrBitflowListener) pipe() *bitflow.SamplePipeline {
	return s.pipes.PeekSingle().(*bitflow.SamplePipeline)
}

func (s *AntlrBitflowListener) ExitInput(ctx *internal.InputContext) {
	inputCtx := ctx.AllName()
	inputs := make([]string, len(inputCtx))
	for i, endpoint := range inputCtx {
		inputs[i] = unwrapString(endpoint.(*internal.NameContext))
	}
	source, err := s.registry.Endpoints.CreateInput(inputs...)
	if err != nil {
		s.pushError(err)
		return
	}
	s.pipe().Source = source
}

func (s *AntlrBitflowListener) ExitOutput(ctx *internal.OutputContext) {
	output := unwrapString(ctx.Name().(*internal.NameContext))
	sink, err := s.registry.Endpoints.CreateOutput(output)
	if err != nil {
		s.pushError(err)
		return
	}
	s.pipe().Add(sink)
}

func (s *AntlrBitflowListener) ExitTransform(ctx *internal.TransformContext) {
	nameCtx := ctx.Name().(*internal.NameContext)
	name := unwrapString(nameCtx)
	params := s.params
	s.params = nil
	isBatched := s.inWindow

	regAnalysis, ok := s.registry.GetAnalysis(name)
	if !ok {
		s.pushError(ParserError{
			Pos:     nameCtx,
			Message: fmt.Sprintf("%v: %v", name, "Unknown Processor."),
		})
		return
	} else if isBatched && !regAnalysis.SupportsBatchProcessing {
		s.pushError(ParserError{
			Pos:     nameCtx,
			Message: fmt.Sprintf("%v: %v", name, "Processor used in window, but does not support batch processing."),
		})
		return
	} else if !isBatched && !regAnalysis.SupportsStreamProcessing {
		s.pushError(ParserError{
			Pos:     nameCtx,
			Message: fmt.Sprintf("%v: %v", name, "Processor used outside window, but does not support stream processing."),
		})
		return
	}

	err := regAnalysis.Params.Verify(params)
	if err == nil {
		err = regAnalysis.Func(s.pipe(), params)
	}
	if err != nil {
		s.pushError(ParserError{
			Pos:     nameCtx,
			Message: fmt.Sprintf("%v: %v", name, err),
		})
	}
}

func (s *AntlrBitflowListener) EnterTransformParameters(ctx *internal.TransformParametersContext) {
	s.params = make(map[string]string)
}

func (s *AntlrBitflowListener) ExitParameter(ctx *internal.ParameterContext) {
	key := unwrapString(ctx.Name().(*internal.NameContext))
	value := unwrapString(ctx.Val().(*internal.ValContext))
	s.params[key] = value
}

func (s *AntlrBitflowListener) EnterPipeline(ctx *internal.PipelineContext) {
	s.pipes.Push(new(bitflow.SamplePipeline))
}

func (s *AntlrBitflowListener) EnterSubPipeline(ctx *internal.SubPipelineContext) {

	log.Println("ENTER SUBPIPE")

	s.EnterPipeline(nil)
}

func (s *AntlrBitflowListener) EnterFork(ctx *internal.ForkContext) {
	s.forkPipes.Push()
}

func (s *AntlrBitflowListener) EnterNamedSubPipeline(ctx *internal.NamedSubPipelineContext) {
	log.Println("ENTER NAMED SUB")
}

func (s *AntlrBitflowListener) ExitNamedSubPipeline(ctx *internal.NamedSubPipelineContext) {
	keyCtx := ctx.AllName()
	keys := make([]string, len(keyCtx))
	for i, key := range keyCtx {
		keys[i] = unwrapString(key.(*internal.NameContext))
	}
	pipe := s.pipes.PopSingle().(*bitflow.SamplePipeline)
	subpipe := &parsedSubpipeline{keys: keys, pipe: pipe}
	s.forkPipes.Append(subpipe)
}

func (s *AntlrBitflowListener) ExitFork(ctx *internal.ForkContext) {
	nameCtx := ctx.Name().(*internal.NameContext)
	name := unwrapString(nameCtx)
	params := s.params
	s.params = nil
	subpipelines := s.forkPipes.Pop()

	forkStep, ok := s.registry.GetFork(name)
	if !ok {
		s.pushError(ParserError{
			Pos:     nameCtx,
			Message: fmt.Sprintf("Pipeline fork '%v' is unknown", name),
		})
		return
	}
	var distributor fork.Distributor
	err := forkStep.Params.Verify(params)
	if err == nil {
		regSubpipelines := make([]reg.Subpipeline, len(subpipelines))
		for i, sub := range subpipelines {
			regSubpipelines[i] = sub.(*parsedSubpipeline)
		}
		distributor, err = forkStep.Func(regSubpipelines, params)
	}
	if err != nil {
		s.pushError(ParserError{
			Pos:     nameCtx,
			Message: fmt.Sprintf("%v: %v", name, err),
		})
	}
	s.pipe().Add(&fork.SampleFork{
		Distributor: distributor,
	})
}

func (s *AntlrBitflowListener) EnterMultiplexFork(ctx *internal.MultiplexForkContext) {
	s.multiplexPipes.Push()
}

func (s *AntlrBitflowListener) ExitMultiplexFork(ctx *internal.MultiplexForkContext) {

}

func (s *AntlrBitflowListener) ExitMultiplexSubPipeline(ctx *internal.MultiplexSubPipelineContext) {
	pipe := s.pipes.PopSingle().(*bitflow.SamplePipeline)
	s.multiplexPipes.Append(pipe)
}

func (s *AntlrBitflowListener) EnterMultiInputPipeline(ctx *internal.MultiInputPipelineContext) {
	// TODO implement
	s.pushError(ParserError{
		Pos:     ctx,
		Message: fmt.Sprintf("Parallel input not yet implemented"),
	})
}

func (s *AntlrBitflowListener) EnterWindow(ctx *internal.WindowContext) {
	s.inWindow = true
}

func (s *AntlrBitflowListener) ExitWindow(ctx *internal.WindowContext) {

	// TODO implement batch mode
	// Take window parameters into account

	s.pushError(ParserError{
		Pos:     ctx,
		Message: fmt.Sprintf("Window mode not implemented"),
	})

	/* batchPipeline */
	_ = s.pipes.PopSingle().(*bitflow.SamplePipeline)
	s.inWindow = false
}

func (s *AntlrBitflowListener) EnterWindowPipeline(ctx *internal.WindowPipelineContext) {
	s.EnterPipeline(nil)
}
