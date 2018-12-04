// Code generated from Bitflow.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Bitflow
import "github.com/antlr/antlr4/runtime/Go/antlr"

type BaseBitflowVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseBitflowVisitor) VisitScript(ctx *ScriptContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitInput(ctx *InputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitOutput(ctx *OutputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitName(ctx *NameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitVal(ctx *ValContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitParameter(ctx *ParameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitTransformParameters(ctx *TransformParametersContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitPipeline(ctx *PipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitMultiInputPipeline(ctx *MultiInputPipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitPipelineElement(ctx *PipelineElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitTransform(ctx *TransformContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitFork(ctx *ForkContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitNamedSubPipeline(ctx *NamedSubPipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitSubPipeline(ctx *SubPipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitMultiplexFork(ctx *MultiplexForkContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitMultiplexSubPipeline(ctx *MultiplexSubPipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitWindow(ctx *WindowContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitWindowPipeline(ctx *WindowPipelineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitSchedulingHints(ctx *SchedulingHintsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseBitflowVisitor) VisitSchedulingParameter(ctx *SchedulingParameterContext) interface{} {
	return v.VisitChildren(ctx)
}
