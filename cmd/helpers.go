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

func RunPipeline(p *bitflow.SamplePipeline) int {
	if p == nil {
		return 0
	}
	defer golib.ProfileCpu()()
	return p.StartAndWait()
}
