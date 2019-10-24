package main

import (
	"errors"
	"os"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/cmd"
	log "github.com/sirupsen/logrus"
)

var NoPositionalArgumentsExpected = errors.New("This program does not expect any positional non-flag command line arguments")

func main() {
	os.Exit(doMain())
}

func doMain() int {
	collector := NewDataCollector()
	collector.RegisterFlags()

	helper := cmd.CmdDataCollector{DefaultOutput: "csv://-"}
	helper.RegisterFlags()
	_, args := cmd.ParseFlags()
	defer golib.ProfileCpu()()
	if err := collector.Initialize(args); err != nil {
		log.Errorln("Initialization failed:", err)
		return 1
	}

	pipe, err := helper.BuildPipeline(collector)
	if err != nil {
		log.Errorln("Failed to build Bitflow pipeline:", err)
		return 2
	}
	return pipe.StartAndWait()
}
