// Code generated from Bitflow.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Bitflow
import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseBitflowListener is a complete listener for a parse tree produced by BitflowParser.
type BaseBitflowListener struct{}

var _ BitflowListener = &BaseBitflowListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseBitflowListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseBitflowListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseBitflowListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseBitflowListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterScript is called when production script is entered.
func (s *BaseBitflowListener) EnterScript(ctx *ScriptContext) {}

// ExitScript is called when production script is exited.
func (s *BaseBitflowListener) ExitScript(ctx *ScriptContext) {}

// EnterOutputFork is called when production outputFork is entered.
func (s *BaseBitflowListener) EnterOutputFork(ctx *OutputForkContext) {}

// ExitOutputFork is called when production outputFork is exited.
func (s *BaseBitflowListener) ExitOutputFork(ctx *OutputForkContext) {}

// EnterFork is called when production fork is entered.
func (s *BaseBitflowListener) EnterFork(ctx *ForkContext) {}

// ExitFork is called when production fork is exited.
func (s *BaseBitflowListener) ExitFork(ctx *ForkContext) {}

// EnterWindow is called when production window is entered.
func (s *BaseBitflowListener) EnterWindow(ctx *WindowContext) {}

// ExitWindow is called when production window is exited.
func (s *BaseBitflowListener) ExitWindow(ctx *WindowContext) {}

// EnterMultiinput is called when production multiinput is entered.
func (s *BaseBitflowListener) EnterMultiinput(ctx *MultiinputContext) {}

// ExitMultiinput is called when production multiinput is exited.
func (s *BaseBitflowListener) ExitMultiinput(ctx *MultiinputContext) {}

// EnterInput is called when production input is entered.
func (s *BaseBitflowListener) EnterInput(ctx *InputContext) {}

// ExitInput is called when production input is exited.
func (s *BaseBitflowListener) ExitInput(ctx *InputContext) {}

// EnterOutput is called when production output is entered.
func (s *BaseBitflowListener) EnterOutput(ctx *OutputContext) {}

// ExitOutput is called when production output is exited.
func (s *BaseBitflowListener) ExitOutput(ctx *OutputContext) {}

// EnterTransform is called when production transform is entered.
func (s *BaseBitflowListener) EnterTransform(ctx *TransformContext) {}

// ExitTransform is called when production transform is exited.
func (s *BaseBitflowListener) ExitTransform(ctx *TransformContext) {}

// EnterSubPipeline is called when production subPipeline is entered.
func (s *BaseBitflowListener) EnterSubPipeline(ctx *SubPipelineContext) {}

// ExitSubPipeline is called when production subPipeline is exited.
func (s *BaseBitflowListener) ExitSubPipeline(ctx *SubPipelineContext) {}

// EnterWindowSubPipeline is called when production windowSubPipeline is entered.
func (s *BaseBitflowListener) EnterWindowSubPipeline(ctx *WindowSubPipelineContext) {}

// ExitWindowSubPipeline is called when production windowSubPipeline is exited.
func (s *BaseBitflowListener) ExitWindowSubPipeline(ctx *WindowSubPipelineContext) {}

// EnterPipeline is called when production pipeline is entered.
func (s *BaseBitflowListener) EnterPipeline(ctx *PipelineContext) {}

// ExitPipeline is called when production pipeline is exited.
func (s *BaseBitflowListener) ExitPipeline(ctx *PipelineContext) {}

// EnterParameter is called when production parameter is entered.
func (s *BaseBitflowListener) EnterParameter(ctx *ParameterContext) {}

// ExitParameter is called when production parameter is exited.
func (s *BaseBitflowListener) ExitParameter(ctx *ParameterContext) {}

// EnterTransformParameters is called when production transformParameters is entered.
func (s *BaseBitflowListener) EnterTransformParameters(ctx *TransformParametersContext) {}

// ExitTransformParameters is called when production transformParameters is exited.
func (s *BaseBitflowListener) ExitTransformParameters(ctx *TransformParametersContext) {}

// EnterName is called when production name is entered.
func (s *BaseBitflowListener) EnterName(ctx *NameContext) {}

// ExitName is called when production name is exited.
func (s *BaseBitflowListener) ExitName(ctx *NameContext) {}

// EnterPipelineName is called when production pipelineName is entered.
func (s *BaseBitflowListener) EnterPipelineName(ctx *PipelineNameContext) {}

// ExitPipelineName is called when production pipelineName is exited.
func (s *BaseBitflowListener) ExitPipelineName(ctx *PipelineNameContext) {}

// EnterSchedulingHints is called when production schedulingHints is entered.
func (s *BaseBitflowListener) EnterSchedulingHints(ctx *SchedulingHintsContext) {}

// ExitSchedulingHints is called when production schedulingHints is exited.
func (s *BaseBitflowListener) ExitSchedulingHints(ctx *SchedulingHintsContext) {}
