package main

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/http"
)

func init() {
	RegisterAnalysisParams("plot", plot, "[<color tag>,]<output filename>")
	RegisterAnalysisParams("plot_separate", separate_plots, "same as plot")
	RegisterAnalysisParams("stats", feature_stats, "output filename for metric statistics")
	RegisterAnalysisParams("http", print_http, "HTTP endpoint to listen for requests")
}

func plot(pipe *SamplePipeline, params string) {
	do_plot(pipe, params, false)
}

func separate_plots(pipe *SamplePipeline, params string) {
	do_plot(pipe, params, true)
}

func do_plot(pipe *SamplePipeline, params string, separatePlots bool) {
	if params == "" {
		log.Fatalln("-e plot needs parameters (-e plot,[<tag>,]<filename>)")
	}
	index := strings.IndexRune(params, ',')
	tag := ""
	filename := params
	if index == -1 {
		log.Warnln("-e plot got no tag parameter, not coloring plot (-e plot,[<tag>,]<filename>)")
	} else {
		tag = params[:index]
		filename = params[index+1:]
	}
	pipe.Add(&Plotter{OutputFile: filename, ColorTag: tag, SeparatePlots: separatePlots})
}

func feature_stats(pipe *SamplePipeline, params string) {
	if params == "" {
		log.Fatalln("-e stats needs parameter: file to store feature statistics")
	} else {
		pipe.Add(NewStoreStats(params))
	}
}

func print_http(p *SamplePipeline, params string) {
	parts := strings.Split(params, ",")
	endpoint := parts[0]
	windowSize := 100
	if len(parts) >= 2 {
		var err error
		windowSize, err = strconv.Atoi(parts[1])
		if err != nil {
			log.Fatalln("Failed to parse second parmeter for -e http (must be integer):", err)
		}
	}
	p.Add(plotHttp.NewHttpPlotter(endpoint, windowSize))
}
