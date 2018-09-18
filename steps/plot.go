package steps

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
	"github.com/antongulenko/golib"
	"github.com/lucasb-eyer/go-colorful"
	log "github.com/sirupsen/logrus"
	plotLib "gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	PlotAxisTime = -1
	PlotAxisNum  = -2
	PlotAxisAuto = -3
	minAxis      = PlotAxisAuto

	PlotWidth  = 20 * vg.Centimeter
	PlotHeight = PlotWidth

	numColors      = 100
	plotTimeFormat = "02.01.2006 15:04:05"
	plotTimeLabel  = "time"
	plotNumLabel   = "num"

	ScatterPlot = PlotType(iota)
	LinePlot
	LinePointPlot
	ClusterPlot
	BoxPlot
	InvalidPlotType
)

type PlotType uint

type PlotProcessor struct {
	bitflow.NoopProcessor
	checker bitflow.HeaderChecker

	Type            PlotType
	NoLegend        bool
	AxisX           int
	AxisY           int
	RadiusDimension int
	OutputFile      string
	ColorTag        string
	SeparatePlots   bool // If true, every ColorTag value will create a new plot

	// If not nil, will override the automatially suggested bounds for the respective axis
	ForceXmin *float64
	ForceXmax *float64
	ForceYmin *float64
	ForceYmax *float64

	data         map[string]plotter.XYs
	radiuses     map[string][]float64
	x, y, radius int
	xName, yName string
}

func (p *PlotProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	if p.Type >= InvalidPlotType {
		return golib.NewStoppedChan(fmt.Errorf("Invalid PlotType: %v", p.Type))
	}
	if p.OutputFile == "" {
		return golib.NewStoppedChan(errors.New("Plotter.OutputFile must be configured"))
	}
	if p.AxisX < minAxis || p.AxisY < minAxis {
		return golib.NewStoppedChan(fmt.Errorf("Invalid plot axis values: X=%v Y=%v", p.AxisX, p.AxisY))
	}
	if p.needsRadius() && (p.RadiusDimension < 0 || p.RadiusDimension == p.AxisX || p.RadiusDimension == p.AxisY) {
		return golib.NewStoppedChan(fmt.Errorf("Invalid cluster plot axis values: X=%v Y=%v Radius=%v", p.AxisX, p.AxisY, p.RadiusDimension))
	}
	p.data = make(map[string]plotter.XYs)
	p.radiuses = make(map[string][]float64)

	if file, err := os.Create(p.OutputFile); err != nil {
		// Check if file can be created to quickly fail
		return golib.NewStoppedChan(err)
	} else {
		_ = file.Close() // Drop error
	}
	return p.NoopProcessor.Start(wg)
}

func (p *PlotProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if p.checker.HeaderChanged(header) {
		if err := p.headerChanged(header); err != nil {
			return err
		}
	}
	p.storeSample(sample)
	return p.NoopProcessor.Sample(sample, header)
}

func (p *PlotProcessor) headerChanged(header *bitflow.Header) error {
	p.x = p.AxisX
	p.y = p.AxisY
	p.radius = p.RadiusDimension
	if p.x == PlotAxisAuto {
		if len(header.Fields) > 1 {
			p.x = 0
		} else {
			p.x = PlotAxisTime
		}
	}
	if p.y == PlotAxisAuto {
		if len(header.Fields) > 1 {
			p.y = 1
		} else {
			p.y = 0
		}
	}

	max := p.x
	if p.y > p.x {
		max = p.y
	}
	if len(header.Fields) <= max {
		return fmt.Errorf("%v: Header has %v fields, cannot plot with X=%v and Y=%v", p, len(header.Fields), p.x, p.y)
	}
	if p.needsRadius() && len(header.Fields) <= p.radius {
		return fmt.Errorf("%v: Header has %v fields, cannot plot with Radius=%v", p, len(header.Fields), p.radius)
	}

	var xName, yName string
	if p.x == PlotAxisNum {
		xName = plotNumLabel
	} else if p.x >= 0 {
		xName = header.Fields[p.x]
	} else {
		xName = plotTimeLabel
	}
	if p.y == PlotAxisNum {
		xName = plotNumLabel
	} else if p.y >= 0 {
		yName = header.Fields[p.y]
	} else {
		yName = plotTimeLabel
	}

	if p.xName == "" && p.yName == "" {
		p.xName = xName
		p.yName = yName
	} else if p.xName != xName || p.yName != yName {
		return fmt.Errorf("%v: Header updated and changed the X/Y metric names from %v, %v -> %v, %v", p, p.xName, p.yName, xName, yName)
	}
	return nil
}

func (p *PlotProcessor) needsRadius() bool {
	return p.Type == ClusterPlot
}

func (p *PlotProcessor) storeSample(sample *bitflow.Sample) {
	key := ""
	if p.ColorTag != "" {
		key = sample.Tag(p.ColorTag)
		if key == "" {
			key = "(none)"
		}
	}
	x := p.getVal(p.x, key, sample)
	y := p.getVal(p.y, key, sample)
	p.data[key] = append(p.data[key], struct{ X, Y float64 }{x, y})
	if p.needsRadius() {
		p.radiuses[key] = append(p.radiuses[key], float64(sample.Values[p.radius]))
	}
}

func (p *PlotProcessor) getVal(index int, key string, sample *bitflow.Sample) (res float64) {
	if index == PlotAxisTime {
		res = float64(sample.Time.Unix())
	} else if index == PlotAxisNum {
		res = float64(len(p.data[key]))
	} else if index < len(sample.Values) {
		res = float64(sample.Values[index])
	}
	return
}

func (p *PlotProcessor) Close() {
	if p.Type >= InvalidPlotType || p.OutputFile == "" {
		return
	}

	defer p.CloseSink()
	if p.checker.LastHeader == nil {
		log.Warnf("%s: No data received for plotting", p)
		return
	}
	plot := Plot{
		LabelX:   p.xName,
		LabelY:   p.yName,
		Type:     p.Type,
		NoLegend: p.NoLegend,
	}
	var err error
	if p.SeparatePlots {
		_ = os.Remove(p.OutputFile) // Delete file created in Start(), drop error.
		err = plot.saveSeparatePlots(p.data, p.radiuses, p.OutputFile, p.ForceXmin, p.ForceXmax, p.ForceYmin, p.ForceYmax)
	} else {
		err = plot.savePlot(p.data, p.radiuses, p.OutputFile, p.ForceXmin, p.ForceXmax, p.ForceYmin, p.ForceYmax)
	}
	if err != nil {
		p.Error(err)
	}
}

func (p *PlotProcessor) String() string {
	colorTag := "not colored"
	if p.ColorTag != "" {
		colorTag = "color: " + p.ColorTag
	}
	file := p.OutputFile
	if p.SeparatePlots {
		file = "separate files: " + file
	} else {
		file = "file: " + file
	}
	return fmt.Sprintf("Plotter (%s)(%s)", colorTag, file)
}

// ================================= Plot =================================

type Plot struct {
	LabelX, LabelY string
	Type           PlotType
	NoLegend       bool
}

func (p *Plot) saveSeparatePlots(plotData map[string]plotter.XYs, radiuses map[string][]float64, targetFile string, xMin, xMax, yMin, yMax *float64) error {
	if xMin == nil || xMax == nil || yMin == nil || yMax == nil {
		bounds, err := p.createPlot(plotData, radiuses, xMin, xMax, yMin, yMax)
		if err != nil {
			return err
		}
		xMin = &bounds.X.Min
		xMax = &bounds.X.Max
		yMin = &bounds.Y.Min
		yMax = &bounds.Y.Max
	}

	group := bitflow.NewFileGroup(targetFile)
	for name, data := range plotData {
		plotData := map[string]plotter.XYs{name: data}
		plotFile := group.BuildFilenameStr(name)
		if err := p.savePlot(plotData, radiuses, plotFile, xMin, xMax, yMin, yMax); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plot) savePlot(plotData map[string]plotter.XYs, radiuses map[string][]float64, targetFile string, xMin, xMax, yMin, yMax *float64) error {
	plot, err := p.createPlot(plotData, radiuses, xMin, xMax, yMin, yMax)
	if err != nil {
		return err
	}
	err = plot.Save(PlotWidth, PlotHeight, targetFile)
	if err != nil {
		err = errors.New("Error saving plot: " + err.Error())
	}
	return err
}

func (p *Plot) createPlot(plotData map[string]plotter.XYs, radiuses map[string][]float64, xMin, xMax, yMin, yMax *float64) (*plotLib.Plot, error) {
	plot, err := plotLib.New()
	if err != nil {
		return nil, errors.New("Error creating new plot: " + err.Error())
	}
	if xMin != nil {
		plot.X.Min = *xMin
	}
	if xMax != nil {
		plot.X.Max = *xMax
	}
	if yMin != nil {
		plot.Y.Min = *yMin
	}
	if yMax != nil {
		plot.Y.Max = *yMax
	}
	p.configureAxes(plot)
	return plot, p.fillPlot(plot, plotData, radiuses)
}

func (p *Plot) configureAxes(plt *plotLib.Plot) {
	plt.X.Label.Text = p.LabelX
	plt.Y.Label.Text = p.LabelY
	if p.LabelX == plotTimeLabel {
		plt.X.Tick.Marker = plotLib.TimeTicks{Format: plotTimeFormat}
	}
	if p.LabelY == plotTimeLabel {
		plt.Y.Tick.Marker = plotLib.TimeTicks{Format: plotTimeFormat}
	}
}

func (p *Plot) fillPlot(plot *plotLib.Plot, plotData map[string]plotter.XYs, radiusData map[string][]float64) error {
	shape, err := NewPlotShapeGenerator(numColors)
	if err != nil {
		return err
	}

	if p.Type == BoxPlot {
		return p.fillBoxPlot(plot, plotData)
	}

	for name, data := range plotData {
		plotColor := shape.Colors.Next()
		legend := name != "" && !p.NoLegend

		var scatter *plotter.Scatter
		var line *plotter.Line
		switch p.Type {
		case ScatterPlot:
			scatter, err = plotter.NewScatter(data)
		case LinePlot:
			line, err = plotter.NewLine(data)
		case LinePointPlot:
			line, scatter, err = plotter.NewLinePoints(data)
		case ClusterPlot:
			errorBars := &plotutil.ErrorPoints{
				XYs:     data,
				XErrors: make(plotter.XErrors, len(data)),
				YErrors: make(plotter.YErrors, len(data)),
			}
			radiuses := radiusData[name]
			for i, r := range radiuses {
				errorBars.XErrors[i].Low = r
				errorBars.XErrors[i].High = r
				errorBars.YErrors[i].Low = r
				errorBars.YErrors[i].High = r
			}
			var xErr *plotter.XErrorBars
			var yErr *plotter.YErrorBars
			xErr, err = plotter.NewXErrorBars(errorBars)
			if err == nil {
				yErr, err = plotter.NewYErrorBars(errorBars)
			}
			if err == nil {
				xErr.Color = plotColor
				yErr.Color = plotColor
				plot.Add(xErr, yErr)
				if legend {
					plot.Legend.Add(name, &boxThumbnail{color: plotColor})
					legend = false
				}
			}
		default:
			return fmt.Errorf("Invalid PlotType: %v", p.Type)
		}
		if err != nil {
			return fmt.Errorf("Error creating plot (type %v): %v", p.Type, err)
		}

		if line != nil {
			line.Color = plotColor
			line.Dashes = shape.Dashes.Next()
			plot.Add(line)
			if legend {
				plot.Legend.Add(name, line)
				legend = false
			}
		}
		if scatter != nil {
			scatter.Color = plotColor
			scatter.Shape = shape.Glyphs.Next()
			plot.Add(scatter)
			if legend && line == nil {
				plot.Legend.Add(name, scatter)
				legend = false
			}
		}
	}
	return nil
}

func (p *Plot) fillBoxPlot(plot *plotLib.Plot, plotData map[string]plotter.XYs) error {
	for name, data := range plotData {
		log.Debugf("BoxPlot data key %v, len %v", name, len(data))
	}

	var maxLen int
	var values []plotter.Values
	for i := 0; ; i++ {
		var newValues plotter.Values
		for _, data := range plotData {
			if len(data) > i {
				newValues = append(newValues, data[i].Y) // Ignore X value, since it is implicitly the index
			}
		}
		if len(newValues) == 0 {
			break
		}
		log.Debugf("BoxPlot bucket %v, len %v", i, len(values))
		values = append(values, newValues)
		if maxLen < len(newValues) {
			maxLen = len(newValues)
		}
	}

	minWidth, maxWidth := float64(5), float64(30)
	for i, values := range values {
		width := (float64(len(values))/float64(maxLen))*(maxWidth-minWidth) + minWidth
		box, err := plotter.NewBoxPlot(vg.Length(width), float64(i), values)
		if err != nil {
			return err
		}
		plot.Add(box)
	}
	return nil
}

type boxThumbnail struct {
	color color.Color
}

func (b *boxThumbnail) Thumbnail(c *draw.Canvas) {
	points := []vg.Point{
		{c.Min.X, c.Min.Y},
		{c.Min.X, c.Max.Y},
		{c.Max.X, c.Max.Y},
		{c.Max.X, c.Min.Y},
	}
	poly := c.ClipPolygonY(points)
	c.FillPolygon(b.color, poly)
}

// ================================= Random Colors/Shapes =================================

type PlotShapeGenerator struct {
	Colors *ColorGenerator
	Glyphs *GlyphGenerator
	Dashes *DashesGenerator
}

func NewPlotShapeGenerator(numColors int) (*PlotShapeGenerator, error) {
	colors, err := NewColorGenerator(numColors)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate %v colors: %v", numColors, err)
	}
	return &PlotShapeGenerator{
		Colors: colors,
		Glyphs: NewGlyphGenerator(),
		Dashes: NewDashesGenerator(),
	}, nil
}

type ColorGenerator struct {
	palette []color.Color
	next    int
}

func NewColorGenerator(numColors int) (*ColorGenerator, error) {
	if numColors < 1 {
		numColors = 1
	}
	palette, err := colorful.HappyPalette(numColors)
	if err != nil {
		return nil, err
	}
	colors := make([]color.Color, len(palette))
	for i, c := range palette {
		colors[i] = c
	}
	return &ColorGenerator{
		palette: colors,
	}, nil
}

func (g *ColorGenerator) Next() color.Color {
	if g.next >= len(g.palette) {
		g.next = 0
	}
	result := g.palette[g.next]
	g.next++
	return result
}

type GlyphGenerator struct {
	glyphs []draw.GlyphDrawer
	next   int
}

func NewGlyphGenerator() *GlyphGenerator {
	return &GlyphGenerator{
		glyphs: []draw.GlyphDrawer{
			draw.RingGlyph{},
			draw.SquareGlyph{},
			draw.TriangleGlyph{},
			draw.CrossGlyph{},
			draw.PlusGlyph{},
			//		draw.CircleGlyph{},
			//		draw.BoxGlyph{},
			//		draw.PyramidGlyph{},
		},
	}
}

func (g *GlyphGenerator) Next() draw.GlyphDrawer {
	if g.next >= len(g.glyphs) {
		g.next = 0
	}
	glyph := g.glyphs[g.next]
	g.next++
	return glyph
}

type DashesGenerator struct {
	dashes [][]vg.Length
	next   int
}

func NewDashesGenerator() *DashesGenerator {
	return &DashesGenerator{
		dashes: plotutil.DefaultDashes,
	}
}

func (g *DashesGenerator) Next() []vg.Length {
	if g.next >= len(g.dashes) {
		g.next = 0
	}
	dashes := g.dashes[g.next]
	g.next++
	return dashes
}

func RegisterPlot(b reg.ProcessorRegistry) {
	setPlotBoundParam := func(outErr *error, params map[string]string, paramName string, target **float64) {
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

	create := func(p *pipeline.SamplePipeline, params map[string]string) error {
		plot := &PlotProcessor{
			AxisX:      PlotAxisAuto,
			AxisY:      PlotAxisAuto,
			OutputFile: params["file"],
			Type:       ScatterPlot,
		}
		if colorName, hasColor := params["color"]; hasColor {
			plot.ColorTag = colorName
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
					plot.SeparatePlots = false
					plot.Type = ClusterPlot
					plot.RadiusDimension = 0
					plot.AxisX = 1
					plot.AxisY = 2
				case "box":
					plot.Type = BoxPlot
					plot.AxisX = PlotAxisNum
					plot.AxisY = 0
				case "separate":
					plot.SeparatePlots = true
				case "force_scatter":
					plot.AxisX = 0
					plot.AxisY = 1
				case "force_time":
					plot.AxisX = PlotAxisTime
					plot.AxisY = 0
				default:
					all_flags := []string{"nolegend", "line", "linepoint", "cluster", "separate", "force_scatter", "force_time"}
					return fmt.Errorf("Unkown flag: '%v'. The 'flags' parameter is a comma-separated list of flags: %v", part, all_flags)
				}
			}
		}
		p.Add(plot)
		return nil
	}

	b.RegisterAnalysisParamsErr("plot", create, "Plot a batch of samples to a given filename. The file ending denotes the file type", reg.RequiredParams("file"), reg.OptionalParams("color", "flags", "xMin", "xMax", "yMin", "yMax"))
}
