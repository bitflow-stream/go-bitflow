package main

import (
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline/plugin"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Fatalln("This package is intended to be loaded as a plugin, not executed directly")
}

var source = RandomSampleGenerator{
	Interval:   300 * time.Millisecond,
	ErrorAfter: 2000,
	CloseAfter: 1000,
	TimeOffset: 0,
	ExtraTags: map[string]string{
		"plugin": "mock",
	},
	Header: []string{
		"num", "x", "y", "z",
	},
}

// The Symbol to be loaded
var Plugin = plugin.SampleSourcePluginImplementation{
	Create: func(params map[string]string) (bitflow.SampleSource, error) {
		generator := source
		err := source.ParseParams(params)
		return &generator, err
	},
}
