package main

import (
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/http"
)

func init() {
	RegisterAnalysisParams("plot", plot, "[(<color-tag>|line|separate|force_time|force_scatter),]*<output filename>")
	RegisterAnalysisParams("stats", feature_stats, "output filename for metric statistics")
	RegisterAnalysisParams("http", print_http, "HTTP endpoint to listen for requests")
}

func plot(pipe *SamplePipeline, params string) {
	if params == "" {
		log.Fatalln("-e plot needs parameters")
	}

	axisX := PlotAxisAuto
	axisY := PlotAxisAuto
	colorTag := ""
	separatePlots := false
	filler := FillScatterPlot
	parts := strings.Split(params, ",")
	filename := parts[len(parts)-1]
	if len(parts) > 0 {
		for _, part := range parts[:len(parts)-1] {
			switch part {
			case "line":
				filler = FillLinePlot
			case "separate":
				separatePlots = true
			case "force_scatter":
				axisX = 0
				axisY = 1
			case "force_time":
				axisX = PlotAxisTime
				axisY = 0
			default:
				if colorTag != "" {
					log.Fatalln("Multiple color-tag parameters given for plot")
				}
				colorTag = part
			}
		}
	}

	if colorTag == "" {
		log.Warnln("Plot got no color-tag parameter, not coloring plot")
	}
	pipe.Add(&Plotter{
		AxisX:         axisX,
		AxisY:         axisY,
		Filler:        filler,
		OutputFile:    filename,
		ColorTag:      colorTag,
		SeparatePlots: separatePlots,
	})
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
