package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/plugin"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/script"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/script_go"
	defaultPlugin "github.com/bitflow-stream/go-bitflow-pipeline/steps/bitflow-plugin-default-steps"
	log "github.com/sirupsen/logrus"
)

const (
	fileFlag            = "f"
	BitflowScriptSuffix = ".bf"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <flags> <bitflow script>\nAll flags must be defined before the first non-flag parameter.\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	fix_arguments(&os.Args)
	os.Exit(do_main())
}

func fix_arguments(argsPtr *[]string) {
	if n := golib.ParseHashbangArgs(argsPtr); n > 0 {
		// Insert -f before the script file, if necessary
		args := *argsPtr
		if args[n-1] != "-"+fileFlag && args[n-1] != "--"+fileFlag {
			args = append(args, "") // Extend by one entry
			copy(args[n+1:], args[n:])
			args[n] = "-" + fileFlag
			*argsPtr = args
		}
	}
}

func do_main() int {
	printAnalyses := flag.Bool("print-analyses", false, "Print a list of available analyses and exit.")
	printPipeline := flag.Bool("print-pipeline", false, "Print the parsed pipeline and exit. Can be used to verify the input script.")
	printCapabilities := flag.Bool("capabilities", false, "Print the capabilities of this pipeline in JSON form and exit.")
	useNewScript := flag.Bool("new", false, "Use the new script parser for processing the input script.")
	scriptFile := ""
	var pluginPaths golib.StringSlice
	flag.Var(&pluginPaths, "p", "Plugins to load for additional functionality")
	flag.StringVar(&scriptFile, fileFlag, "", "File to read a Bitflow script from (alternative to providing the script on the command line)")
	registry := reg.NewProcessorRegistry()
	bitflow.RegisterGolibFlags()
	registry.Endpoints.RegisterFlags()
	flag.Parse()
	golib.ConfigureLogging()
	golib.Checkerr(load_plugins(registry, pluginPaths))

	if *printCapabilities {
		registry.PrintJsonCapabilities(os.Stdout)
		return 0
	}
	if *printAnalyses {
		fmt.Printf("Available analysis steps:\n%v\n", registry.PrintAllAnalyses())
		return 0
	}

	rawScript, err := get_script(flag.Args(), scriptFile)
	golib.Checkerr(err)
	make_pipeline := make_pipeline_old
	if *useNewScript {
		log.Println("Running using new ANTLR script implementation")
		make_pipeline = make_pipeline_new
	}
	pipe, err := make_pipeline(registry, rawScript)
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

func get_script(parsedArgs []string, scriptFile string) (string, error) {
	if scriptFile != "" && len(parsedArgs) > 0 {
		return "", errors.New("Please provide a bitflow pipeline script either via -f or as parameter, not both.")
	}
	if len(parsedArgs) == 1 && strings.HasSuffix(parsedArgs[0], BitflowScriptSuffix) {
		// Special case when passing a single existing .bf file as positional argument: Treat it as a script file
		info, err := os.Stat(parsedArgs[0])
		if err == nil && info.Mode().IsRegular() {
			scriptFile = parsedArgs[0]
		}
	}
	var rawScript string
	if scriptFile != "" {
		scriptBytes, err := ioutil.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("Error reading bitflow script file %v: %v", scriptFile, err)
		}
		rawScript = string(scriptBytes)
	} else {
		rawScript = strings.TrimSpace(strings.Join(parsedArgs, " "))
	}
	if rawScript == "" {
		return "", errors.New("Please provide a bitflow pipeline script via -f or directly as parameter.")
	}
	return rawScript, nil
}

func make_pipeline_old(registry reg.ProcessorRegistry, scriptStr string) (*pipeline.SamplePipeline, error) {
	queryBuilder := script_go.PipelineBuilder{registry}
	parser := script_go.NewParser(bytes.NewReader([]byte(scriptStr)))
	pipe, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return queryBuilder.MakePipeline(pipe)
}

func make_pipeline_new(registry reg.ProcessorRegistry, scriptStr string) (*pipeline.SamplePipeline, error) {
	s, err := script.NewAntlrBitflowParser(registry).ParseScript(scriptStr)
	return s, err.NilOrError()
}

func load_plugins(registry reg.ProcessorRegistry, pluginPaths []string) error {
	loadedNames := make(map[string]bool)
	for _, path := range pluginPaths {
		if name, err := plugin.LoadPlugin(registry, path); err != nil {
			return fmt.Errorf("Failed to load plugin %v: %v", path, err)
		} else {
			loadedNames[name] = true
		}
	}

	// Load the default pipeline steps
	// TODO add a plugin discovery mechanism
	return defaultPlugin.Plugin.Init(registry)
}
