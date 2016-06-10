package main

import (
	"flag"
	"os"
	"strconv"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/analysis/dbscan"
	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
	"github.com/mitchellh/go-homedir"
)

func do_main() int {
	var p sample.CmdSamplePipeline
	p.ParseFlags()
	flag.Parse()
	defer golib.ProfileCpu()()
	p.Init()
	registerProcessingSteps(&p.SamplePipeline)
	return p.StartAndWait()
}

func main() {
	os.Exit(do_main())
}

func registerProcessingSteps(p *sample.SamplePipeline) {
	// p.Add(NewMetricFilter().ExcludeRegex("libvirt|ovsdb")) // .IncludeRegex("cpu|load/|mem/|net-io/|disk-usage///|num_procs"))
	// filterNoiseClusters(p)

	// dbscanRtreeCluster(p)
	// dbscanParallelCluster(p)

	// p.Add(new(BatchProcessor))
	// .Add(new(TimestampSort))
	// .Add(new(MinMaxScaling))
	// .Add(new(StandardizationScaling))
	// .Add(&PCABatchProcessing{ContainedVariance: 0.99})

	// p.Add(new(SamplePrinter))
	// p.Add(new(AbstractProcessor))
	// p.Add(&DecouplingProcessor{ChannelBuffer: 150000})

	plots(p, false)
}

func plots(p *sample.SamplePipeline, separatePlots bool) {
	home, _ := homedir.Dir()
	p.Add(&Plotter{OutputFile: home + "/clusters/clusters.jpg", ColorTag: "cluster", SeparatePlots: separatePlots})
	p.Add(&Plotter{OutputFile: home + "/clusters/classes.jpg", ColorTag: "cls", SeparatePlots: separatePlots})
}

func dbscanRtreeCluster(p *sample.SamplePipeline) {
	p.Add(new(BatchProcessor).Add(new(MinMaxScaling)).Add(
		&dbscan.DbscanBatchClusterer{
			Dbscan:          dbscan.Dbscan{Eps: 0.1, MinPts: 5},
			TreeMinChildren: 25,
			TreeMaxChildren: 50,
			TreePointWidth:  0.0001,
		}))
}

func dbscanParallelCluster(p *sample.SamplePipeline) {
	p.Add(new(BatchProcessor).Add(new(MinMaxScaling)).Add(&dbscan.ParallelDbscanBatchClusterer{Eps: 0.3, MinPts: 5}))
}

func filterNoiseClusters(p *sample.SamplePipeline) {
	noise := dbscan.ClusterPrefix + strconv.Itoa(dbscan.ClusterNoise)
	p.Add(&SampleFilter{IncludeFilter: func(s *sample.Sample) bool {
		return s.Tags[dbscan.ClusterTag] != noise
	}})
}
