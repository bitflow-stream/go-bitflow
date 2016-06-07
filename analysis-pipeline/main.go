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

var home, _ = homedir.Dir()

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
	//	filterNoiseClusters(p)

	//	dbscanRtreeCluster(p)
	//	dbscanParallelCluster(p)

	//	p.Add(new(BatchProcessor))
	// .Add(new(TimestampSort))
	// .Add(new(MinMaxScaling))
	// .Add(new(StandardizationScaling))
	// .Add(&PCABatchProcessing{ContainedVariance: 0.99})
	// .Add(&DbscanBatchClusterer{Eps: 0.1, MinPts: 5})

	// p.Add(new(SamplePrinter))
	// p.Add(new(AbstractProcessor))
	// p.Add(&DecouplingProcessor{ChannelBuffer: 150000})

	// p.Add(&Plotter{OutputFile: home + "/clusters.jpg", ColorTag: "cluster"})
	// p.Add(&Plotter{OutputFile: home + "/classes.jpg", ColorTag: "cls"})
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
	p.Add(new(BatchProcessor).Add(new(MinMaxScaling)).Add(&DbscanBatchClusterer{Eps: 0.1, MinPts: 5}))
}

func filterNoiseClusters(p *sample.SamplePipeline) {
	noise := dbscan.ClusterPrefix + strconv.Itoa(dbscan.ClusterNoise)
	p.Add(&SampleFilter{IncludeFilter: func(s *sample.Sample) bool {
		return s.Tags[dbscan.ClusterTag] != noise
	}})
}
