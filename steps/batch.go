package steps

import (
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
)

func RegisterGenericBatch(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("batch",
		func(p *pipeline.SamplePipeline, params map[string]string) (err error) {
			timeout := reg.DurationParam(params, "timeout", 0, true, &err)
			if err == nil {
				p.Add(&pipeline.BatchProcessor{
					FlushTags:    []string{params["tag"]},
					FlushTimeout: timeout,
				})
			}
			return
		},
		"Collect samples and flush them on different events (wall time/sample time/tag change/number of samples). Affects the follow-up analysis step, if it is also a batch analysis", reg.RequiredParams("tag"), reg.OptionalParams("timeout"))
}
