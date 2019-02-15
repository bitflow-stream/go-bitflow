package cmd

import (
	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
)

func ParseFlags() {
	bitflow.RegisterGolibFlags()
	golib.ParseFlags()
	golib.ConfigureLogging()
}
