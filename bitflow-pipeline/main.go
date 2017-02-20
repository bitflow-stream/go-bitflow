package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <flags> <bitflow script>\nAll flags must be defined before the first non-flag parameter.\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	os.Exit(do_main())
}

func do_main() int {
	printAnalyses := flag.Bool("print-analyses", false, "Print a list of available analyses and exit.")
	printPipeline := flag.Bool("print-pipeline", false, "Print the parsed pipeline and exit. Can be used to verify the input script.")

	bitflow.RegisterGolibFlags()
	builder.Endpoints.RegisterFlags()
	flag.Parse()
	golib.ConfigureLogging()
	if *printAnalyses {
		fmt.Printf("Available analysis steps:\n%v\n", builder.PrintAllAnalyses())
		return 0
	}

	pipeline, err := make_pipeline()
	if err != nil {
		log.Errorln(err)
		log.Fatalln("Use -print-analyses to print all available analysis steps.")
	}
	defer golib.ProfileCpu()()
	for _, str := range pipeline.FormatLines() {
		log.Println(str)
	}
	if *printPipeline {
		return 0
	}
	return pipeline.StartAndWait()
}

func make_pipeline() (*pipeline.SamplePipeline, error) {
	script := strings.Join(flag.Args(), " ")
	if strings.TrimSpace(script) == "" {
		return nil, errors.New("Please provide a bitflow pipeline script")
	}
	parser := query.NewParser(bytes.NewReader([]byte(script)))
	pipe, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return builder.MakePipeline(pipe)
}
