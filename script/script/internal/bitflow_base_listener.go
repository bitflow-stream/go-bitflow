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

// EnterDataInput is called when production dataInput is entered.
func (s *BaseBitflowListener) EnterDataInput(ctx *DataInputContext) {}

// ExitDataInput is called when production dataInput is exited.
func (s *BaseBitflowListener) ExitDataInput(ctx *DataInputContext) {}

// EnterDataOutput is called when production dataOutput is entered.
func (s *BaseBitflowListener) EnterDataOutput(ctx *DataOutputContext) {}

// ExitDataOutput is called when production dataOutput is exited.
func (s *BaseBitflowListener) ExitDataOutput(ctx *DataOutputContext) {}

// EnterName is called when production name is entered.
func (s *BaseBitflowListener) EnterName(ctx *NameContext) {}

// ExitName is called when production name is exited.
func (s *BaseBitflowListener) ExitName(ctx *NameContext) {}

// EnterParameter is called when production parameter is entered.
func (s *BaseBitflowListener) EnterParameter(ctx *ParameterContext) {}

// ExitParameter is called when production parameter is exited.
func (s *BaseBitflowListener) ExitParameter(ctx *ParameterContext) {}

// EnterParameterValue is called when production parameterValue is entered.
func (s *BaseBitflowListener) EnterParameterValue(ctx *ParameterValueContext) {}

// ExitParameterValue is called when production parameterValue is exited.
func (s *BaseBitflowListener) ExitParameterValue(ctx *ParameterValueContext) {}

// EnterPrimitiveValue is called when production primitiveValue is entered.
func (s *BaseBitflowListener) EnterPrimitiveValue(ctx *PrimitiveValueContext) {}

// ExitPrimitiveValue is called when production primitiveValue is exited.
func (s *BaseBitflowListener) ExitPrimitiveValue(ctx *PrimitiveValueContext) {}

// EnterListValue is called when production listValue is entered.
func (s *BaseBitflowListener) EnterListValue(ctx *ListValueContext) {}

// ExitListValue is called when production listValue is exited.
func (s *BaseBitflowListener) ExitListValue(ctx *ListValueContext) {}

// EnterMapValue is called when production mapValue is entered.
func (s *BaseBitflowListener) EnterMapValue(ctx *MapValueContext) {}

// ExitMapValue is called when production mapValue is exited.
func (s *BaseBitflowListener) ExitMapValue(ctx *MapValueContext) {}

// EnterMapValueElement is called when production mapValueElement is entered.
func (s *BaseBitflowListener) EnterMapValueElement(ctx *MapValueElementContext) {}

// ExitMapValueElement is called when production mapValueElement is exited.
func (s *BaseBitflowListener) ExitMapValueElement(ctx *MapValueElementContext) {}

// EnterParameterList is called when production parameterList is entered.
func (s *BaseBitflowListener) EnterParameterList(ctx *ParameterListContext) {}

// ExitParameterList is called when production parameterList is exited.
func (s *BaseBitflowListener) ExitParameterList(ctx *ParameterListContext) {}

// EnterParameters is called when production parameters is entered.
func (s *BaseBitflowListener) EnterParameters(ctx *ParametersContext) {}

// ExitParameters is called when production parameters is exited.
func (s *BaseBitflowListener) ExitParameters(ctx *ParametersContext) {}

// EnterPipelines is called when production pipelines is entered.
func (s *BaseBitflowListener) EnterPipelines(ctx *PipelinesContext) {}

// ExitPipelines is called when production pipelines is exited.
func (s *BaseBitflowListener) ExitPipelines(ctx *PipelinesContext) {}

// EnterPipeline is called when production pipeline is entered.
func (s *BaseBitflowListener) EnterPipeline(ctx *PipelineContext) {}

// ExitPipeline is called when production pipeline is exited.
func (s *BaseBitflowListener) ExitPipeline(ctx *PipelineContext) {}

// EnterPipelineElement is called when production pipelineElement is entered.
func (s *BaseBitflowListener) EnterPipelineElement(ctx *PipelineElementContext) {}

// ExitPipelineElement is called when production pipelineElement is exited.
func (s *BaseBitflowListener) ExitPipelineElement(ctx *PipelineElementContext) {}

// EnterPipelineTailElement is called when production pipelineTailElement is entered.
func (s *BaseBitflowListener) EnterPipelineTailElement(ctx *PipelineTailElementContext) {}

// ExitPipelineTailElement is called when production pipelineTailElement is exited.
func (s *BaseBitflowListener) ExitPipelineTailElement(ctx *PipelineTailElementContext) {}

// EnterProcessingStep is called when production processingStep is entered.
func (s *BaseBitflowListener) EnterProcessingStep(ctx *ProcessingStepContext) {}

// ExitProcessingStep is called when production processingStep is exited.
func (s *BaseBitflowListener) ExitProcessingStep(ctx *ProcessingStepContext) {}

// EnterFork is called when production fork is entered.
func (s *BaseBitflowListener) EnterFork(ctx *ForkContext) {}

// ExitFork is called when production fork is exited.
func (s *BaseBitflowListener) ExitFork(ctx *ForkContext) {}

// EnterNamedSubPipeline is called when production namedSubPipeline is entered.
func (s *BaseBitflowListener) EnterNamedSubPipeline(ctx *NamedSubPipelineContext) {}

// ExitNamedSubPipeline is called when production namedSubPipeline is exited.
func (s *BaseBitflowListener) ExitNamedSubPipeline(ctx *NamedSubPipelineContext) {}

// EnterSubPipeline is called when production subPipeline is entered.
func (s *BaseBitflowListener) EnterSubPipeline(ctx *SubPipelineContext) {}

// ExitSubPipeline is called when production subPipeline is exited.
func (s *BaseBitflowListener) ExitSubPipeline(ctx *SubPipelineContext) {}

// EnterMultiplexFork is called when production multiplexFork is entered.
func (s *BaseBitflowListener) EnterMultiplexFork(ctx *MultiplexForkContext) {}

// ExitMultiplexFork is called when production multiplexFork is exited.
func (s *BaseBitflowListener) ExitMultiplexFork(ctx *MultiplexForkContext) {}

// EnterWindow is called when production window is entered.
func (s *BaseBitflowListener) EnterWindow(ctx *WindowContext) {}

// ExitWindow is called when production window is exited.
func (s *BaseBitflowListener) ExitWindow(ctx *WindowContext) {}

// EnterSchedulingHints is called when production schedulingHints is entered.
func (s *BaseBitflowListener) EnterSchedulingHints(ctx *SchedulingHintsContext) {}

// ExitSchedulingHints is called when production schedulingHints is exited.
func (s *BaseBitflowListener) ExitSchedulingHints(ctx *SchedulingHintsContext) {}
