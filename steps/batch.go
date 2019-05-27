package steps

import (
	"time"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

// These functions are placed here (and not directly in the bitflow package, next to the BatchProcessor type),
// to avoid an import cycle between the packages bitflow and reg.

// TODO implement DontFlushOnHeaderChange. Requires refactoring of the BatchProcessingStep interface.

// TODO "ignore-header-change"
var BatchProcessorParameters = reg.RegisteredParameters{}.
	Optional("tags", reg.List(reg.String()), []string{}).
	Optional("timeout", reg.Duration(), time.Duration(0)).
	Optional("ignore-close", reg.Bool(), false).
	Optional("forward-immediately", reg.Bool(), false)

func MakeBatchProcessor(params map[string]interface{}) (res *bitflow.BatchProcessor, err error) {
	return &bitflow.BatchProcessor{
		FlushTags:          params["tags"].([]string),
		FlushTimeout:       params["timeout"].(time.Duration),
		DontFlushOnClose:   params["ignore-close"].(bool),
		ForwardImmediately: params["forward-immediately"].(bool),
	}, nil

	// DontFlushOnHeaderChange: reg.BoolParam(params, "ignore-header-change", false, true, &err),
}
