package steps

import (
	"fmt"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	log "github.com/sirupsen/logrus"
)

func RegisterDropErrorsStep(b *query.PipelineBuilder) {
	b.RegisterAnalysis("drop_errors",
		func(p *pipeline.SamplePipeline) {
			p.Add(new(DropErrorsProcessor))
		},
		"All errors of subsequent processing steps are only logged and not forwarded to the steps before")
}

type DropErrorsProcessor struct {
	bitflow.NoopProcessor
}

func (p *DropErrorsProcessor) String() string {
	return fmt.Sprintf("Drop errors of subsequent steps")
}

func (p *DropErrorsProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	err := p.NoopProcessor.Sample(sample, header)
	if err != nil {
		log.Errorln("(Dropped error)", err)
	}
	return nil
}
