package steps

import (
	"fmt"
	"strconv"
	"strings"

	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

func REGISTER_PLOT(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("plot", plot, "Plot a batch of samples to a given filename. The file ending denotes the file type", []string{"file"}, "color", "flags", "xMin", "xMax", "yMin", "yMax")
	b.RegisterAnalysisParams("stats", feature_stats, "Output statistics about processed samples to a given ini-file", []string{"file"})
}

func plot(p *pipeline.SamplePipeline, params map[string]string) error {
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
			case "cluster":
				plot.Type = ClusterPlot
				plot.RadiusDimension = 0
				plot.AxisX = 1
				plot.AxisY = 2
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

func feature_stats(p *pipeline.SamplePipeline, params map[string]string) {
	p.Add(NewStoreStats(params["file"]))
}
