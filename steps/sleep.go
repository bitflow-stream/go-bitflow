package steps

import (
	"fmt"
	"time"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterSleep(b reg.ProcessorRegistry) {
	b.RegisterStep("sleep", _create_sleep_processor,
		"Between every two samples, sleep the time difference between their timestamps").
		Optional("time", reg.Duration(), 0).
		Optional("onChangedTag", reg.String(), "")
}

func _create_sleep_processor(p *bitflow.SamplePipeline, params map[string]interface{}) error {
	timeout := params["time"].(time.Duration)
	changedTag := params["onChangedTag"].(string)

	desc := "sleep between samples"
	if timeout > 0 {
		desc += fmt.Sprintf(" (%v)", timeout)
	} else {
		desc += " (timestamp difference)"
	}
	if changedTag != "" {
		desc += " when tag " + changedTag + " changes"
	}

	previousTag := ""
	var lastTimestamp time.Time
	processor := &bitflow.SimpleProcessor{
		Description: desc,
	}
	processor.Process = func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
		doSleep := true
		if changedTag != "" {
			newTag := sample.Tag(changedTag)
			if newTag == previousTag {
				doSleep = false
			}
			previousTag = newTag
		}
		if doSleep {
			if timeout > 0 {
				processor.StopChan.WaitTimeout(timeout)
			} else {
				last := lastTimestamp
				if !last.IsZero() {
					diff := sample.Time.Sub(last)
					if diff > 0 {
						processor.StopChan.WaitTimeout(diff)
					}
				}
				lastTimestamp = sample.Time
			}
		}
		return sample, header, nil
	}
	p.Add(processor)
	return nil
}
