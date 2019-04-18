package steps

import (
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

// These functions are placed here (and not directly in the bitflow package, next to the BatchProcessor type),
// to avoid an import cycle between the packages bitflow and reg.

func MakeBatchProcessorParameters() reg.RegisteredParameters {
	return reg.RegisteredParameters{
		Optional: []string{"tag", "timeout"},
	}
}

func MakeBatchProcessor(params map[string]string) (res *bitflow.BatchProcessor, err error) {
	res = new(bitflow.BatchProcessor)
	if tag, ok := params["tag"]; ok {
		res.FlushTags = []string{tag}
	}
	res.FlushTimeout = reg.DurationParam(params, "timeout", 0, true, &err)
	return
}
