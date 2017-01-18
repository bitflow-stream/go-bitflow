package pipeline

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	PlotAxisTime = -1
	PlotAxisAuto = -2
	minAxis      = PlotAxisAuto

	PlotWidth  = 20 * vg.Centimeter
	PlotHeight = PlotWidth

	numColors      = 100
	plotTimeFormat = "02.01.2006 15:04:05"
)

type PlotFiller func(plot *plot.Plot, data map[string]plotter.XYer) error

type Plotter struct {
	bitflow.AbstractProcessor
	checker bitflow.HeaderChecker
	data    map[string][]*bitflow.Sample

	Filler        PlotFiller
	AxisX         int
	AxisY         int
	OutputFile    string
	ColorTag      string
	SeparatePlots bool // If true, every ColorTag value will create a new plot
	NoLegend      bool

	x      int
	y      int
	colors *ColorGenerator
	glyphs *GlyphGenerator
	dashes *DashesGenerator
}

func (p *Plotter) Start(wg *sync.WaitGroup) golib.StopChan {
	var err error
	p.dashes = NewDashesGenerator()
	p.glyphs = NewGlyphGenerator()
	p.colors, err = NewColorGenerator(numColors)
	if err != nil {
		return golib.TaskFinishedError(fmt.Errorf("Failed to generate %v colors: %v", numColors, err))
	}

	p.data = make(map[string][]*bitflow.Sample)
	if p.Filler == nil {
		return golib.TaskFinishedError(errors.New("Plotter.Filler field must not be nil"))
	}
	if p.OutputFile == "" {
		return golib.TaskFinishedError(errors.New("Plotter.OutputFile must be configured"))
	}
	if p.AxisX < minAxis || p.AxisY < minAxis {
		return golib.TaskFinishedError(fmt.Errorf("Invalid plot axis values: X=%v Y=%v", p.AxisX, p.AxisY))
	}

	if file, err := os.Create(p.OutputFile); err != nil {
		// Check if file can be created to quickly fail
		return golib.TaskFinishedError(err)
	} else {
		_ = file.Close() // Drop error
	}
	return p.AbstractProcessor.Start(wg)
}

func (p *Plotter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	if p.checker.HeaderChanged(header) {
		if err := p.headerChanged(header); err != nil {
			return err
		}
	}
	p.storeSample(sample)
	return p.OutgoingSink.Sample(sample, header)
}

func (p *Plotter) headerChanged(header *bitflow.Header) error {
	p.x = p.AxisX
	p.y = p.AxisY
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
	return nil
}

func (p *Plotter) storeSample(sample *bitflow.Sample) {
	key := ""
	if p.ColorTag != "" {
		key = sample.Tag(p.ColorTag)
		if key == "" {
			key = "(none)"
		}
	}
	p.data[key] = append(p.data[key], sample)
}

func (p *Plotter) Close() {
	if p.Filler == nil || p.OutputFile == "" {
		return
	}

	defer p.CloseSink()
	if p.checker.LastHeader == nil {
		log.Warnf("%s: No data received for plotting", p)
		return
	}
	var err error
	if p.SeparatePlots {
		_ = os.Remove(p.OutputFile) // Delete file created in Start(), drop error.
		err = p.saveSeparatePlots()
	} else {
		err = p.savePlot(p.data, nil, p.OutputFile)
	}
	if err != nil {
		p.Error(err)
	}
}

func (p *Plotter) saveSeparatePlots() error {
	bounds, err := p.createPlot(p.data, nil)
	if err != nil {
		return err
	}
	group := bitflow.NewFileGroup(p.OutputFile)
	for name, data := range p.data {
		plotData := map[string][]*bitflow.Sample{name: data}
		plotFile := group.BuildFilenameStr(name)
		if err := p.savePlot(plotData, bounds, plotFile); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plotter) savePlot(plotData map[string][]*bitflow.Sample, copyBounds *plot.Plot, targetFile string) error {
	plot, err := p.createPlot(plotData, copyBounds)
	if err != nil {
		return err
	}
	err = plot.Save(PlotWidth, PlotHeight, targetFile)
	if err != nil {
		err = errors.New("Error saving plot: " + err.Error())
	}
	return err
}

func (p *Plotter) createPlot(plotData map[string][]*bitflow.Sample, copyBounds *plot.Plot) (*plot.Plot, error) {
	plot, err := plot.New()
	if err != nil {
		return nil, errors.New("Error creating new plot: " + err.Error())
	}
	if copyBounds != nil {
		plot.X.Min = copyBounds.X.Min
		plot.X.Max = copyBounds.X.Max
		plot.Y.Min = copyBounds.Y.Min
		plot.Y.Max = copyBounds.Y.Max
	}
	p.fillPlot(plot, p.checker.LastHeader, plotData)
	return plot, nil
}

func (self *Plotter) fillPlot(p *plot.Plot, header *bitflow.Header, plotSamples map[string][]*bitflow.Sample) error {
	if self.x < 0 {
		p.X.Label.Text = "time"
		p.X.Tick.Marker = plot.TimeTicks{Format: plotTimeFormat}
	} else {
		p.X.Label.Text = header.Fields[self.x]
	}
	if self.y < 0 {
		p.Y.Label.Text = "time"
		p.Y.Tick.Marker = plot.TimeTicks{Format: plotTimeFormat}
	} else {
		p.Y.Label.Text = header.Fields[self.y]
	}

	plotData := make(map[string]plotter.XYer)
	for name, data := range plotSamples {
		plotData[name] = plotDataContainer{
			values: data,
			x:      self.x,
			y:      self.y,
		}
	}
	return self.Filler(p, plotData)
}

func (p *Plotter) String() string {
	return fmt.Sprintf("Plotter (color: %s)(file: %s)", p.ColorTag, p.OutputFile)
}

type plotDataContainer struct {
	values []*bitflow.Sample
	x      int
	y      int
}

func (data plotDataContainer) Len() int {
	return len(data.values)
}

func (data plotDataContainer) XY(i int) (x, y float64) {
	if data.x < 0 {
		x = float64(data.values[i].Time.Unix())
	} else {
		x = float64(data.values[i].Values[data.x])
	}
	if data.y < 0 {
		y = float64(data.values[i].Time.Unix())
	} else {
		y = float64(data.values[i].Values[data.y])
	}
	return
}

func (p *Plotter) FillScatterPlot(plot *plot.Plot, plotData map[string]plotter.XYer) error {
	for name, data := range plotData {
		scatter, err := plotter.NewScatter(data)
		if err != nil {
			return fmt.Errorf("Error creating scatter plot: %v", err)
		}
		scatter.Color = p.colors.Next()
		scatter.Shape = p.glyphs.Next()
		plot.Add(scatter)
		if name != "" && !p.NoLegend {
			plot.Legend.Add(name, scatter)
		}
	}
	return nil
}

func (p *Plotter) FillLinePointPlot(plot *plot.Plot, plotData map[string]plotter.XYer) error {
	for name, data := range plotData {
		lines, scatter, err := plotter.NewLinePoints(data)
		if err != nil {
			return fmt.Errorf("Error creating line point plot: %v", err)
		}
		color := p.colors.Next()
		lines.Color = color
		scatter.Color = color
		lines.Dashes = p.dashes.Next()
		plot.Add(lines, scatter)
		if name != "" && !p.NoLegend {
			plot.Legend.Add(name, lines)
		}
	}
	return nil
}

func (p *Plotter) FillLinePlot(plot *plot.Plot, plotData map[string]plotter.XYer) error {
	for name, data := range plotData {
		lines, err := plotter.NewLine(data)
		if err != nil {
			return fmt.Errorf("Error creating line plot: %v", err)
		}
		color := p.colors.Next()
		lines.Color = color
		lines.Dashes = p.dashes.Next()
		plot.Add(lines)
		if name != "" && !p.NoLegend {
			plot.Legend.Add(name, lines)
		}
	}
	return nil
}

// ================================= Random Colors/Shapes =================================

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
	color := g.palette[g.next]
	g.next++
	return color
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
