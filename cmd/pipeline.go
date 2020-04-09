package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/plugin"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/script/script"
	"github.com/bitflow-stream/go-bitflow/steps"
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
	externalCommands      golib.StringSlice
}

func (c *CmdPipelineBuilder) RegisterFlags() {
	flag.BoolVar(&c.printPipeline, "print-pipeline", false, "Print the parsed pipeline and exit. Can be used to verify the input script.")
	flag.BoolVar(&c.printCapabilities, "capabilities", false, "Print a list of available processing steps and exit.")
	flag.BoolVar(&c.printJsonCapabilities, "json-capabilities", false, "Print the capabilities of this pipeline in JSON form and exit.")
	flag.Var(&c.pluginPaths, "p", "Plugins to load for additional functionality")
	flag.Var(&c.externalCommands, "exe", "Register external executable to be used as step. Format: '<short name>;<executable path>;<initial executable parameters>'")

	c.ProcessorRegistry = reg.NewProcessorRegistry(bitflow.NewEndpointFactory())
	c.Endpoints.RegisterGeneralFlagsTo(flag.CommandLine)
	c.Endpoints.RegisterOutputFlagsTo(flag.CommandLine)
	if !c.SkipInputFlags {
		c.Endpoints.RegisterInputFlagsTo(flag.CommandLine)
	}
}

func (c *CmdPipelineBuilder) BuildPipeline(getScript func() (string, error)) (*bitflow.SamplePipeline, error) {
	err := loadPlugins(c.ProcessorRegistry, c.pluginPaths)
	if err != nil {
		return nil, err
	}
	err = registerExternalExecutables(c.ProcessorRegistry, c.externalCommands)
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

// Print the pipeline and return true, if the program should continue by executing it.
// If false is returned, the program should exit after printing.
func (c *CmdPipelineBuilder) PrintPipeline(pipe *bitflow.SamplePipeline) bool {
	for _, str := range pipe.FormatLines() {
		log.Println(str)
	}
	return !c.printPipeline
}

func loadPlugins(registry reg.ProcessorRegistry, pluginPaths []string) error {
	for _, path := range pluginPaths {
		if _, err := plugin.LoadPlugin(registry, path); err != nil {
			return fmt.Errorf("Failed to load plugin %v: %v", path, err)
		}
	}

	// Load the default pipeline steps
	// TODO add a plugin discovery mechanism
	return defaultPlugin.Plugin.Init(registry)
}

func registerExternalExecutables(registry reg.ProcessorRegistry, commands golib.StringSlice) error {
	for _, description := range commands {
		if err := steps.RegisterExecutable(registry, description); err != nil {
			return err
		}
	}
	return nil
}
