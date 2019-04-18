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
	defaultPlugin "github.com/bitflow-stream/go-bitflow/steps/bitflow-plugin-default-steps"
	log "github.com/sirupsen/logrus"
)

type CmdPipelineBuilder struct {
	reg.ProcessorRegistry
	SkipInputFlags bool

	printPipeline         bool
	printCapabilities     bool
	printJsonCapabilities bool
	pluginPaths           golib.StringSlice
}

func (c *CmdPipelineBuilder) RegisterFlags() {
	flag.BoolVar(&c.printPipeline, "print-pipeline", false, "Print the parsed pipeline and exit. Can be used to verify the input script.")
	flag.BoolVar(&c.printCapabilities, "capabilities", false, "Print a list of available processing steps and exit.")
	flag.BoolVar(&c.printJsonCapabilities, "json-capabilities", false, "Print the capabilities of this pipeline in JSON form and exit.")
	flag.Var(&c.pluginPaths, "p", "Plugins to load for additional functionality")

	c.ProcessorRegistry = reg.NewProcessorRegistry()
	c.Endpoints.RegisterGeneralFlagsTo(flag.CommandLine)
	c.Endpoints.RegisterOutputFlagsTo(flag.CommandLine)
	if !c.SkipInputFlags {
		c.Endpoints.RegisterInputFlagsTo(flag.CommandLine)
	}
}

func (c *CmdPipelineBuilder) BuildPipeline(getScript func() (string, error)) (*bitflow.SamplePipeline, error) {
	err := load_plugins(c.ProcessorRegistry, c.pluginPaths)
	if err != nil {
		return nil, err
	}
	if c.printJsonCapabilities {
		return nil, c.FormatJsonCapabilities(os.Stdout)
	}
	if c.printCapabilities {
		return nil, c.FormatCapabilities(os.Stdout)
	}

	scriptStr, err := getScript()
	if err != nil {
		return nil, err
	}
	parser := &script.BitflowScriptParser{Registry: c.ProcessorRegistry}
	s, parseErr := parser.ParseScript(scriptStr)
	return s, parseErr.NilOrError()
}

func (c *CmdPipelineBuilder) PrintPipeline(pipe *bitflow.SamplePipeline) *bitflow.SamplePipeline {
	for _, str := range pipe.FormatLines() {
		log.Println(str)
	}
	if c.printPipeline {
		pipe = nil
	}
	return pipe
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
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
