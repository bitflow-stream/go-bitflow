package cmd

import (
	"flag"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
)

func ParseFlags() {
	bitflow.RegisterGolibFlags()
	flag.Parse()
	golib.ConfigureLogging()
}
