package steps

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterPauseTagger(b reg.ProcessorRegistry) {
	create := func(pipeline *bitflow.SamplePipeline, params map[string]interface{}) error {
		pipeline.Add(&PauseTagger{
			MinimumPause: params["minPause"].(time.Duration),
			Tag:          params["tag"].(string),
		})
		return nil
	}
	b.RegisterStep("tag-pauses", create,
		"Set a given tag to an integer value, that increments whenever the timestamps of two samples are more apart than a given duration").
		Required("tag", reg.String()).
		Required("minPause", reg.Duration())
}

type PauseTagger struct {
	bitflow.NoopProcessor
	MinimumPause time.Duration
	Tag          string

	counter  int
	lastTime time.Time
}

func (d *PauseTagger) String() string {
	return fmt.Sprintf("increment tag '%v' after pauses of %v", d.Tag, d.MinimumPause.String())
}

func (d *PauseTagger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	last := d.lastTime
	d.lastTime = sample.Time
	if !last.IsZero() && sample.Time.Sub(last) >= d.MinimumPause {
		d.counter++
	}
	sample.SetTag(d.Tag, strconv.Itoa(d.counter))
	return d.GetSink().Sample(sample, header)
}
