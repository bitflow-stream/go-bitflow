package cmd

import (
	"flag"
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

// TODO this helper type is awkwardly placed, find better package.

type CmdDataCollector struct {
	Endpoints     *bitflow.EndpointFactory
	DefaultOutput string
	RestApis      []RestApiPath // HttpTagger and FileOutputFilterApi included by default

	restApiEndpoint string
	fileOutputApi   FileOutputFilterApi
	outputs         golib.StringSlice
	flagTags        golib.KeyValueStringSlice
}

func (c *CmdDataCollector) ParseFlags() {
	flag.Var(&c.outputs, "o", "Data sink(s) for outputting data")
	flag.Var(&c.flagTags, "tag", "All collected samples will have the given tags (key=value) attached.")
	// flag.Var(&splitMetrics, "split", "Provide a regex. Metrics that are matched will be converted to separate samples. When the regex contains named groups, their names and values will be added as tags, and an individual sample will be created for each unique value combination.")
	flag.BoolVar(&c.fileOutputApi.FileOutputEnabled, "default-enable-file-output", false, "Enables file output immediately. By default it must be enable through the REST API first.")
	flag.StringVar(&c.restApiEndpoint, "api", "", "Enable REST API for controlling the collector. "+
		"The API can be used to control tags and enable/disable file output.")

	// Parse command line flags
	if c.Endpoints == nil {
		c.Endpoints = bitflow.NewEndpointFactory()
	}
	c.Endpoints.RegisterGeneralFlagsTo(flag.CommandLine)
	c.Endpoints.RegisterOutputFlagsTo(flag.CommandLine)
	bitflow.RegisterGolibFlags()
	flag.Parse()
	golib.ConfigureLogging()
}

func (c *CmdDataCollector) MakePipeline() *bitflow.SamplePipeline {
	// Configure the data collector pipeline
	p := new(bitflow.SamplePipeline)
	if len(c.flagTags.Keys) > 0 {
		p.Add(steps.NewTaggingProcessor(c.flagTags.Map()))
	}
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
	c.add_outputs(p)
	return p
}

func (c *CmdDataCollector) add_outputs(p *bitflow.SamplePipeline) {
	outputs := c.create_outputs()
	if len(outputs) == 1 {
		c.set_sink(p, outputs[0])
	} else {
		// Create a multiplex-fork for all outputs
		dist := new(fork.MultiplexDistributor)
		for _, sink := range outputs {
			pipe := new(bitflow.SamplePipeline)
			c.set_sink(pipe, sink)
			dist.Subpipelines = append(dist.Subpipelines, pipe)
		}
		p.Add(&fork.SampleFork{Distributor: dist})
	}
}

func (c *CmdDataCollector) create_outputs() []bitflow.SampleProcessor {
	if len(c.outputs) == 0 && c.DefaultOutput != "" {
		c.outputs = []string{c.DefaultOutput}
	}
	var sinks []bitflow.SampleProcessor
	consoleOutputs := 0
	for _, output := range c.outputs {
		sink, err := c.Endpoints.CreateOutput(output)
		sinks = append(sinks, sink)
		golib.Checkerr(err)
		if bitflow.IsConsoleOutput(sink) {
			consoleOutputs++
		}
		if consoleOutputs > 1 {
			golib.Fatalln("Cannot define multiple outputs to stdout")
		}
	}
	return sinks
}

func (c *CmdDataCollector) set_sink(p *bitflow.SamplePipeline, sink bitflow.SampleProcessor) {
	// Add a filter to file outputs
	if _, isFile := sink.(*bitflow.FileSink); isFile {
		if c.restApiEndpoint != "" {
			p.Add(&steps.SampleFilter{
				Description: bitflow.String("Filter samples based on /file_output REST API."),
				IncludeFilter: func(sample *bitflow.Sample, header *bitflow.Header) (bool, error) {
					return c.fileOutputApi.FileOutputEnabled, nil
				},
			})
		}
	}
	p.Add(sink)
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
