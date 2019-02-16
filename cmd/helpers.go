package cmd

import (
	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
)

func ParseFlags() []string {
	bitflow.RegisterGolibFlags()
	args := golib.ParseFlags()
	golib.ConfigureLogging()
	return args
}
