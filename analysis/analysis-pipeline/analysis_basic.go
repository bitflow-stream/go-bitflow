package main

import (
	"flag"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

var (
	metric_filter_include golib.StringSlice
	metric_filter_exclude golib.StringSlice
)

func init() {
	RegisterSampleHandler("src", &SampleTagger{SourceTags: []string{SourceTag}})
	RegisterSampleHandler("src-append", &SampleTagger{SourceTags: []string{SourceTag}, DontOverwrite: true})

	RegisterAnalysis("decouple", decouple_samples)
	RegisterAnalysis("merge_headers", merge_headers)
	RegisterAnalysis("pick", pick_x_percent)   // param: samples to keep 0..1
	RegisterAnalysis("filter_tag", filter_tag) // param: tag=value or tag!=value

	RegisterAnalysis("shuffle", shuffle_data)
	RegisterAnalysis("sort", sort_data) // Param: comma-separated list of tags

	RegisterAnalysis("min_max", normalize_min_max)
	RegisterAnalysis("standardize", normalize_standardize)

	RegisterAnalysis("plot", plot)
	RegisterAnalysis("separate_plots", separate_plots)

	RegisterAnalysis("filter_metrics", filter_metrics)
	flag.Var(&metric_filter_include, "metrics_include", "Include regex used with '-e filter_metrics'")
	flag.Var(&metric_filter_exclude, "metrics_exclude", "Exclude regex used with '-e filter_metrics'")
}

func shuffle_data(p *SamplePipeline, _ string) {
	p.Batch(new(SampleShuffler))
}

func sort_data(p *SamplePipeline, params string) {
	tags := strings.Split(params, ",")
	p.Batch(&SampleSorter{tags})
}

func merge_headers(p *SamplePipeline, _ string) {
	p.Add(NewMultiHeaderMerger())
}

func normalize_min_max(p *SamplePipeline, _ string) {
	p.Batch(new(MinMaxScaling))
}

func normalize_standardize(p *SamplePipeline, _ string) {
	p.Batch(new(StandardizationScaling))
}

func pick_x_percent(p *SamplePipeline, params string) {
	pick_percentage, err := strconv.ParseFloat(params, 64)
	if err != nil {
		log.Fatalln("Failed to parse parameter for -e pick:", err)
	}
	counter := float64(0)
	p.Add(&SampleFilter{
		IncludeFilter: func(inSample *sample.Sample) bool {
			counter += pick_percentage
			if counter > 1.0 {
				counter -= 1.0
				return true
			}
			return false
		},
	})
}

func filter_metrics(p *SamplePipeline, _ string) {
	filter := NewMetricFilter()
	for _, include := range metric_filter_include {
		filter.IncludeRegex(include)
	}
	for _, exclude := range metric_filter_exclude {
		filter.ExcludeRegex(exclude)
	}
	p.Add(filter)
}

func filter_tag(p *SamplePipeline, params string) {
	val := ""
	equals := true
	index := strings.Index(params, "!=")
	if index >= 0 {
		val = params[index+2:]
		equals = false
	} else {
		index = strings.IndexRune(params, '=')
		if index == -1 {
			log.Fatalln("Parameter for -e filter_tag must be '<tag>=<value>' or '<tag>!=<value>'")
		} else {
			val = params[index+1:]
		}
	}
	tag := params[:index]
	p.Add(&SampleFilter{
		IncludeFilter: func(inSample *sample.Sample) bool {
			if equals {
				return inSample.Tag(tag) == val
			} else {
				return inSample.Tag(tag) != val
			}
		},
	})
}

func plot(pipe *SamplePipeline, params string) {
	do_plot(pipe, params, false)
}

func separate_plots(pipe *SamplePipeline, params string) {
	do_plot(pipe, params, true)
}

func do_plot(pipe *SamplePipeline, params string, separatePlots bool) {
	index := strings.IndexRune(params, ',')
	tag := ""
	filename := params
	if index == -1 {
		log.Warnln("-e plot got no tag parameter, not coloring plot (-e plot,<tag>,<filename>)")
	} else {
		tag = params[:index]
		filename = params[index+1:]
	}
	pipe.Add(&Plotter{OutputFile: filename, ColorTag: tag, SeparatePlots: separatePlots})
}

func decouple_samples(pipe *SamplePipeline, params string) {
	buf := 150000
	if params != "" {
		var err error
		if buf, err = strconv.Atoi(params); err != nil {
			log.Fatalln("Failed to parse parameter for -e decouple:", err)
		}
	} else {
		log.Println("No parameter for -e decouple, default channel buffer:", buf)
	}
	pipe.Add(&DecouplingProcessor{ChannelBuffer: buf})
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
