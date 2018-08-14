package steps

import (
	"fmt"
	"time"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

func RegisterDuplicateTimestampFilter(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("filter-duplicate-timestamps",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			var err error
			interval := query.DurationParam(params, "interval", 0, false, &err)
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
		}, "Filter samples that follow each other too closely", []string{"interval"})
}
