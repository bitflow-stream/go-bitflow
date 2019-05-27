package steps

import (
	"fmt"
	"time"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterDuplicateTimestampFilter(b reg.ProcessorRegistry) {
	b.RegisterStep("filter-duplicate-timestamps",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			var err error
			interval := params["interval"].(time.Duration)
			if err != nil {
				return err
			}

			var lastTimestamp time.Time
			processor := &bitflow.SimpleProcessor{
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
		}, "Filter samples that follow each other too closely").
		Required("interval", reg.Duration())
}
