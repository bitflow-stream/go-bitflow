package recovery

import (
	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

type RecoveryEvaluationProcessor struct {
	bitflow.NoopProcessor
	RecoveryTags
}

func RegisterRecoveryEvaluation(b *query.PipelineBuilder) {
	b.RegisterAnalysisParams("recovery-evaluation", func(p *pipeline.SamplePipeline, params map[string]string) {
		step := &RecoveryEvaluationProcessor{
		// TODO
		}
		step.ParseRecoveryTags(params)
		p.Add(step)
	}, "Evaluation procedure for the recovery() step. Fully reads all input data before performing the evaluation.",
		append([]string{
		// TODO
		}, RecoveryTagParams...))
}

func (p *RecoveryEvaluationProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	node, state := p.GetRecoveryTags(sample)
	if node != "" && state != "" {
		p.storeEvaluationSample(node, state, sample, header)
	}
	return p.NoopProcessor.Sample(sample, header)
}

func (p *RecoveryEvaluationProcessor) storeEvaluationSample(node, state string, sample *bitflow.Sample, header *bitflow.Header) {
	// TODO store sample
}

func (p *RecoveryEvaluationProcessor) Close() {
	// TODO perform evaluation
	p.NoopProcessor.Close()
}
