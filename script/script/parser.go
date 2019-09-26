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
	case ctx.Batch() != nil:
		s.buildBatch(pipe, ctx.Batch().(*internal.BatchContext))
	}
}

func (s *_bitflowScriptParser) buildProcessingStep(pipe *bitflow.SamplePipeline, ctx *internal.ProcessingStepContext) {
	nameCtx := ctx.Name().(*internal.NameContext)
	name := unwrapString(nameCtx)
	regAnalysis := s.registry.GetStep(name)
	if regAnalysis == nil {
		// Check, if a batch or fork step step with that name exists.
		extraHelp := ""
		if s.registry.GetBatchStep(name) != nil {
			extraHelp = ", but a batch step with that name exists. Did you mean to use '" + name + "' inside the batch() {...} environment?"
		} else if s.registry.GetFork(name) != nil {
			extraHelp = ", but a fork step with that name exists. Did you mean to use '" + name + "' with attached sub-pipelines?"
		}
		s.pushError(nameCtx, "Processing step '%v' is unknown%s", name, extraHelp)
		return
	}

	params, err := s.buildParameters(regAnalysis.Params, ctx.Parameters().(*internal.ParametersContext))
	if err == nil {
		err = s.safeExecute(func() error {
			return regAnalysis.Func(pipe, params)
		})
	}
	if err != nil {
		s.pushError(nameCtx, "%v: %v", name, err)
	}
}

func (s *_bitflowScriptParser) buildParameters(params reg.RegisteredParameters, ctx *internal.ParametersContext) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if lst := ctx.ParameterList(); lst != nil {
		for _, paramCtxI := range lst.(*internal.ParameterListContext).AllParameter() {
			paramCtx := paramCtxI.(*internal.ParameterContext)
			key := unwrapString(paramCtx.Name().(*internal.NameContext))
			param, ok := params[key]
			if !ok {
				return nil, fmt.Errorf("Unknown parameter '%v'", key)
			}
			if _, defined := result[key]; defined {
				return nil, fmt.Errorf("Parameter '%v' is redefined", key)
			}
			valCtx := paramCtx.ParameterValue().(*internal.ParameterValueContext)

			var value interface{}
			var err error
			switch {
			case valCtx.PrimitiveValue() != nil:
				primitiveVal := valCtx.PrimitiveValue().(*internal.PrimitiveValueContext).Name()
				stringValue := unwrapString(primitiveVal.(*internal.NameContext))
				value, err = param.Parser.ParsePrimitive(stringValue)
			case valCtx.ListValue() != nil:
				primitiveValues := valCtx.ListValue().(*internal.ListValueContext).AllPrimitiveValue()
				stringValues := make([]string, len(primitiveValues))
				for i, primitive := range primitiveValues {
					stringValues[i] = unwrapString(primitive.(*internal.PrimitiveValueContext).Name().(*internal.NameContext))
				}
				value, err = param.Parser.ParseList(stringValues)
			case valCtx.MapValue() != nil:
				primitiveValues := valCtx.MapValue().(*internal.MapValueContext).AllMapValueElement()
				stringValues := make(map[string]string, len(primitiveValues))
				for _, primitive := range primitiveValues {
					mapElement := primitive.(*internal.MapValueElementContext)
					key := unwrapString(mapElement.Name().(*internal.NameContext))
					value := unwrapString(mapElement.PrimitiveValue().(*internal.PrimitiveValueContext).Name().(*internal.NameContext))
					stringValues[key] = value
				}
				value, err = param.Parser.ParseMap(stringValues)
			}
			if err != nil {
				return nil, err
			}
			result[key] = value
		}
	}
	return result, params.ValidateAndSetDefaults(result)
}

func (s *_bitflowScriptParser) buildFork(pipe *bitflow.SamplePipeline, ctx *internal.ForkContext) {
	nameCtx := ctx.Name().(*internal.NameContext)
	name := unwrapString(nameCtx)

	// Lookup fork step and verify parameters
	forkStep := s.registry.GetFork(name)
	if forkStep == nil {
		// Check, if a batch or non-batch step step with that name exists.
		extraHelp := ""
		if s.registry.GetStep(name) != nil {
			extraHelp = ", but a non-fork step with that name exists. Did you mean to use '" + name + "' without attached sub-pipelines?"
		} else if s.registry.GetBatchStep(name) != nil {
			extraHelp = ", but a batch step with that name exists. Did you mean to use '" + name + "' inside the batch() {...} environment and without attached sub-pipelines?"
		}
		s.pushError(nameCtx, "Pipeline fork '%v' is unknown%s", name, extraHelp)
		return
	}

	params, err := s.buildParameters(forkStep.Params, ctx.Parameters().(*internal.ParametersContext))
	if err != nil {
		s.pushError(nameCtx, "%v: %v", name, err)
		return
	}

	// Parse sub-pipelines
	subpipelines := make([]reg.Subpipeline, len(ctx.AllNamedSubPipeline()))
	for i, namedSubPipe := range ctx.AllNamedSubPipeline() {
		subpipelines[i] = s.buildNamedSubPipeline(namedSubPipe.(*internal.NamedSubPipelineContext))
	}

	var distributor fork.Distributor
	err = s.safeExecute(func() (err error) {
		distributor, err = forkStep.Func(subpipelines, params)
		return
	})
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

func (s *_bitflowScriptParser) buildBatch(pipe *bitflow.SamplePipeline, ctx *internal.BatchContext) {
	params, err := s.buildParameters(steps.BatchProcessorParameters, ctx.Parameters().(*internal.ParametersContext))
	if err != nil {
		s.pushError(ctx, "Invalid batch parameters: %v", err)
		return
	}
	batch, err := steps.MakeBatchProcessor(params)
	if err != nil {
		s.pushError(ctx, "Failed to create batch step: %v", err)
		return
	}

	allSteps := ctx.BatchPipeline().(*internal.BatchPipelineContext).AllProcessingStep()
	for _, batchStepI := range allSteps {
		batchStep := batchStepI.(*internal.ProcessingStepContext)
		stepName := unwrapString(batchStep.Name().(*internal.NameContext))
		registeredStep := s.registry.GetBatchStep(stepName)
		if registeredStep == nil {
			// Check, if a streaming step or fork step with that name exists.
			extraHelp := ""
			if s.registry.GetStep(stepName) != nil {
				extraHelp = ", but a non-batch step with that name exists. Did you mean to use '" + stepName + "' outside of the batch() {...} environment?"
			} else if s.registry.GetFork(stepName) != nil {
				extraHelp = ", but a fork step with that name exists. Did you mean to use '" + stepName + "' outside of the batch() {...} environment, as a fork?"
			}
			s.pushError(batchStep, "Batch step '%v' is unknown%s", stepName, extraHelp)
			return
		}
		stepParams, err := s.buildParameters(registeredStep.Params, batchStep.Parameters().(*internal.ParametersContext))
		if err != nil {
			s.pushError(batchStep, "%v: %v", stepName, err)
			return
		}

		var createdStep bitflow.BatchProcessingStep
		err = s.safeExecute(func() (err error) {
			createdStep, err = registeredStep.Func(stepParams)
			return
		})
		if err != nil {
			s.pushError(batchStep, "%v: %v", stepName, err)
			return
		}
		batch.Add(createdStep)
	}
	pipe.Add(batch)
}

func (s *_bitflowScriptParser) safeExecute(do func() error) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			recErr := fmt.Errorf("Recovered panic: %v", rec)
			if err != nil {
				err = golib.MultiError{err, recErr}
			} else {
				err = recErr
			}
		}
	}()
	err = do()
	return
}
