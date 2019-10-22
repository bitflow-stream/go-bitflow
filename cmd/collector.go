package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/bitflow/fork"
	"github.com/bitflow-stream/go-bitflow/steps"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	RestApiPathPrefix = "/api"
)

type CmdDataCollector struct {
	CmdPipelineBuilder
	DefaultOutput string
	RestApis      []RestApiPath // HttpTagger and FileOutputFilterApi included by default

	restApiEndpoint string
	fileOutputApi   FileOutputFilterApi
	script          string
	scriptFile      string
	outputs         golib.StringSlice
}

func (c *CmdDataCollector) RegisterFlags() {
	c.SkipInputFlags = true
	c.CmdPipelineBuilder.RegisterFlags()

	flag.StringVar(&c.script, "s", "", "Provide a Bitflow Script snippet, that will be executed before outputting the produced samples. The script must not contain an input.")
	flag.StringVar(&c.scriptFile, "f", "", "Like -s, but provide a script file instead.")
	flag.Var(&c.outputs, "o", "Data sink(s) for outputting data. Will be appended at the end of provided Bitflow script(s), if any.")
	flag.BoolVar(&c.fileOutputApi.FileOutputEnabled, "default-enable-file-output", false, "Enables file output immediately. By default it must be enable through the REST API first.")
	flag.StringVar(&c.restApiEndpoint, "api", "", "Enable REST API for controlling the collector. "+
		"The API can be used to control tags and enable/disable file output.")
}

func (c *CmdDataCollector) BuildPipeline() (*bitflow.SamplePipeline, error) {
	p, err := c.CmdPipelineBuilder.BuildPipeline(c.getScript)
	if err != nil || p == nil {
		return p, err
	}

	// TODO move this extra functionality to reusable pipeline steps. Probably define some default script snippet to load.
	if c.restApiEndpoint != "" {
		router := mux.NewRouter()
		tagger := steps.NewHttpTagger(RestApiPathPrefix, router)
		c.fileOutputApi.Register(RestApiPathPrefix, router)
		for _, api := range c.RestApis {
			api.Register(RestApiPathPrefix, router)
		}
		server := http.Server{
			Addr:    c.restApiEndpoint,
			Handler: router,
		}
		// Do not add this routine to any wait group, as it cannot be stopped
		go func() {
			tagger.Error(server.ListenAndServe())
		}()
		p.Add(tagger)
	}
	if err := c.addOutputs(p); err != nil {
		return nil, err
	}
	p = c.CmdPipelineBuilder.PrintPipeline(p)
	return p, nil
}

func (c *CmdDataCollector) getScript() (string, error) {
	if c.script != "" && c.scriptFile != "" {
		return "", fmt.Errorf("Cannot specify both an immediate Bitflow script through -s and a Bitflow script file through -f")
	}
	script := "empty://-" // Prepend an empty input to ensure the script compiles
	extraScript := c.script
	if c.scriptFile != "" {
		scriptBytes, err := ioutil.ReadFile(c.scriptFile)
		if err != nil {
			return "", fmt.Errorf("Error reading Bitflow script file %v: %v", c.scriptFile, err)
		}
		extraScript = string(scriptBytes)
	}
	if extraScript != "" {
		script += " -> " + extraScript
	}
	return script, nil
}

func (c *CmdDataCollector) addOutputs(p *bitflow.SamplePipeline) error {
	outputs, err := c.createOutputs()
	if err != nil {
		return err
	}
	if len(outputs) == 1 {
		c.setSink(p, outputs[0])
	} else {
		// Create a multiplex-fork for all outputs
		dist := new(fork.MultiplexDistributor)
		for _, sink := range outputs {
			pipe := new(bitflow.SamplePipeline)
			c.setSink(pipe, sink)
			dist.Subpipelines = append(dist.Subpipelines, pipe)
		}
		p.Add(&fork.SampleFork{Distributor: dist})
	}
	return nil
}

func (c *CmdDataCollector) createOutputs() ([]bitflow.SampleProcessor, error) {
	if len(c.outputs) == 0 && c.DefaultOutput != "" {
		c.outputs = []string{c.DefaultOutput}
	}
	var sinks []bitflow.SampleProcessor
	consoleOutputs := 0
	for _, output := range c.outputs {
		sink, err := c.Endpoints.CreateOutput(output)
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, sink)
		if bitflow.IsConsoleOutput(sink) {
			consoleOutputs++
		}
		if consoleOutputs > 1 {
			return nil, fmt.Errorf("Cannot define multiple outputs to stdout")
		}
	}
	return sinks, nil
}

func (c *CmdDataCollector) setSink(p *bitflow.SamplePipeline, sink bitflow.SampleProcessor) {
	// Add a filter to file outputs
	if c.isFileSink(sink) && c.restApiEndpoint != "" {
		p.Add(&steps.SampleFilter{
			Description: bitflow.String("Filter samples based on /file_output REST API."),
			IncludeFilter: func(sample *bitflow.Sample, header *bitflow.Header) (bool, error) {
				return c.fileOutputApi.FileOutputEnabled, nil
			},
		})
	}
	p.Add(sink)
}

func (c *CmdDataCollector) isFileSink(sink bitflow.SampleProcessor) (res bool) {
	// A file sink is either a simple FileSink object, or a fork with a MultiFileDistributor
	_, res = sink.(*bitflow.FileSink)
	if !res {
		if sampleFork, isFork := sink.(*fork.SampleFork); isFork {
			_, res = sampleFork.Distributor.(*fork.MultiFileDistributor)
		}
	}
	return
}

type RestApiPath interface {
	Register(pathPrefix string, router *mux.Router)
}

type FileOutputFilterApi struct {
	lock              sync.Mutex
	FileOutputEnabled bool
}

func (api *FileOutputFilterApi) Register(pathPrefix string, router *mux.Router) {
	router.HandleFunc(pathPrefix+"/file_output", api.handleRequest).Methods("GET", "POST", "PUT", "DELETE")
}

func (api *FileOutputFilterApi) handleRequest(w http.ResponseWriter, r *http.Request) {
	api.lock.Lock()
	oldStatus := api.FileOutputEnabled
	newStatus := oldStatus
	switch r.Method {
	case "GET":
	case "POST", "PUT":
		newStatus = true
	case "DELETE":
		newStatus = false
	}
	api.FileOutputEnabled = newStatus
	api.lock.Unlock()

	var status string
	if api.FileOutputEnabled {
		status = "enabled"
	} else {
		status = "disabled"
	}
	status = "File output is " + status
	if oldStatus != newStatus {
		log.Println(status)
	}
	w.Write([]byte(status + "\n"))
}
