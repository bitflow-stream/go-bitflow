package script

import (
	"errors"
	"fmt"

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
	parser := &_bitflowScriptParser{registry: &s.Registry}
	res := parser.parseScript(script, s.RecoverPanics)
	return res, parser.MultiError
}

type _bitflowScriptParser struct {
	antlr.DefaultErrorListener
	golib.MultiError
	registry *reg.ProcessorRegistry
}

func (s *_bitflowScriptParser) parseScript(script string, recoverPanics bool) *bitflow.SamplePipeline {
	if recoverPanics {
		defer func() {
			if r := recover(); r != nil {
				s.pushAnyError(fmt.Errorf("Recovered panic: %v", r))
			}
		}()
	}
	input := antlr.NewInputStream(script)
	lexer := internal.NewBitflowLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := internal.NewBitflowParser(stream)
	p.RemoveErrorListeners()
	p.AddErrorListener(s)
	p.BuildParseTrees = true
	tree := p.Script()
	return s.buildScript(tree.(*internal.ScriptContext))
}

type parsedSubpipeline struct {
	keys     []string
	pipe     *internal.SubPipelineContext
	registry *reg.ProcessorRegistry
}

func (s *parsedSubpipeline) Build() (*bitflow.SamplePipeline, error) {
	pipe := new(bitflow.SamplePipeline)
	parser := &_bitflowScriptParser{
		registry: s.registry,
	}
	parser.buildPipelineTail(pipe, s.pipe.AllPipelineElement())
	return pipe, parser.NilOrError()
}

func (s *parsedSubpipeline) Keys() []string {
	return s.keys
}

// ==============
// Error handling
// ==============

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

func (s *_bitflowScriptParser) pushAnyError(err error) {
	if err != nil {
		s.Add(err)
	}
}

func (s *_bitflowScriptParser) pushError(pos antlr.ParserRuleContext, msgFormat string, params ...interface{}) {
	s.Add(&ParserError{
		Pos:     pos,
		Message: fmt.Sprintf(msgFormat, params...),
	})
}

func (s *_bitflowScriptParser) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	if msg == "" {
		msg = e.GetMessage()
	}
	s.Add(errors.New(formatParserError(line, column, "", msg)))
}

// ==============
// Parsing
// ==============

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

func (s *_bitflowScriptParser) buildScript(ctx *internal.ScriptContext) *bitflow.SamplePipeline {
	pipe := new(bitflow.SamplePipeline)
	s.buildMultiInputPipeline(pipe, ctx.MultiInputPipeline().(*internal.MultiInputPipelineContext))
	return pipe
}

func (s *_bitflowScriptParser) buildMultiInputPipeline(pipe *bitflow.SamplePipeline, ctx *internal.MultiInputPipelineContext) {
	pipes := ctx.AllPipeline()
	if len(pipes) == 1 {
		// Optimization: avoid unnecessary parallelization of one single pipeline
		s.buildInputPipeline(pipe, pipes[0].(*internal.PipelineContext))
	} else {
		input := new(fork.MultiMetricSource)
		for _, subPipeCtx := range pipes {
			parallelPipe := new(bitflow.SamplePipeline)
			s.buildInputPipeline(parallelPipe, subPipeCtx.(*internal.PipelineContext))
			input.Add(parallelPipe)
		}
		pipe.Source = input
	}
}

func (s *_bitflowScriptParser) buildInputPipeline(pipe *bitflow.SamplePipeline, ctx *internal.PipelineContext) {
	switch {
	case ctx.MultiInputPipeline() != nil:
		s.buildMultiInputPipeline(pipe, ctx.MultiInputPipeline().(*internal.MultiInputPipelineContext))
	case ctx.DataInput() != nil:
		s.buildInput(pipe, ctx.DataInput().(*internal.DataInputContext))
	default:
		// This case is actually not possible due to the grammar. Just cover the case for the sake of completeness.
		pipe.Source = new(bitflow.EmptySampleSource)
	}
	s.buildPipelineTail(pipe, ctx.AllPipelineElement())
}

func (s *_bitflowScriptParser) buildInput(pipe *bitflow.SamplePipeline, ctx *internal.DataInputContext) {
	inputCtx := ctx.AllName()
	inputs := make([]string, len(inputCtx))
	for i, endpoint := range inputCtx {
		inputs[i] = unwrapString(endpoint.(*internal.NameContext))
	}
	source, err := s.registry.Endpoints.CreateInput(inputs...)
	s.pushAnyError(err)
	if err == nil && source != nil {
		pipe.Source = source
	}
}

func (s *_bitflowScriptParser) buildOutput(pipe *bitflow.SamplePipeline, ctx *internal.DataOutputContext) {
	output := unwrapString(ctx.Name().(*internal.NameContext))
	sink, err := s.registry.Endpoints.CreateOutput(output)
	s.pushAnyError(err)
	if err == nil && sink != nil {
		pipe.Add(sink)
	}
}

func (s *_bitflowScriptParser) buildPipelineElement(pipe *bitflow.SamplePipeline, ctx *internal.PipelineElementContext) {
	switch {
	case ctx.Transform() != nil:
		s.buildTransform(pipe, ctx.Transform().(*internal.TransformContext))
	case ctx.Fork() != nil:
		s.buildFork(pipe, ctx.Fork().(*internal.ForkContext))
	case ctx.MultiplexFork() != nil:
		s.buildMultiplexFork(pipe, ctx.MultiplexFork().(*internal.MultiplexForkContext))
	case ctx.Window() != nil:
		s.buildWindow(pipe, ctx.Window().(*internal.WindowContext))
	case ctx.DataOutput() != nil:
		s.buildOutput(pipe, ctx.DataOutput().(*internal.DataOutputContext))
	}
}

func (s *_bitflowScriptParser) buildTransform(pipe *bitflow.SamplePipeline, ctx *internal.TransformContext) {
	nameCtx := ctx.Name().(*internal.NameContext)
	name := unwrapString(nameCtx)
	params := s.buildTransformParameters(ctx.TransformParameters().(*internal.TransformParametersContext))
	isBatched := false // TODO implement window mode

	regAnalysis, ok := s.registry.GetAnalysis(name)
	if !ok {
		s.pushError(nameCtx, "%v: %v", name, "Unknown Processor.")
		return
	} else if isBatched && !regAnalysis.SupportsBatchProcessing {
		s.pushError(nameCtx, "%v: %v", name, "Processor used in window, but does not support batch processing.")
		return
	} else if !isBatched && !regAnalysis.SupportsStreamProcessing {
		s.pushError(nameCtx, "%v: %v", name, "Processor used outside window, but does not support stream processing.")
		return
	}

	err := regAnalysis.Params.Verify(params)
	if err == nil {
		err = regAnalysis.Func(pipe, params)
	}
	if err != nil {
		s.pushError(nameCtx, "%v: %v", name, err)
	}
}

func (s *_bitflowScriptParser) buildTransformParameters(ctx *internal.TransformParametersContext) map[string]string {
	params := make(map[string]string)
	for _, paramCtxI := range ctx.AllParameter() {
		paramCtx := paramCtxI.(*internal.ParameterContext)
		key := unwrapString(paramCtx.Name().(*internal.NameContext))
		value := unwrapString(paramCtx.Val().(*internal.ValContext))
		params[key] = value
	}
	return params
}

func (s *_bitflowScriptParser) buildFork(pipe *bitflow.SamplePipeline, ctx *internal.ForkContext) {
	nameCtx := ctx.Name().(*internal.NameContext)
	name := unwrapString(nameCtx)
	params := s.buildTransformParameters(ctx.TransformParameters().(*internal.TransformParametersContext))

	// Lookup fork step and verify parameters
	forkStep, ok := s.registry.GetFork(name)
	if !ok {
		s.pushError(nameCtx, "Pipeline fork '%v' is unknown", name)
		return
	}
	err := forkStep.Params.Verify(params)
	if err != nil {
		s.pushError(nameCtx, "%v: %v", name, err)
		return
	}

	// Parse sub-pipelines
	subpipelines := make([]reg.Subpipeline, len(ctx.AllNamedSubPipeline()))
	for i, namedSubPipe := range ctx.AllNamedSubPipeline() {
		subpipelines[i] = s.buildNamedSubPipeline(namedSubPipe.(*internal.NamedSubPipelineContext))
	}

	distributor, err := forkStep.Func(subpipelines, params)
	if err != nil {
		s.pushError(nameCtx, "%v: %v", name, err)
		return
	}
	pipe.Add(&fork.SampleFork{
		Distributor: distributor,
	})
}

func (s *_bitflowScriptParser) buildNamedSubPipeline(ctx *internal.NamedSubPipelineContext) reg.Subpipeline {
	keyCtx := ctx.AllName()
	keys := make([]string, len(keyCtx))
	for i, key := range keyCtx {
		keys[i] = unwrapString(key.(*internal.NameContext))
	}
	return &parsedSubpipeline{
		keys:     keys,
		pipe:     ctx.SubPipeline().(*internal.SubPipelineContext),
		registry: s.registry,
	}
}

func (s *_bitflowScriptParser) buildPipelineTail(pipe *bitflow.SamplePipeline, elements []internal.IPipelineElementContext) {
	for _, elem := range elements {
		s.buildPipelineElement(pipe, elem.(*internal.PipelineElementContext))
	}
}

func (s *_bitflowScriptParser) buildMultiplexFork(pipe *bitflow.SamplePipeline, ctx *internal.MultiplexForkContext) {
	var multi fork.MultiplexDistributor
	for _, subpipeCtx := range ctx.AllMultiplexSubPipeline() {
		subpipe := new(bitflow.SamplePipeline)
		s.buildPipelineTail(subpipe, subpipeCtx.(*internal.MultiplexSubPipelineContext).SubPipeline().(*internal.SubPipelineContext).AllPipelineElement())
		if len(subpipe.Processors) > 0 {
			multi.Subpipelines = append(multi.Subpipelines, subpipe)
		}
	}
	if len(multi.Subpipelines) > 0 {
		pipe.Add(&fork.SampleFork{
			Distributor: &multi,
		})
	}
}

func (s *_bitflowScriptParser) buildWindow(pipe *bitflow.SamplePipeline, ctx *internal.WindowContext) {

	// TODO implement batch mode
	// Take window parameters into account

	s.pushError(ctx, "Window mode not implemented")
}
