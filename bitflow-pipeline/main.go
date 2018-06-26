package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/http"
	"github.com/antongulenko/go-bitflow-pipeline/http_tags"
	"github.com/antongulenko/go-bitflow-pipeline/plugin"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/go-bitflow-pipeline/steps"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

var builder = query.NewPipelineBuilder()

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
	printCapabilities := flag.Bool("capabilities", false, "Print the capablities of this pipeline in JSON form and exit.")
	scriptFile := ""
	flag.StringVar(&scriptFile, "f", "", "File to read a Bitflow script from (alternative to providing the script on the command line)")

	plugin.RegisterPluginDataSource(&builder.Endpoints)
	register_analyses(builder)

	bitflow.RegisterGolibFlags()
	builder.Endpoints.RegisterFlags()
	flag.Parse()
	golib.ConfigureLogging()
	if *printCapabilities {
		builder.PrintJsonCapabilities(os.Stdout)
		return 0
	}
	if *printAnalyses {
		fmt.Printf("Available analysis steps:\n%v\n", builder.PrintAllAnalyses())
		return 0
	}

	script := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if scriptFile != "" && script != "" {
		golib.Fatalln("Please provide a bitflow pipeline script either via -f or as parameter, not both.")
	}
	if scriptFile != "" {
		scriptBytes, err := ioutil.ReadFile(scriptFile)
		if err != nil {
			golib.Fatalf("Error reading bitflow script file $v: %v", scriptFile, err)
		}
		script = string(scriptBytes)
	}
	if script == "" {
		golib.Fatalln("Please provide a bitflow pipeline script via -f or directly as parameter.")
	}

	pipe, err := make_pipeline(script)
	if err != nil {
		log.Errorln(err)
		golib.Fatalln("Use -print-analyses to print all available analysis steps.")
	}
	defer golib.ProfileCpu()()
	for _, str := range pipe.FormatLines() {
		log.Println(str)
	}
	if *printPipeline {
		return 0
	}
	return pipe.StartAndWait()
}

func make_pipeline(script string) (*pipeline.SamplePipeline, error) {
	parser := query.NewParser(bytes.NewReader([]byte(script)))
	pipe, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return builder.MakePipeline(pipe)
}

func register_analyses(b *query.PipelineBuilder) {
	RegisterTaggingAnalyses(b)

	steps.REGISTER_BASIC(b)
	steps.REGISTER_MATH(b)
	steps.REGISTER_PLOT(b)
	steps.REGISTER_PRINT(b)

	plotHttp.RegisterHttpPlotter(b)
	steps.RegisterDecouple(b)
	steps.RegisterGenericBatch(b)
	steps.RegisterSleep(b)
	steps.RegisterNoop(b)
	steps.RegisterTaggingProcessor(b)
	steps.RegisterPipelineRateSynchronizer(b)
	steps.RegisterSampleMerger(b)
	steps.RegisterForks(b)
	http_tags.RegisterHttpTagger(b)
}
