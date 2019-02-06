// Code generated from Bitflow.g4 by ANTLR 4.7.1. DO NOT EDIT.

package parser // Bitflow
import "github.com/antlr/antlr4/runtime/Go/antlr"

// BitflowListener is a complete listener for a parse tree produced by BitflowParser.
type BitflowListener interface {
	antlr.ParseTreeListener

	// EnterScript is called when entering the script production.
	EnterScript(c *ScriptContext)

	// EnterDataInput is called when entering the dataInput production.
	EnterDataInput(c *DataInputContext)

	// EnterDataOutput is called when entering the dataOutput production.
	EnterDataOutput(c *DataOutputContext)

	// EnterName is called when entering the name production.
	EnterName(c *NameContext)

	// EnterParameter is called when entering the parameter production.
	EnterParameter(c *ParameterContext)

	// EnterParameterList is called when entering the parameterList production.
	EnterParameterList(c *ParameterListContext)

	// EnterParameters is called when entering the parameters production.
	EnterParameters(c *ParametersContext)

	// EnterPipelines is called when entering the pipelines production.
	EnterPipelines(c *PipelinesContext)

	// EnterPipeline is called when entering the pipeline production.
	EnterPipeline(c *PipelineContext)

	// EnterPipelineElement is called when entering the pipelineElement production.
	EnterPipelineElement(c *PipelineElementContext)

	// EnterPipelineTailElement is called when entering the pipelineTailElement production.
	EnterPipelineTailElement(c *PipelineTailElementContext)

	// EnterProcessingStep is called when entering the processingStep production.
	EnterProcessingStep(c *ProcessingStepContext)

	// EnterFork is called when entering the fork production.
	EnterFork(c *ForkContext)

	// EnterNamedSubPipeline is called when entering the namedSubPipeline production.
	EnterNamedSubPipeline(c *NamedSubPipelineContext)

	// EnterSubPipeline is called when entering the subPipeline production.
	EnterSubPipeline(c *SubPipelineContext)

	// EnterMultiplexFork is called when entering the multiplexFork production.
	EnterMultiplexFork(c *MultiplexForkContext)

	// EnterWindow is called when entering the window production.
	EnterWindow(c *WindowContext)

	// EnterSchedulingHints is called when entering the schedulingHints production.
	EnterSchedulingHints(c *SchedulingHintsContext)

	// ExitScript is called when exiting the script production.
	ExitScript(c *ScriptContext)

	// ExitDataInput is called when exiting the dataInput production.
	ExitDataInput(c *DataInputContext)

	// ExitDataOutput is called when exiting the dataOutput production.
	ExitDataOutput(c *DataOutputContext)

	// ExitName is called when exiting the name production.
	ExitName(c *NameContext)

	// ExitParameter is called when exiting the parameter production.
	ExitParameter(c *ParameterContext)

	// ExitParameterList is called when exiting the parameterList production.
	ExitParameterList(c *ParameterListContext)

	// ExitParameters is called when exiting the parameters production.
	ExitParameters(c *ParametersContext)

	// ExitPipelines is called when exiting the pipelines production.
	ExitPipelines(c *PipelinesContext)

	// ExitPipeline is called when exiting the pipeline production.
	ExitPipeline(c *PipelineContext)

	// ExitPipelineElement is called when exiting the pipelineElement production.
	ExitPipelineElement(c *PipelineElementContext)

	// ExitPipelineTailElement is called when exiting the pipelineTailElement production.
	ExitPipelineTailElement(c *PipelineTailElementContext)

	// ExitProcessingStep is called when exiting the processingStep production.
	ExitProcessingStep(c *ProcessingStepContext)

	// ExitFork is called when exiting the fork production.
	ExitFork(c *ForkContext)

	// ExitNamedSubPipeline is called when exiting the namedSubPipeline production.
	ExitNamedSubPipeline(c *NamedSubPipelineContext)

	// ExitSubPipeline is called when exiting the subPipeline production.
	ExitSubPipeline(c *SubPipelineContext)

	// ExitMultiplexFork is called when exiting the multiplexFork production.
	ExitMultiplexFork(c *MultiplexForkContext)

	// ExitWindow is called when exiting the window production.
	ExitWindow(c *WindowContext)

	// ExitSchedulingHints is called when exiting the schedulingHints production.
	ExitSchedulingHints(c *SchedulingHintsContext)
}
