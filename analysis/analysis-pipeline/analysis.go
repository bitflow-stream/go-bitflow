package main

import (
	"path/filepath"
	"strconv"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/analysis/dbscan"
	"github.com/antongulenko/data2go/sample"
	"github.com/mitchellh/go-homedir"
)

var setSampleSource = &SampleTagger{SourceTags: []string{SourceTag}, DontOverwrite: true}

func init() {
	RegisterAnalysis("", setSampleSource, nil)
	RegisterAnalysis("source", &SampleTagger{SourceTags: []string{SourceTag}, DontOverwrite: false}, nil)
	RegisterAnalysis("no-source", nil, nil)
	RegisterAnalysis("x", setSampleSource, handlePipelineGeneric)
}

// For ad-hoc experiments...
func handlePipelineGeneric(pipe *sample.CmdSamplePipeline) {
	// convertFilenames(&pipe.SamplePipeline)
	// p := &pipe.SamplePipeline

	// p.Add(NewMetricFilter().ExcludeRegex("libvirt|ovsdb")) // .IncludeRegex("cpu|load/|mem/|net-io/|disk-usage///|num_procs"))
	// filterNoiseClusters(p)

	// dbscanRtreeCluster(p)
	// dbscanParallelCluster(p)
	// p.Add(NewMultiHeaderMerger())

	//	p.Add(new(BatchProcessor))
	// .Add(new(SampleShuffler))
	// .Add(new(TimestampSort))
	// .Add(new(MinMaxScaling))
	// .Add(new(StandardizationScaling))
	// .Add(&PCABatchProcessing{ContainedVariance: 0.99})

	// p.Add(new(SamplePrinter))
	// p.Add(new(AbstractProcessor))
	// p.Add(&DecouplingProcessor{ChannelBuffer: 150000})

	// plots(p, false)
}

func plots(p *sample.SamplePipeline, separatePlots bool) {
	home, _ := homedir.Dir()
	p.Add(&Plotter{OutputFile: home + "/clusters/clusters.jpg", ColorTag: ClusterTag, SeparatePlots: separatePlots})
	p.Add(&Plotter{OutputFile: home + "/clusters/classes.jpg", ColorTag: ClassTag, SeparatePlots: separatePlots})
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
	noise := ClusterName(ClusterNoise)
	p.Add(&SampleFilter{IncludeFilter: func(s *sample.Sample) bool {
		return s.Tag(ClusterTag) != noise
	}})
}

func convertFilenames(p *sample.SamplePipeline) {
	if filesource, ok := p.Source.(*sample.FileSource); ok {
		filesource.ConvertFilename = func(filename string) string {
			return filepath.Base(filepath.Dir(filepath.Dir(filename)))
		}
	}
}

type SampleTagger struct {
	SourceTags    []string
	DontOverwrite bool
}

func (h *SampleTagger) HandleHeader(header *sample.Header, source string) {
	header.HasTags = true
}

func (h *SampleTagger) HandleSample(sample *sample.Sample, source string) {
	for _, tag := range h.SourceTags {
		if h.DontOverwrite {
			base := tag
			tag = base
			for i := 0; sample.HasTag(tag); i++ {
				tag = base + strconv.Itoa(i)
			}
		}
		sample.SetTag(tag, source)
	}
}
