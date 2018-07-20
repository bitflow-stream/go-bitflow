package steps

import (
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

func RegisterGenericBatch(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) {
		p.Add(&pipeline.BatchProcessor{
			FlushTag: params["tag"],
		})
	}

	b.RegisterAnalysisParams("batch", create, "Collect samples and flush them when the given tag changes its value. Affects the follow-up analysis step, if it is also a batch analysis", []string{"tag"})
}
