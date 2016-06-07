package analysis

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
)

const (
	PlotWidth    = 20 * vg.Centimeter
	PlotHeight   = PlotWidth
	PlottedXAxis = 0
	PlottedYAxis = 1
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

type Plotter struct {
	AbstractProcessor
	OutputFile string
	ColorTag   string
	plot       *plot.Plot
	data       map[string]PlotData
}

type PlotData []sample.Sample

func (data PlotData) Len() int {
	return len(data)
}

func (data PlotData) XY(i int) (x, y float64) {
	values := data[i].Values
	return float64(values[PlottedXAxis]), float64(values[PlottedYAxis])
}

func (p *Plotter) Header(header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	} else if len(header.Fields) < 2 {
		return fmt.Errorf("Cannot plot header with %v fields, need at least 2", len(header.Fields))
	} else {
		if err := p.configurePlot(header); err != nil {
			return err
		}
		return p.OutgoingSink.Header(header)
	}
}

func (p *Plotter) configurePlot(header sample.Header) (err error) {
	p.data = make(map[string]PlotData)
	if p.plot, err = plot.New(); err != nil {
		return
	}
	p.plot.X.Label.Text = header.Fields[PlottedXAxis]
	p.plot.Y.Label.Text = header.Fields[PlottedYAxis]
	return
}

func (p *Plotter) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	p.plotSample(sample)
	return p.OutgoingSink.Sample(sample, header)
}

func (p *Plotter) plotSample(sample sample.Sample) {
	key, ok := sample.Tags[p.ColorTag]
	if !ok {
		key = "(none)"
	}
	p.data[key] = append(p.data[key], sample)
}

func (p *Plotter) Start(wg *sync.WaitGroup) golib.StopChan {
	if file, err := os.Create(p.OutputFile); err != nil {
		return golib.TaskFinishedError(err)
	} else {
		_ = file.Close() // Drop error
	}
	return nil
}

func (p *Plotter) Close() {
	p.savePlot()
	p.CloseSink(nil)
}

func (p *Plotter) savePlot() {
	if p.plot == nil {
		return
	}
	var parameters []interface{}
	for name, data := range p.data {
		parameters = append(parameters, name, data)
	}
	if err := plotutil.AddScatters(p.plot, parameters...); err != nil {
		log.Println("Error creating plot:", err)
		return
	}
	if err := p.plot.Save(PlotWidth, PlotHeight, p.OutputFile); err != nil {
		log.Println("Error saving plot:", err)
	}
}

func (p *Plotter) String() string {
	return "Plotter"
}
