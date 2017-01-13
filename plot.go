package pipeline

import (
	"errors"
	"fmt"
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
)

const (
	PlotAxisTime = -1
	PlotAxisAuto = -2
	minAxis      = PlotAxisAuto

	PlotWidth  = 20 * vg.Centimeter
	PlotHeight = PlotWidth
)

func init() {
	plotutil.DefaultColors = append(plotutil.DefaultColors, plotutil.DarkColors...)
	plotutil.DefaultGlyphShapes = []draw.GlyphDrawer{
		draw.RingGlyph{},
		draw.SquareGlyph{},
		draw.TriangleGlyph{},
		draw.CrossGlyph{},
		draw.PlusGlyph{},
	}
}

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
}

func (p *Plotter) Start(wg *sync.WaitGroup) golib.StopChan {
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
	if p.checker.LastHeader == nil {
		if p.AxisX == PlotAxisAuto {
			if len(header.Fields) > 1 {
				p.AxisX = 0
			} else {
				p.AxisX = PlotAxisTime
			}
		}
		if p.AxisY == PlotAxisAuto {
			if len(header.Fields) > 1 {
				p.AxisY = 1
			} else {
				p.AxisY = 0
			}
		}

		log.Println("X:", p.AxisX, "Y:", p.AxisY)

		max := p.AxisX
		if p.AxisY > p.AxisX {
			max = p.AxisY
		}
		if len(header.Fields) <= max {
			return fmt.Errorf("%v: Header has %v fields, cannot plot with X=%v and Y=%v", p, len(header.Fields), p.AxisX, p.AxisY)
		}
	}
	if p.checker.InitializedHeaderChanged(header) {
		return fmt.Errorf("%v: Cannot handle changed header", p)
	}
	p.storeSample(sample)
	return p.OutgoingSink.Sample(sample, header)
}

func (p *Plotter) storeSample(sample *bitflow.Sample) {
	key := sample.Tag(p.ColorTag)
	if key == "" && p.ColorTag != "" {
		key = "(none)"
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

func (p *Plotter) fillPlot(plot *plot.Plot, header *bitflow.Header, plotSamples map[string][]*bitflow.Sample) error {
	if p.AxisX < 0 {
		plot.X.Label.Text = "time"
	} else {
		plot.X.Label.Text = header.Fields[p.AxisX]
	}
	if p.AxisY < 0 {
		plot.Y.Label.Text = "time"
	} else {
		plot.Y.Label.Text = header.Fields[p.AxisY]
	}

	plotData := make(map[string]plotter.XYer)
	for name, data := range plotSamples {
		plotData[name] = plotDataContainer{
			values: data,
			x:      p.AxisX,
			y:      p.AxisY,
		}
	}
	return p.Filler(plot, plotData)
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

func FillScatterPlot(plot *plot.Plot, data map[string]plotter.XYer) error {
	var parameters []interface{}
	for name, data := range data {
		parameters = append(parameters, name, data)
	}
	if err := plotutil.AddScatters(plot, parameters...); err != nil {
		return fmt.Errorf("Error creating scatter plot: %v", err)
	}
	return nil
}

func FillLinePlot(plot *plot.Plot, plotData map[string]plotter.XYer) error {
	var parameters []interface{}
	for name, data := range plotData {
		parameters = append(parameters, name, data)
	}
	if err := plotutil.AddLines(plot, parameters...); err != nil {
		return fmt.Errorf("Error creating line plot: %v", err)
	}
	return nil
}
