package cmd

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/plugin"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/script/script"
	"github.com/bitflow-stream/go-bitflow/script/script_go"
	defaultPlugin "github.com/bitflow-stream/go-bitflow/steps/bitflow-plugin-default-steps"
	log "github.com/sirupsen/logrus"
)

type CmdPipelineBuilder struct {
	registry reg.ProcessorRegistry

	printAnalyses                bool
	printPipeline                bool
	printPipelineSchedulingHints bool
	printCapabilities            bool
	useNewScript                 bool
	pluginPaths                  golib.StringSlice
}

func (c *CmdPipelineBuilder) RegisterFlags() {
	flag.BoolVar(&c.printAnalyses, "print-analyses", false, "Print a list of available analyses and exit.")
	flag.BoolVar(&c.printPipeline, "print-pipeline", false, "Print the parsed pipeline and exit. Can be used to verify the input script.")
	flag.BoolVar(&c.printPipelineSchedulingHints, "print-scheduling-hints", false, "Print the subpipelines with scheduling hints as json and exit.")
	flag.BoolVar(&c.printCapabilities, "capabilities", false, "Print the capabilities of this pipeline in JSON form and exit.")
	flag.BoolVar(&c.useNewScript, "new", false, "Use the new script parser for processing the input script.")
	flag.Var(&c.pluginPaths, "p", "Plugins to loadfor additional functionality")

	c.registry = reg.NewProcessorRegistry()
	c.registry.Endpoints.RegisterFlags()
}

func (c *CmdPipelineBuilder) BuildPipeline(script string) (*bitflow.SamplePipeline, error) {
	err := load_plugins(c.registry, c.pluginPaths)
	if c.printCapabilities {
		return nil, c.registry.PrintJsonCapabilities(os.Stdout)
	}
	if c.printAnalyses {
		fmt.Printf("Available analysis steps:\n%v\n", c.registry.PrintAllAnalyses())
		return nil, nil
	}

	if c.printPipelineSchedulingHints {
		return nil, convertAndPrintSchedulingHints(script)
	}
	make_pipeline := make_pipeline_old
	if c.useNewScript {
		log.Println("Running using new ANTLR script implementation")
		make_pipeline = make_pipeline_new
	}
	pipe, err := make_pipeline(c.registry, script)
	if err != nil {
		return nil, err
	}

	for _, str := range pipe.FormatLines() {
		log.Println(str)
	}
	if c.printPipeline {
		pipe = nil
	}
	return pipe, nil
}

func convertAndPrintSchedulingHints(rawScript string) error {
	scripts, errs := new(script.BitflowScriptScheduleParser).ParseScript(rawScript)
	if err := errs.NilOrError(); err != nil {
		return err
	}
	j, err := JSONMarshal(scripts)
	if err != nil {
		return err
	}
	fmt.Println(string(j))
	return nil
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func make_pipeline_old(registry reg.ProcessorRegistry, scriptStr string) (*bitflow.SamplePipeline, error) {
	queryBuilder := script_go.PipelineBuilder{registry}
	parser := script_go.NewParser(bytes.NewReader([]byte(scriptStr)))
	pipe, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return queryBuilder.MakePipeline(pipe)
}

func make_pipeline_new(registry reg.ProcessorRegistry, scriptStr string) (*bitflow.SamplePipeline, error) {
	s, err := (&script.BitflowScriptParser{Registry: registry}).ParseScript(scriptStr)
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
