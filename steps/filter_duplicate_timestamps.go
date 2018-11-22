package steps

import (
	"fmt"
	"time"

	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
)

func RegisterDuplicateTimestampFilter(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("filter-duplicate-timestamps",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			var err error
			interval := reg.DurationParam(params, "interval", 0, false, &err)
			if err != nil {
				return err
			}

			var lastTimestamp time.Time
			processor := &pipeline.SimpleProcessor{
				Description: fmt.Sprintf("Drop samples with timestamps closer than %v", interval),
			}
			processor.Process = func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
				if lastTimestamp.IsZero() || sample.Time.Sub(lastTimestamp) > interval {
					lastTimestamp = sample.Time
					return sample, header, nil
				}
				return nil, nil, nil
			}
			p.Add(processor)
			return nil
		}, "Filter samples that follow each other too closely", reg.RequiredParams("interval"))
}
