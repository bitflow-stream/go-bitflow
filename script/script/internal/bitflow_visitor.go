// Code generated from Bitflow.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Bitflow
import "github.com/antlr/antlr4/runtime/Go/antlr"

// A complete Visitor for a parse tree produced by BitflowParser.
type BitflowVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by BitflowParser#script.
	VisitScript(ctx *ScriptContext) interface{}

	// Visit a parse tree produced by BitflowParser#input.
	VisitInput(ctx *InputContext) interface{}

	// Visit a parse tree produced by BitflowParser#output.
	VisitOutput(ctx *OutputContext) interface{}

	// Visit a parse tree produced by BitflowParser#name.
	VisitName(ctx *NameContext) interface{}

	// Visit a parse tree produced by BitflowParser#val.
	VisitVal(ctx *ValContext) interface{}

	// Visit a parse tree produced by BitflowParser#parameter.
	VisitParameter(ctx *ParameterContext) interface{}

	// Visit a parse tree produced by BitflowParser#transformParameters.
	VisitTransformParameters(ctx *TransformParametersContext) interface{}

	// Visit a parse tree produced by BitflowParser#pipeline.
	VisitPipeline(ctx *PipelineContext) interface{}

	// Visit a parse tree produced by BitflowParser#multiInputPipeline.
	VisitMultiInputPipeline(ctx *MultiInputPipelineContext) interface{}

	// Visit a parse tree produced by BitflowParser#pipelineElement.
	VisitPipelineElement(ctx *PipelineElementContext) interface{}

	// Visit a parse tree produced by BitflowParser#transform.
	VisitTransform(ctx *TransformContext) interface{}

	// Visit a parse tree produced by BitflowParser#fork.
	VisitFork(ctx *ForkContext) interface{}

	// Visit a parse tree produced by BitflowParser#namedSubPipeline.
	VisitNamedSubPipeline(ctx *NamedSubPipelineContext) interface{}

	// Visit a parse tree produced by BitflowParser#subPipeline.
	VisitSubPipeline(ctx *SubPipelineContext) interface{}

	// Visit a parse tree produced by BitflowParser#multiplexFork.
	VisitMultiplexFork(ctx *MultiplexForkContext) interface{}

	// Visit a parse tree produced by BitflowParser#multiplexSubPipeline.
	VisitMultiplexSubPipeline(ctx *MultiplexSubPipelineContext) interface{}

	// Visit a parse tree produced by BitflowParser#window.
	VisitWindow(ctx *WindowContext) interface{}

	// Visit a parse tree produced by BitflowParser#windowPipeline.
	VisitWindowPipeline(ctx *WindowPipelineContext) interface{}

	// Visit a parse tree produced by BitflowParser#schedulingHints.
	VisitSchedulingHints(ctx *SchedulingHintsContext) interface{}

	// Visit a parse tree produced by BitflowParser#schedulingParameter.
	VisitSchedulingParameter(ctx *SchedulingParameterContext) interface{}
}
