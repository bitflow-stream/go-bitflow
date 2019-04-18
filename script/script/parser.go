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
	"github.com/bitflow-stream/go-bitflow/steps"
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
	parser.buildPipelineTail(pipe, s.pipe.AllPipelineTailElement())
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
	s.buildInputPipelines(pipe, ctx.Pipelines().(*internal.PipelinesContext))
	return pipe
}

func (s *_bitflowScriptParser) buildInputPipelines(pipe *bitflow.SamplePipeline, ctx *internal.PipelinesContext) {
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
	case ctx.Pipelines() != nil:
		s.buildInputPipelines(pipe, ctx.Pipelines().(*internal.PipelinesContext))
	case ctx.DataInput() != nil:
		s.buildInput(pipe, ctx.DataInput().(*internal.DataInputContext))
	case ctx.PipelineElement() != nil:
		s.buildPipelineElement(pipe, ctx.PipelineElement().(*internal.PipelineElementContext))
	default:
		// This case is actually not possible due to the grammar. Just cover the case for the sake of completeness.
		pipe.Source = new(bitflow.EmptySampleSource)
	}
	s.buildPipelineTail(pipe, ctx.AllPipelineTailElement())
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
	case ctx.ProcessingStep() != nil:
		s.buildProcessingStep(pipe, ctx.ProcessingStep().(*internal.ProcessingStepContext))
	case ctx.Fork() != nil:
		s.buildFork(pipe, ctx.Fork().(*internal.ForkContext))
	case ctx.Window() != nil:
		s.buildWindow(pipe, ctx.Window().(*internal.WindowContext))
	}
}

func (s *_bitflowScriptParser) buildProcessingStep(pipe *bitflow.SamplePipeline, ctx *internal.ProcessingStepContext) {
	nameCtx := ctx.Name().(*internal.NameContext)
	name := unwrapString(nameCtx)
	params := s.buildParameters(ctx.Parameters().(*internal.ParametersContext))

	regAnalysis, ok := s.registry.GetStep(name)
	if !ok {
		// Check, if a batch or fork step step with that name exists.
		extraHelp := ""
		if _, ok := s.registry.GetBatchStep(name); ok {
			extraHelp = ", but a batch step with that name exists. Did you mean to use '" + name + "' inside the batch() {...} environment?"
		} else if _, ok := s.registry.GetFork(name); ok {
			extraHelp = ", but a fork step with that name exists. Did you mean to use '" + name + "' with attached sub-pipelines?"
		}
		s.pushError(nameCtx, "Processing step '%v' is unknown%s", name, extraHelp)
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

func (s *_bitflowScriptParser) buildParameters(ctx *internal.ParametersContext) map[string]string {
	params := make(map[string]string)
	if lst := ctx.ParameterList(); lst != nil {
		for _, paramCtxI := range lst.(*internal.ParameterListContext).AllParameter() {
			paramCtx := paramCtxI.(*internal.ParameterContext)
			key := unwrapString(paramCtx.Name(0).(*internal.NameContext))
			value := unwrapString(paramCtx.Name(1).(*internal.NameContext))
			params[key] = value
		}
	}
	return params
}

func (s *_bitflowScriptParser) buildFork(pipe *bitflow.SamplePipeline, ctx *internal.ForkContext) {
	nameCtx := ctx.Name().(*internal.NameContext)
	name := unwrapString(nameCtx)
	params := s.buildParameters(ctx.Parameters().(*internal.ParametersContext))

	// Lookup fork step and verify parameters
	forkStep, ok := s.registry.GetFork(name)
	if !ok {
		// Check, if a batch or non-batch step step with that name exists.
		extraHelp := ""
		if _, ok := s.registry.GetStep(name); ok {
			extraHelp = ", but a non-fork step with that name exists. Did you mean to use '" + name + "' without attached sub-pipelines?"
		} else if _, ok := s.registry.GetBatchStep(name); ok {
			extraHelp = ", but a batch step with that name exists. Did you mean to use '" + name + "' inside the batch() {...} environment and without attached sub-pipelines?"
		}
		s.pushError(nameCtx, "Pipeline fork '%v' is unknown%s", name, extraHelp)
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

func (s *_bitflowScriptParser) buildPipelineTail(pipe *bitflow.SamplePipeline, elements []internal.IPipelineTailElementContext) {
	for _, elemI := range elements {
		elem := elemI.(*internal.PipelineTailElementContext)
		switch {
		case elem.PipelineElement() != nil:
			s.buildPipelineElement(pipe, elem.PipelineElement().(*internal.PipelineElementContext))
		case elem.MultiplexFork() != nil:
			s.buildMultiplexFork(pipe, elem.MultiplexFork().(*internal.MultiplexForkContext))
		case elem.DataOutput() != nil:
			s.buildOutput(pipe, elem.DataOutput().(*internal.DataOutputContext))
		}
	}
}

func (s *_bitflowScriptParser) buildMultiplexFork(pipe *bitflow.SamplePipeline, ctx *internal.MultiplexForkContext) {
	var multi fork.MultiplexDistributor
	for _, subpipeCtx := range ctx.AllSubPipeline() {
		subpipe := new(bitflow.SamplePipeline)
		s.buildPipelineTail(subpipe, subpipeCtx.(*internal.SubPipelineContext).AllPipelineTailElement())
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
	batchParams := steps.MakeBatchProcessorParameters()
	params := s.buildParameters(ctx.Parameters().(*internal.ParametersContext))

	if err := batchParams.Verify(params); err != nil {
		s.pushError(ctx, "Invalid batch parameters: %v", err)
		return
	}
	batch, err := steps.MakeBatchProcessor(params)
	if err != nil {
		s.pushError(ctx, "Failed to create batch step: %v", err)
		return
	}
	for _, batchStepI := range ctx.AllProcessingStep() {
		batchStep := batchStepI.(*internal.ProcessingStepContext)
		stepName := unwrapString(batchStep.Name().(*internal.NameContext))
		registeredStep, ok := s.registry.GetBatchStep(stepName)
		if !ok {
			// Check, if a streaming step or fork step with that name exists.
			extraHelp := ""
			if _, ok := s.registry.GetStep(stepName); ok {
				extraHelp = ", but a non-batch step with that name exists. Did you mean to use '" + stepName + "' outside of the batch() {...} environment?"
			} else if _, ok := s.registry.GetFork(stepName); ok {
				extraHelp = ", but a fork step with that name exists. Did you mean to use '" + stepName + "' outside of the batch() {...} environment, as a fork?"
			}
			s.pushError(batchStep, "Batch step '%v' is unknown%s", stepName, extraHelp)
			return
		}
		stepParams := s.buildParameters(batchStep.Parameters().(*internal.ParametersContext))
		if err = registeredStep.Params.Verify(stepParams); err != nil {
			s.pushError(batchStep, "%v: %v", stepName, err)
			return
		}
		createdStep, err := registeredStep.Func(stepParams)
		if err != nil {
			s.pushError(batchStep, "%v: %v", stepName, err)
			return
		}
		batch.Add(createdStep)
	}
	pipe.Add(batch)
}
