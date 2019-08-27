package plot

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"sync"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/lucasb-eyer/go-colorful"
	log "github.com/sirupsen/logrus"
	plotLib "gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	AxisTime = -1
	AxisNum  = -2
	AxisAuto = -3
	minAxis  = AxisAuto

	DefaultWidth  = 20 * vg.Centimeter
	DefaultHeight = DefaultWidth

	numColors      = 100
	plotTimeFormat = "02.01.2006 15:04:05"
	plotTimeLabel  = "time"
	plotNumLabel   = "num"

	ScatterPlot = Type(iota)
	LinePlot
	LinePointPlot
	ClusterPlot
	BoxPlot
	InvalidPlotType
)

type Type uint

type Processor struct {
	bitflow.NoopProcessor
	checker bitflow.HeaderChecker

	Type            Type
	NoLegend        bool
	AxisX           int
	AxisY           int
	RadiusDimension int
	OutputFile      string
	ColorTag        string
	SeparatePlots   bool // If true, every ColorTag value will create a new plot

	// If not nil, will override the automatically suggested bounds for the respective axis
	ForceXmin *float64
	ForceXmax *float64
	ForceYmin *float64
	ForceYmax *float64

	data         map[string]plotter.XYs
	radiuses     map[string][]float64
	x, y, radius int
	xName, yName string
}

func (p *Processor) Start(wg *sync.WaitGroup) golib.StopChan {
	if p.Type >= InvalidPlotType {
		return golib.NewStoppedChan(fmt.Errorf("Invalid Type: %v", p.Type))
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

func (p *Processor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if p.checker.HeaderChanged(header) {
		if err := p.headerChanged(header); err != nil {
			return err
		}
	}
	p.storeSample(sample)
	return p.NoopProcessor.Sample(sample, header)
}

func (p *Processor) headerChanged(header *bitflow.Header) error {
	p.x = p.AxisX
	p.y = p.AxisY
	p.radius = p.RadiusDimension
	if p.x == AxisAuto {
		if len(header.Fields) > 1 {
			p.x = 0
		} else {
			p.x = AxisTime
		}
	}
	if p.y == AxisAuto {
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
	if p.x == AxisNum {
		xName = plotNumLabel
	} else if p.x >= 0 {
		xName = header.Fields[p.x]
	} else {
		xName = plotTimeLabel
	}
	if p.y == AxisNum {
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

func (p *Processor) needsRadius() bool {
	return p.Type == ClusterPlot
}

func (p *Processor) storeSample(sample *bitflow.Sample) {
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

func (p *Processor) getVal(index int, key string, sample *bitflow.Sample) (res float64) {
	if index == AxisTime {
		res = float64(sample.Time.Unix())

	} else if index == AxisNum {
		res = float64(len(p.data[key]))
	} else if index < len(sample.Values) {
		res = float64(sample.Values[index])
	}
	return
}

func (p *Processor) Close() {
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
	plotBounds := bounds{
		xMin: p.ForceXmin,
		xMax: p.ForceXmax,
		yMin: p.ForceYmin,
		yMax: p.ForceYmax,
	}
	var err error
	if p.SeparatePlots {
		_ = os.Remove(p.OutputFile) // Delete file created in Start(), drop error.
		err = plot.saveSeparatePlots(p.data, p.radiuses, p.OutputFile, plotBounds)
	} else {
		err = plot.savePlot(p.data, p.radiuses, p.OutputFile, plotBounds)
	}
	if err != nil {
		p.Error(err)
	}
}

func (p *Processor) String() string {
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
	Type           Type
	NoLegend       bool
}

type bounds struct {
	xMin, xMax, yMin, yMax *float64
}

func (b bounds) hasAllBounds() bool {
	return b.xMin != nil && b.xMax != nil && b.yMin != nil && b.yMax != nil
}

func (p *Plot) saveSeparatePlots(plotData map[string]plotter.XYs, radiuses map[string][]float64, targetFile string, bounds bounds) error {
	if !bounds.hasAllBounds() {
		boundsPlot, err := p.createPlot(plotData, radiuses, bounds)
		if err != nil {
			return err
		}
		bounds.xMin = &boundsPlot.X.Min
		bounds.xMax = &boundsPlot.X.Max
		bounds.yMin = &boundsPlot.Y.Min
		bounds.yMax = &boundsPlot.Y.Max
	}

	group := bitflow.NewFileGroup(targetFile)
	for name, data := range plotData {
		subPlotData := map[string]plotter.XYs{name: data}
		plotFile := group.BuildFilenameStr(name)
		if err := p.savePlot(subPlotData, radiuses, plotFile, bounds); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plot) savePlot(plotData map[string]plotter.XYs, radiuses map[string][]float64, targetFile string, bounds bounds) error {
	plot, err := p.createPlot(plotData, radiuses, bounds)
	if err != nil {
		return err
	}
	err = plot.Save(DefaultWidth, DefaultHeight, targetFile)
	if err != nil {
		err = errors.New("Error saving plot: " + err.Error())
	}
	return err
}

func (p *Plot) createPlot(plotData map[string]plotter.XYs, radiuses map[string][]float64, bounds bounds) (*plotLib.Plot, error) {
	plot, err := plotLib.New()
	if err != nil {
		return nil, errors.New("Error creating new plot: " + err.Error())
	}
	if bounds.xMin != nil {
		plot.X.Min = *bounds.xMin
	}
	if bounds.xMax != nil {
		plot.X.Max = *bounds.xMax
	}
	if bounds.yMin != nil {
		plot.Y.Min = *bounds.yMin
	}
	if bounds.yMax != nil {
		plot.Y.Max = *bounds.yMax
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
			return fmt.Errorf("Invalid Type: %v", p.Type)
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
	for i, value := range values {
		width := (float64(len(value))/float64(maxLen))*(maxWidth-minWidth) + minWidth
		box, err := plotter.NewBoxPlot(vg.Length(width), float64(i), value)
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
		{X: c.Min.X, Y: c.Min.Y},
		{X: c.Min.X, Y: c.Max.Y},
		{X: c.Max.X, Y: c.Max.Y},
		{X: c.Max.X, Y: c.Min.Y},
	}
	poly := c.ClipPolygonY(points)
	c.FillPolygon(b.color, poly)
}

// ================================= Random Colors/Shapes =================================

type ShapeGenerator struct {
	Colors *ColorGenerator
	Glyphs *GlyphGenerator
	Dashes *DashesGenerator
}

func NewPlotShapeGenerator(numColors int) (*ShapeGenerator, error) {
	colors, err := NewColorGenerator(numColors)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate %v colors: %v", numColors, err)
	}
	return &ShapeGenerator{
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
			// 		draw.CircleGlyph{},
			// 		draw.BoxGlyph{},
			// 		draw.PyramidGlyph{},
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
	create := func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
		xMin := params["xMin"].(float64)
		xMax := params["xMax"].(float64)
		yMin := params["yMin"].(float64)
		yMax := params["yMax"].(float64)

		plot := &Processor{
			AxisX:      AxisAuto,
			AxisY:      AxisAuto,
			OutputFile: params["file"].(string),
			ColorTag:   params["color"].(string),
			Type:       ScatterPlot,
			ForceXmin:  &xMin,
			ForceXmax:  &xMax,
			ForceYmin:  &yMin,
			ForceYmax:  &yMax,
		}

		plot.NoLegend = params["nolegend"].(bool)
		plot.SeparatePlots = params["separate"].(bool)

		is_line := params["line"].(bool)
		is_linepoint := params["linepoint"].(bool)
		is_cluster := params["cluster"].(bool)
		is_box := params["box"].(bool)
		// Check if exactly one of the following flags is true
		areAnyTrue := is_line || is_linepoint || is_cluster || is_box
		areTwoTrue := (is_line && areAnyTrue) || (is_linepoint && areAnyTrue) || (is_cluster && areAnyTrue) ||
			(is_box && areAnyTrue)
		if !(areAnyTrue && !areTwoTrue) {
			return fmt.Errorf("Parameters: Only one of the follwoing parameters can be true, but is " +
				"line=%v, linepoint=%v, clusetr=%v, box=%v.", is_line, is_linepoint, is_cluster, is_box)
		}
		if is_line {
			plot.Type = LinePlot
		}
		if is_linepoint {
			plot.Type = LinePointPlot
		}
		if is_cluster {
			plot.SeparatePlots = false
			plot.Type = ClusterPlot
			plot.RadiusDimension = 0
			plot.AxisX = 1
			plot.AxisY = 2
		}
		if is_box {
			plot.Type = BoxPlot
			plot.AxisX = AxisNum
			plot.AxisY = 0
		}

		force_scatter := params["force_scatter"].(bool)
		force_time := params["force_time"].(bool)
		if force_scatter && force_time {
			return fmt.Errorf("Parameters: Only one of the follwoing parameters can be true, but is "+
				"force_scatter=%v, force_time=%v.", force_scatter, force_time)
		}
		if force_scatter {
			plot.AxisX = 0
			plot.AxisY = 1
		}
		if force_time {
			plot.AxisX = AxisTime
			plot.AxisY = 0
			//  Fix to make time axis autoscale. If it is 0.0, the time axis starts at 1970...
			if *plot.ForceXmin == 0.0 {
				plot.ForceXmin = nil
			}
			if *plot.ForceXmax == 0.0 {
				plot.ForceXmin = nil
			}
		}
		p.Add(plot)
		return nil
	}

	b.RegisterStep("plot", create,
		"Plot a batch of samples to a given filename. The file ending denotes the file type").
		Required("file", reg.String()).
		Optional("color", reg.String(), "").
		Optional("flags", reg.List(reg.String()), []string{}).
		Optional("nolegend", reg.Bool(), false).
		Optional("line", reg.Bool(), false).
		Optional("linepoint", reg.Bool(), false).
		Optional("cluster", reg.Bool(), false).
		Optional("box", reg.Bool(), false).
		Optional("separate", reg.Bool(), false).
		Optional("force_scatter", reg.Bool(), true).
		Optional("force_time", reg.Bool(), true).
		Optional("xMin", reg.Float(), 0.0).
		Optional("xMax", reg.Float(), 0.0).
		Optional("yMin", reg.Float(), 0.0).
		Optional("yMax", reg.Float(), 0.0)
}
