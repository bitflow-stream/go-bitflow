package steps

import (
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
)

func RegisterGenericBatch(b builder.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("batch",
		func(p *pipeline.SamplePipeline, params map[string]string) (err error) {
			timeout := builder.DurationParam(params, "timeout", 0, true, &err)
			if err == nil {
				p.Add(&pipeline.BatchProcessor{
					FlushTag:     params["tag"],
					FlushTimeout: timeout,
				})
			}
			return
		},
		"Collect samples and flush them when the given tag changes its value. Affects the follow-up analysis step, if it is also a batch analysis", builder.RequiredParams("tag"), builder.OptionalParams("timeout"))
}
