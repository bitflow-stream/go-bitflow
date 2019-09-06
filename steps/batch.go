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
	Optional("window-size-samples", reg.Int(), 0).
	Optional("step-size-samples", reg.Int(), 0).
	Optional("window-size-time", reg.Duration(), time.Duration(0)).
	Optional("step-size-time", reg.Duration(), time.Duration(0)).
	Optional("ignore-close", reg.Bool(), false, "Do not flush the remaining samples, when the pipeline is closed", "The default behavior is to flush on close").
	Optional("forward-immediately", reg.Bool(), false, "In addition to the regular batching functionality, output each incoming sample immediately", "This will possibly duplicate each incoming sample, since the regular batch processing results are forwarded as well")

func MakeBatchProcessor(params map[string]interface{}) (res *bitflow.BatchProcessor, err error) {
	windowSizeSamples := params["window-size-samples"].(int)
	windowSizeTime := params["window-size-time"].(time.Duration)

	isSampleWindow := windowSizeSamples > 0
	isTimeWindow := windowSizeTime > 0
	if isSampleWindow && isTimeWindow { // Can only be either sample or time window. Both false is OK
		return nil, fmt.Errorf("Arguments 'window-size-samples' and 'window-size-samples' are mutually exclusive." +
			" Set either none of them or the one or the other to a positive int value.")
	}

	var handler bitflow.WindowHandler
	if isSampleWindow {
		stepSizeSamples := params["step-size-samples"].(int)
		if stepSizeSamples <= 0 {
			stepSizeSamples = windowSizeSamples
		}
		handler = &bitflow.SampleWindowHandler{
			Size:     windowSizeSamples,
			StepSize: stepSizeSamples,
		}
	} else if isTimeWindow {
		stepSizeTime := params["step-size-time"].(time.Duration)
		if stepSizeTime <= 0 {
			stepSizeTime = windowSizeTime
		}
		handler = &bitflow.TimeWindowHandler{
			Size:     windowSizeTime,
			StepSize: stepSizeTime,
		}
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
