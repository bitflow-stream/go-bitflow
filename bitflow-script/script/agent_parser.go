package script

// See README for details about implementation approach

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/reg"
	internal "github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/script/internal"
	"github.com/bitflow-stream/go-bitflow-pipeline/fork"
	"github.com/bugsnag/bugsnag-go/errors"
)

var emptyTokenMap = make(map[token]token)

type BitflowScriptParser interface {
	// ParseScript takes a Bitflow Script as input and returns the parsed SamplePipeline and occurred parsing errors as output
	ParseScript(script string) (*pipeline.SamplePipeline, golib.MultiError)
}

func NewAntlrBitflowParser(b reg.ProcessorRegistry) BitflowScriptParser {
	r := new(AntlrBitflowListener)
	r.endpointFactory = bitflow.NewEndpointFactory()
	r.registry = b
	r.resetListener()
	return r
}

type subpipeline struct {
	keys []string
	pipe *pipeline.SamplePipeline
}

func (s *subpipeline) Build() (*pipeline.SamplePipeline, error) {
	return s.pipe, nil
}

func (s *subpipeline) Keys() []string {
	return s.keys
}

type ParserError struct {
	Pos     token
	Message string
}

func (e ParserError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "Unknown parser error"
	}
	return fmt.Sprintf("%v (%v)", msg, e.Pos.String())
}

// next line to ensure AntlrBitflowListener satisfies BitflowListener interface
var _ internal.BitflowListener = &AntlrBitflowListener{}

// AntlrBitflowListener is a complete listener for a parse tree produced by BitflowParser.
type AntlrBitflowListener struct {
	internal.BaseBitflowListener
	state           StackMap
	endpointFactory *bitflow.EndpointFactory
	registry        reg.ProcessorRegistry
	errors          []error
}

func (s *AntlrBitflowListener) pushError(err error) {
	s.errors = append(s.errors, err)
}

func (s *AntlrBitflowListener) resetListener() {
	s.state = make(StackMap)
	s.errors = make([]error, 0)
}

// ParseScript takes a Bitflow Script as input and returns the parsed SamplePipeline and occurred parsing errors as output
func (s *AntlrBitflowListener) ParseScript(script string) (pipe *pipeline.SamplePipeline, errors golib.MultiError) {
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
	input := antlr.NewInputStream(rawScript)
	lexer := internal.NewBitflowLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := internal.NewBitflowParser(stream)
	p.AddErrorListener(antlr.NewConsoleErrorListener())
	p.BuildParseTrees = true
	tree := p.Script()
	antlr.ParseTreeWalkerDefault.Walk(s, tree)
	pipe = s.pipe()
	return pipe, s.errors
}

// ExitFork is called when production fork is exited.
func (s *AntlrBitflowListener) ExitFork(ctx *internal.ForkContext) {
	pipe := s.pipe()
	forkName := s.state.Pop("processor_name").(token)
	forkParams := s.state.Pop("params").(map[token]token)
	params := convertParams(forkParams)

	var subpipelines []*subpipeline
	s.state.PopAll("fork_subpipelines", &subpipelines)
	//
	// if forkName.Lit == "" {
	//	// default to multiplex, if no name is given
	//	forkName.Lit = MultiplexForkName
	//	params[MultiplexForkParam] = strconv.Itoa(len(subPipes))
	// }

	forkStep, ok := s.registry.GetFork(forkName.Lit)
	if !ok {
		s.pushError(ParserError{
			Pos:     forkName,
			Message: fmt.Sprintf("Pipeline fork '%v' is unknown", forkName.Lit),
		})
		return
	}
	var distributor fork.Distributor
	err := forkStep.Params.Verify(params)
	if err == nil {
		regSubpipelines := make([]reg.Subpipeline, len(subpipelines))
		for i, subpipeline := range subpipelines {
			regSubpipelines[i] = subpipeline
		}
		distributor, err = forkStep.Func(regSubpipelines, params)
	}
	if err != nil {
		s.pushError(ParserError{
			Pos:     forkName,
			Message: fmt.Sprintf("%v: %v", forkName.Content(), err),
		})
	}
	pipe.Add(&fork.SampleFork{
		Distributor: distributor,
	})
}

// ExitInput is called when production input is exited.
func (s *AntlrBitflowListener) ExitInput(ctx *internal.InputContext) {
	i := stripQuotes(s.state.Pop("processor_name").(token).Lit)

	if s.state.Peek("is_multi_input") != nil {
		s.state.Push("input_description", i)
	} else {
		source, err := s.endpointFactory.CreateInput(i)
		if err != nil {
			s.pushError(err)
			return
		}
		s.pipe().Source = source
	}
}

// EnterMultiinput is called when production multiinput is entered.
func (s *AntlrBitflowListener) EnterMultiinput(ctx *internal.MultiinputContext) {
	s.state.Push("is_multi_input", true)
}

// ExitMultiinput is called when production multiinput is exited.
func (s *AntlrBitflowListener) ExitMultiinput(ctx *internal.MultiinputContext) {
	var inputDescriptions []string
	s.state.PopAll("input_description", &inputDescriptions)
	s.state.Pop("is_multi_input") // reset
	source, err := s.endpointFactory.CreateInput(inputDescriptions...)
	if err != nil {
		s.pushError(err)
		return
	}
	s.pipe().Source = source
}

// EnterPipeline is called when production pipeline is entered.
func (s *AntlrBitflowListener) EnterPipeline(ctx *internal.PipelineContext) {
	s.state.Push("pipeline", new(pipeline.SamplePipeline))
}

// EnterSub_pipeline is called when production sub_pipeline is entered.
func (s *AntlrBitflowListener) EnterSubPipeline(ctx *internal.SubPipelineContext) {
	s.state.Push("pipeline", new(pipeline.SamplePipeline))
}

// ExitSub_pipeline is called when production sub_pipeline is exited.
func (s *AntlrBitflowListener) ExitSubPipeline(ctx *internal.SubPipelineContext) {
	var subpipeName string
	if name := s.state.Pop("pipeline_name"); name != nil {
		subpipeName = name.(token).Lit
	} else {
		subpipeName = strconv.Itoa(s.state.Len("fork_subpipelines") + 1)
	}
	pipe := s.state.Pop("pipeline").(*pipeline.SamplePipeline)
	subpipe := &subpipeline{keys: []string{subpipeName}, pipe: pipe}
	s.state.Push("fork_subpipelines", subpipe)
}

// ExitTransform is called when production transform is exited.
func (s *AntlrBitflowListener) ExitTransform(ctx *internal.TransformContext) {
	procName := s.state.Pop("processor_name").(token)

	transformParams := s.state.PopOrDefault("params", emptyTokenMap).(map[token]token)
	isBatched := s.state.PeekOrDefault("is_batched", false).(bool)

	regAnalysis, ok := s.registry.GetAnalysis(procName.Lit)
	if !ok {
		s.pushError(ParserError{
			Pos:     procName,
			Message: fmt.Sprintf("%v: %v", procName.Lit, "Unknown Processor."),
		})

		return
	} else if isBatched && !regAnalysis.SupportsBatchProcessing {
		s.pushError(ParserError{
			Pos:     procName,
			Message: fmt.Sprintf("%v: %v", procName.Lit, "Processor used in window, but does not support batch processing."),
		})

		return
	} else if !isBatched && !regAnalysis.SupportsStreamProcessing {
		s.pushError(ParserError{
			Pos:     procName,
			Message: fmt.Sprintf("%v: %v", procName.Lit, "Processor used outside window, but does not support stream processing."),
		})

		return
	}

	params := convertParams(transformParams)
	err := regAnalysis.Params.Verify(params)
	if err == nil {
		err = regAnalysis.Func(s.pipe(), params)
	}

	if err != nil {
		s.pushError(ParserError{
			Pos:     procName,
			Message: fmt.Sprintf("%v: %v", procName.Lit, err),
		})
		return
	}

}

// ExitParameter is called when production parameter is exited.
func (s *AntlrBitflowListener) ExitParameter(ctx *internal.ParameterContext) {
	if ctx.GetChildCount() != 3 {
		msg := fmt.Sprintf("Invalid parameter. "+
			"Parameters must be given in the form of <key>=<value>, but was %v", ctx.GetText())
		s.pushError(errors.New(msg, 0))
	}
	key := ctx.GetChild(0).(*antlr.TerminalNodeImpl)
	val := ctx.GetChild(2).(*antlr.TerminalNodeImpl)
	k := token{
		Start: key.GetSourceInterval().Start,
		End:   key.GetSourceInterval().Stop,
		Lit:   key.GetText(),
	}
	v := token{
		Start: val.GetSourceInterval().Start,
		End:   val.GetSourceInterval().Stop,
		Lit:   val.GetText(),
	}

	s.state.Peek("params").(map[token]token)[k] = v
}

// EnterTransform_parameters is called when production transform_parameters is entered.
func (s *AntlrBitflowListener) EnterTransformParameters(ctx *internal.TransformParametersContext) {
	s.state.Push("params", make(map[token]token))
}

// ExitName is called when production name is exited.
func (s *AntlrBitflowListener) ExitName(ctx *internal.NameContext) {
	s.state.Push("processor_name", token{
		Start: ctx.GetStart().GetStart(),
		End:   ctx.GetStop().GetStop(),
		Lit:   strings.TrimSpace(ctx.GetText()),
	})
}

// ExitOutput is called when production output is exited.
func (s *AntlrBitflowListener) ExitOutput(ctx *internal.OutputContext) {
	o := stripQuotes(s.state.Pop("processor_name").(token).Lit)
	output, err := s.endpointFactory.CreateOutput(o)
	if err != nil {
		s.pushError(err)
		return
	}
	s.pipe().Add(output)
}

// ExitOutputFork is called when production outputFork is exited.
func (s *AntlrBitflowListener) ExitOutputFork(ctx *internal.OutputForkContext) {
	// Nothing to do, all Outputs are chained to the end of pipeline
}

// ExitPipeline_name is called when production pipeline_name is exited.
func (s *AntlrBitflowListener) ExitPipelineName(ctx *internal.PipelineNameContext) {
	s.state.Push("pipeline_name", token{
		Start: ctx.GetStart().GetStart(),
		End:   ctx.GetStop().GetStop(),
		Lit:   ctx.GetText(),
	})
}

func (s *AntlrBitflowListener) pipe() *pipeline.SamplePipeline {
	return s.state.Peek("pipeline").(*pipeline.SamplePipeline)
}

func stripQuotes(in string) string {
	if strings.HasPrefix(in, "\"") {
		in = in[1:]
	}
	if strings.HasSuffix(in, "\"") {
		in = in[:len(in)-1]
	}
	return in
}

// EnterWindow is called when production window is entered.
func (s *AntlrBitflowListener) EnterWindow(ctx *internal.WindowContext) {
	isBatched := s.state.PeekOrDefault("is_batched", false).(bool)
	if isBatched {
		t := token{
			Start: ctx.GetStart().GetStart(),
			End:   ctx.GetStop().GetStop(),
			Lit:   strings.TrimSpace(ctx.GetText()),
		}
		s.pushError(ParserError{
			Pos:     t,
			Message: fmt.Sprintf("window{ ... }: %v", "Window inside Window is not allowed."),
		})
	}
	s.state.Push("is_batched", true)
}

// ExitWindow is called when production window is exited.
func (s *AntlrBitflowListener) ExitWindow(ctx *internal.WindowContext) {
	s.state.Pop("is_batched")
}
