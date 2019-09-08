package steps

import (
	"fmt"
	"time"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

// These functions are placed here (and not directly in the bitflow package, next to the BatchProcessor type),
// to avoid an import cycle between the packages bitflow and reg.

// TODO implement DontFlushOnHeaderChange. Requires refactoring of the BatchProcessingStep interface.

// TODO "ignore-header-change"
var BatchProcessorParameters = reg.RegisteredParameters{}.
	Optional("flush-tags", reg.List(reg.String()), []string{}, "Flush the current batch when one or more of the given tags change").
	Optional("flush-no-samples-timeout", reg.Duration(), time.Duration(0)).
	Optional("flush-sample-lag-timeout", reg.Duration(), time.Duration(0)).
	Optional("sample-window-size", reg.Int(), 0).
	Optional("sample-step-size", reg.Int(), 0).
	Optional("time_window-size", reg.Duration(), time.Duration(0)).
	Optional("time-step-size", reg.Duration(), time.Duration(0)).
	Optional("ignore-close", reg.Bool(), false,
		"Do not flush the remaining samples, when the pipeline is closed",
		"The default behavior is to flush on close").
	Optional("forward-immediately", reg.Bool(), false,
		"In addition to the regular batching functionality, output each incoming sample immediately",
		"This will possibly duplicate each incoming sample, since the regular batch processing results are forwarded as well")

func MakeBatchProcessor(params map[string]interface{}) (res *bitflow.BatchProcessor, err error) {
	sampleWindowSize := params["sample-window-size"].(int)
	timeWindowSize := params["time-window-size"].(time.Duration)

	if sampleWindowSize > 0 && timeWindowSize > 0 { // Can only be either sample or time window. Both false is OK
		return nil, fmt.Errorf("arguments 'sample-window-size' and 'time-window-size' are mutually exclusive."+
			" Set either none of them or the one or the other to a positive int value. Current setting: "+
			"'sample-window-size'=%v and 'time-window-size'=%v", sampleWindowSize, timeWindowSize)
	}

	var handler bitflow.WindowHandler
	if timeWindowSize > 0 {
		timeStepSize := params["time-step-size"].(time.Duration)
		if timeStepSize <= 0 {
			timeStepSize = timeWindowSize
		}
		handler = &bitflow.TimeWindowHandler{
			Size:     timeStepSize,
			StepSize: timeStepSize,
		}
	} else if sampleWindowSize > 0 {
		sampleStepSize := params["sample-step-size"].(int)
		if sampleStepSize <= 0 {
			sampleStepSize = sampleWindowSize
		}
		handler = &bitflow.SampleWindowHandler{
			Size:     sampleWindowSize,
			StepSize: sampleStepSize,
		}
	} else {
		handler = &bitflow.SimpleWindowHandler{}
	}

	return &bitflow.BatchProcessor{
		FlushTags:            params["flush-tags"].([]string),
		FlushNoSampleTimeout: params["flush-no-samples-timeout"].(time.Duration),
		FlushSampleLag:       params["flush-sample-lag-timeout"].(time.Duration),
		Handler:              handler,
		DontFlushOnClose:     params["ignore-close"].(bool),
		ForwardImmediately:   params["forward-immediately"].(bool),
	}, nil
	// DontFlushOnHeaderChange: reg.BoolParam(params, "ignore-header-change", false, true, &err),
}
