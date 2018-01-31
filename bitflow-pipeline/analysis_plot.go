package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/http"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

func RegisterPlots(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("plot", plot, "Plot a batch of samples to a given filename. The file ending denotes the file type", []string{"file"}, "color", "flags", "xMin", "xMax", "yMin", "yMax")
	b.RegisterAnalysisParams("stats", feature_stats, "Output statistics about processed samples to a given ini-file", []string{"file"})
	b.RegisterAnalysisParamsErr("http", print_http, "Serve HTTP-based plots about processed metrics values to the given HTTP endpoint", []string{"endpoint"}, "window", "local_static")
}

func plot(p *SamplePipeline, params map[string]string) error {
	plot := &PlotProcessor{
		AxisX:      PlotAxisAuto,
		AxisY:      PlotAxisAuto,
		OutputFile: params["file"],
		Type:       ScatterPlot,
	}
	if color, hasColor := params["color"]; hasColor {
		plot.ColorTag = color
	}
	var err error
	setPlotBoundParam(&err, params, "xMin", &plot.ForceXmin)
	setPlotBoundParam(&err, params, "xMax", &plot.ForceXmax)
	setPlotBoundParam(&err, params, "yMin", &plot.ForceYmin)
	setPlotBoundParam(&err, params, "yMax", &plot.ForceYmax)
	if err != nil {
		return err
	}

	if flagsStr, hasFlags := params["flags"]; hasFlags {
		flags := strings.Split(flagsStr, ",")
		for _, part := range flags {
			switch part {
			case "nolegend":
				plot.NoLegend = true
			case "line":
				plot.Type = LinePlot
			case "linepoint":
				plot.Type = LinePointPlot
			case "separate":
				plot.SeparatePlots = true
			case "force_scatter":
				plot.AxisX = 0
				plot.AxisY = 1
			case "force_time":
				plot.AxisX = PlotAxisTime
				plot.AxisY = 0
			default:
				all_flags := []string{"nolegend", "line", "linepoint", "separate", "force_scatter", "force_time"}
				return fmt.Errorf("Unkown flag: '%v'. The 'flags' parameter is a comma-separated list of flags: %v", part, all_flags)
			}
		}
	}
	p.Add(plot)
	return nil
}

func setPlotBoundParam(outErr *error, params map[string]string, paramName string, target **float64) {
	param, hasParam := params[paramName]
	if *outErr == nil && hasParam {
		val, err := strconv.ParseFloat(param, 64)
		if err != nil {
			*outErr = fmt.Errorf("Failed to parse argument of '%s': %v", paramName, err)
			return
		}
		*target = &val
	}
}

func feature_stats(p *SamplePipeline, params map[string]string) {
	p.Add(NewStoreStats(params["file"]))
}

func print_http(p *SamplePipeline, params map[string]string) error {
	windowSize := 100
	if windowStr, ok := params["window"]; ok {
		var err error
		windowSize, err = strconv.Atoi(windowStr)
		if err != nil {
			return query.ParameterError("window", err)
		}
	}
	useLocalStatic := false
	static, ok := params["local_static"]
	if ok {
		if static == "true" {
			useLocalStatic = true
		} else {
			return query.ParameterError("local_static", errors.New("The only accepted value is 'true'"))
		}
	}
	p.Add(plotHttp.NewHttpPlotter(params["endpoint"], windowSize, useLocalStatic))
	return nil
}
