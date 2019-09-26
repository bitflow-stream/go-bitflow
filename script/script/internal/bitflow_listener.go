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

	// EnterParameterValue is called when entering the parameterValue production.
	EnterParameterValue(c *ParameterValueContext)

	// EnterPrimitiveValue is called when entering the primitiveValue production.
	EnterPrimitiveValue(c *PrimitiveValueContext)

	// EnterListValue is called when entering the listValue production.
	EnterListValue(c *ListValueContext)

	// EnterMapValue is called when entering the mapValue production.
	EnterMapValue(c *MapValueContext)

	// EnterMapValueElement is called when entering the mapValueElement production.
	EnterMapValueElement(c *MapValueElementContext)

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

	// EnterBatchPipeline is called when entering the batchPipeline production.
	EnterBatchPipeline(c *BatchPipelineContext)

	// EnterMultiplexFork is called when entering the multiplexFork production.
	EnterMultiplexFork(c *MultiplexForkContext)

	// EnterBatch is called when entering the batch production.
	EnterBatch(c *BatchContext)

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

	// ExitParameterValue is called when exiting the parameterValue production.
	ExitParameterValue(c *ParameterValueContext)

	// ExitPrimitiveValue is called when exiting the primitiveValue production.
	ExitPrimitiveValue(c *PrimitiveValueContext)

	// ExitListValue is called when exiting the listValue production.
	ExitListValue(c *ListValueContext)

	// ExitMapValue is called when exiting the mapValue production.
	ExitMapValue(c *MapValueContext)

	// ExitMapValueElement is called when exiting the mapValueElement production.
	ExitMapValueElement(c *MapValueElementContext)

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

	// ExitBatchPipeline is called when exiting the batchPipeline production.
	ExitBatchPipeline(c *BatchPipelineContext)

	// ExitMultiplexFork is called when exiting the multiplexFork production.
	ExitMultiplexFork(c *MultiplexForkContext)

	// ExitBatch is called when exiting the batch production.
	ExitBatch(c *BatchContext)

	// ExitSchedulingHints is called when exiting the schedulingHints production.
	ExitSchedulingHints(c *SchedulingHintsContext)
}
