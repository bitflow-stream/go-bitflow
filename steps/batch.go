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
	Optional("flush-num-samples", reg.Int(), 0).
	Optional("flush-time-diff", reg.Duration(), time.Duration(0)).
	Optional("ignore-close", reg.Bool(), false, "Do not flush the remaining samples, when the pipeline is closed", "The default behavior is to flush on close").
	Optional("forward-immediately", reg.Bool(), false, "In addition to the regular batching functionality, output each incoming sample immediately", "This will possibly duplicate each incoming sample, since the regular batch processing results are forwarded as well")

func MakeBatchProcessor(params map[string]interface{}) (res *bitflow.BatchProcessor, err error) {

	if params["flush-time-diff"].(time.Duration) != 0 && params["flush-num-samples"].(int) != 0 {
		return nil, fmt.Errorf("Arguments 'flush-time-diff' and 'flush-num-samples' are mutually exclusive." +
			" Set either the one or the other.")
	}
	return &bitflow.BatchProcessor{
		FlushTags:            params["flush-tags"].([]string),
		FlushNoSampleTimeout: params["flush-no-samples-timeout"].(time.Duration),
		FlushSampleLag:       params["flush-sample-lag-timeout"].(time.Duration),
		FlushAfterNumSamples: params["flush-num-samples"].(int),
		FlushAfterTime:       params["flush-time-diff"].(time.Duration),
		DontFlushOnClose:     params["ignore-close"].(bool),
		ForwardImmediately:   params["forward-immediately"].(bool),
	}, nil
	// DontFlushOnHeaderChange: reg.BoolParam(params, "ignore-header-change", false, true, &err),
}
