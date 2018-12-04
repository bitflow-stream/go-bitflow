// Code generated from Bitflow.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Bitflow
import "github.com/antlr/antlr4/runtime/Go/antlr"

// BitflowListener is a complete listener for a parse tree produced by BitflowParser.
type BitflowListener interface {
	antlr.ParseTreeListener

	// EnterScript is called when entering the script production.
	EnterScript(c *ScriptContext)

	// EnterInput is called when entering the input production.
	EnterInput(c *InputContext)

	// EnterOutput is called when entering the output production.
	EnterOutput(c *OutputContext)

	// EnterName is called when entering the name production.
	EnterName(c *NameContext)

	// EnterVal is called when entering the val production.
	EnterVal(c *ValContext)

	// EnterParameter is called when entering the parameter production.
	EnterParameter(c *ParameterContext)

	// EnterTransformParameters is called when entering the transformParameters production.
	EnterTransformParameters(c *TransformParametersContext)

	// EnterPipeline is called when entering the pipeline production.
	EnterPipeline(c *PipelineContext)

	// EnterMultiInputPipeline is called when entering the multiInputPipeline production.
	EnterMultiInputPipeline(c *MultiInputPipelineContext)

	// EnterPipelineElement is called when entering the pipelineElement production.
	EnterPipelineElement(c *PipelineElementContext)

	// EnterTransform is called when entering the transform production.
	EnterTransform(c *TransformContext)

	// EnterFork is called when entering the fork production.
	EnterFork(c *ForkContext)

	// EnterNamedSubPipeline is called when entering the namedSubPipeline production.
	EnterNamedSubPipeline(c *NamedSubPipelineContext)

	// EnterSubPipeline is called when entering the subPipeline production.
	EnterSubPipeline(c *SubPipelineContext)

	// EnterMultiplexFork is called when entering the multiplexFork production.
	EnterMultiplexFork(c *MultiplexForkContext)

	// EnterMultiplexSubPipeline is called when entering the multiplexSubPipeline production.
	EnterMultiplexSubPipeline(c *MultiplexSubPipelineContext)

	// EnterWindow is called when entering the window production.
	EnterWindow(c *WindowContext)

	// EnterWindowPipeline is called when entering the windowPipeline production.
	EnterWindowPipeline(c *WindowPipelineContext)

	// EnterSchedulingHints is called when entering the schedulingHints production.
	EnterSchedulingHints(c *SchedulingHintsContext)

	// EnterSchedulingParameter is called when entering the schedulingParameter production.
	EnterSchedulingParameter(c *SchedulingParameterContext)

	// ExitScript is called when exiting the script production.
	ExitScript(c *ScriptContext)

	// ExitInput is called when exiting the input production.
	ExitInput(c *InputContext)

	// ExitOutput is called when exiting the output production.
	ExitOutput(c *OutputContext)

	// ExitName is called when exiting the name production.
	ExitName(c *NameContext)

	// ExitVal is called when exiting the val production.
	ExitVal(c *ValContext)

	// ExitParameter is called when exiting the parameter production.
	ExitParameter(c *ParameterContext)

	// ExitTransformParameters is called when exiting the transformParameters production.
	ExitTransformParameters(c *TransformParametersContext)

	// ExitPipeline is called when exiting the pipeline production.
	ExitPipeline(c *PipelineContext)

	// ExitMultiInputPipeline is called when exiting the multiInputPipeline production.
	ExitMultiInputPipeline(c *MultiInputPipelineContext)

	// ExitPipelineElement is called when exiting the pipelineElement production.
	ExitPipelineElement(c *PipelineElementContext)

	// ExitTransform is called when exiting the transform production.
	ExitTransform(c *TransformContext)

	// ExitFork is called when exiting the fork production.
	ExitFork(c *ForkContext)

	// ExitNamedSubPipeline is called when exiting the namedSubPipeline production.
	ExitNamedSubPipeline(c *NamedSubPipelineContext)

	// ExitSubPipeline is called when exiting the subPipeline production.
	ExitSubPipeline(c *SubPipelineContext)

	// ExitMultiplexFork is called when exiting the multiplexFork production.
	ExitMultiplexFork(c *MultiplexForkContext)

	// ExitMultiplexSubPipeline is called when exiting the multiplexSubPipeline production.
	ExitMultiplexSubPipeline(c *MultiplexSubPipelineContext)

	// ExitWindow is called when exiting the window production.
	ExitWindow(c *WindowContext)

	// ExitWindowPipeline is called when exiting the windowPipeline production.
	ExitWindowPipeline(c *WindowPipelineContext)

	// ExitSchedulingHints is called when exiting the schedulingHints production.
	ExitSchedulingHints(c *SchedulingHintsContext)

	// ExitSchedulingParameter is called when exiting the schedulingParameter production.
	ExitSchedulingParameter(c *SchedulingParameterContext)
}
