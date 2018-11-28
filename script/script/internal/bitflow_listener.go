// Code generated from Bitflow.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Bitflow
import "github.com/antlr/antlr4/runtime/Go/antlr"

// BitflowListener is a complete listener for a parse tree produced by BitflowParser.
type BitflowListener interface {
	antlr.ParseTreeListener

	// EnterScript is called when entering the script production.
	EnterScript(c *ScriptContext)

	// EnterOutputFork is called when entering the outputFork production.
	EnterOutputFork(c *OutputForkContext)

	// EnterFork is called when entering the fork production.
	EnterFork(c *ForkContext)

	// EnterWindow is called when entering the window production.
	EnterWindow(c *WindowContext)

	// EnterMultiinput is called when entering the multiinput production.
	EnterMultiinput(c *MultiinputContext)

	// EnterInput is called when entering the input production.
	EnterInput(c *InputContext)

	// EnterOutput is called when entering the output production.
	EnterOutput(c *OutputContext)

	// EnterTransform is called when entering the transform production.
	EnterTransform(c *TransformContext)

	// EnterSubPipeline is called when entering the subPipeline production.
	EnterSubPipeline(c *SubPipelineContext)

	// EnterWindowSubPipeline is called when entering the windowSubPipeline production.
	EnterWindowSubPipeline(c *WindowSubPipelineContext)

	// EnterPipeline is called when entering the pipeline production.
	EnterPipeline(c *PipelineContext)

	// EnterParameter is called when entering the parameter production.
	EnterParameter(c *ParameterContext)

	// EnterTransformParameters is called when entering the transformParameters production.
	EnterTransformParameters(c *TransformParametersContext)

	// EnterName is called when entering the name production.
	EnterName(c *NameContext)

	// EnterPipelineName is called when entering the pipelineName production.
	EnterPipelineName(c *PipelineNameContext)

	// EnterSchedulingHints is called when entering the schedulingHints production.
	EnterSchedulingHints(c *SchedulingHintsContext)

	// ExitScript is called when exiting the script production.
	ExitScript(c *ScriptContext)

	// ExitOutputFork is called when exiting the outputFork production.
	ExitOutputFork(c *OutputForkContext)

	// ExitFork is called when exiting the fork production.
	ExitFork(c *ForkContext)

	// ExitWindow is called when exiting the window production.
	ExitWindow(c *WindowContext)

	// ExitMultiinput is called when exiting the multiinput production.
	ExitMultiinput(c *MultiinputContext)

	// ExitInput is called when exiting the input production.
	ExitInput(c *InputContext)

	// ExitOutput is called when exiting the output production.
	ExitOutput(c *OutputContext)

	// ExitTransform is called when exiting the transform production.
	ExitTransform(c *TransformContext)

	// ExitSubPipeline is called when exiting the subPipeline production.
	ExitSubPipeline(c *SubPipelineContext)

	// ExitWindowSubPipeline is called when exiting the windowSubPipeline production.
	ExitWindowSubPipeline(c *WindowSubPipelineContext)

	// ExitPipeline is called when exiting the pipeline production.
	ExitPipeline(c *PipelineContext)

	// ExitParameter is called when exiting the parameter production.
	ExitParameter(c *ParameterContext)

	// ExitTransformParameters is called when exiting the transformParameters production.
	ExitTransformParameters(c *TransformParametersContext)

	// ExitName is called when exiting the name production.
	ExitName(c *NameContext)

	// ExitPipelineName is called when exiting the pipelineName production.
	ExitPipelineName(c *PipelineNameContext)

	// ExitSchedulingHints is called when exiting the schedulingHints production.
	ExitSchedulingHints(c *SchedulingHintsContext)
}
