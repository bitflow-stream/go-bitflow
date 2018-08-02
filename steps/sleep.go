package steps

import (
	"fmt"
	"time"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
)

func RegisterSleep(b builder.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("sleep", _create_sleep_processor, "Between every two samples, sleep the time difference between their timestamps", builder.OptionalParams("time", "onChangedTag"))
}

func _create_sleep_processor(p *pipeline.SamplePipeline, params map[string]string) error {
	var timeout time.Duration
	timeoutStr, hasTimeout := params["time"]
	changedTag, hasOnTagChange := params["onChangedTag"]
	if hasTimeout {
		var err error
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			return builder.ParameterError("time", err)
		}
	}

	desc := "sleep between samples"
	if hasTimeout {
		desc += fmt.Sprintf(" (%v)", timeout)
	} else {
		desc += " (timestamp difference)"
	}
	if hasOnTagChange {
		desc += " when tag " + changedTag + " changes"
	}

	previousTag := ""
	var lastTimestamp time.Time
	processor := &pipeline.SimpleProcessor{
		Description: desc,
	}
	processor.Process = func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
		doSleep := true
		if hasOnTagChange {
			newTag := sample.Tag(changedTag)
			if newTag == previousTag {
				doSleep = false
			}
			previousTag = newTag
		}
		if doSleep {
			if hasTimeout {
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
