package main

import (
	"os"

	"github.com/antongulenko/data2go/sample"
)

var pipeline sample.CmdSamplePipeline

func main() {
	os.Exit(pipeline.RunMain())
}
