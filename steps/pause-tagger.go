package steps

import (
	"fmt"
	"strconv"
	"time"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
)

func RegisterPauseTagger(b reg.ProcessorRegistry) {
	create := func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		tag := params["tag"]
		duration, err := time.ParseDuration(params["minPause"])
		if err != nil {
			return reg.ParameterError("minPause", err)
		}
		pipeline.Add(&PauseTagger{MinimumPause: duration, Tag: tag})
		return nil
	}

	b.RegisterAnalysisParamsErr("tag-pauses", create, "Set a given tag to an integer value, that increments whenever the timestamps of two samples are more apart than a given duration", reg.RequiredParams("tag", "minPause"))
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
