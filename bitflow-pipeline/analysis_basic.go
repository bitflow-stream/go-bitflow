package main

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/golib"
)

var (
	metric_filter_include golib.StringSlice
	metric_filter_exclude golib.StringSlice
)

func init() {
	RegisterSampleHandler("src", &SampleTagger{SourceTags: []string{SourceTag}})
	RegisterSampleHandler("src-append", &SampleTagger{SourceTags: []string{SourceTag}, DontOverwrite: true})

	RegisterAnalysisParams("decouple", decouple_samples, "number of buffered samples")
	RegisterAnalysis("merge_headers", merge_headers)
	RegisterAnalysisParams("pick", pick_x_percent, "samples to keep 0..1")
	RegisterAnalysisParams("head", pick_head, "number of first samples to keep")
	RegisterAnalysis("print", print_samples)
	RegisterAnalysisParams("filter_tag", filter_tag, "tag=value or tag!=value")

	RegisterAnalysis("shuffle", shuffle_data)
	RegisterAnalysisParams("sort", sort_data, "comma-separated list of tags")

	RegisterAnalysis("scale_min_max", normalize_min_max)
	RegisterAnalysis("standardize", normalize_standardize)

	RegisterAnalysisParams("plot", plot, "[<color tag>,]<output filename>")
	RegisterAnalysisParams("plot_separate", separate_plots, "same as plot")
	RegisterAnalysisParams("stats", feature_stats, "output filename for metric statistics")

	RegisterAnalysisParams("remap", remap_features, "comma-separated list of metrics")
	RegisterAnalysisParams("filter_variance", filter_variance, "minimum weighted stddev of the population (stddev / mean)")

	RegisterAnalysisParams("tags", set_tags, "comma-separated list of key-value tags")
	RegisterAnalysisParams("rename", rename_metrics, "comma-separated list of regex=replace pairs")

	RegisterAnalysis("filter_metrics", filter_metrics)
	flag.Var(&metric_filter_include, "metrics_include", "Include regex used with '-e filter_metrics'")
	flag.Var(&metric_filter_exclude, "metrics_exclude", "Exclude regex used with '-e filter_metrics'")
}

func print_samples(p *SamplePipeline) {
	p.Add(new(SamplePrinter))
}

func shuffle_data(p *SamplePipeline) {
	p.Batch(new(SampleShuffler))
}

func sort_data(p *SamplePipeline, params string) {
	var tags []string
	if params != "" {
		tags = strings.Split(params, ",")
	}
	p.Batch(&SampleSorter{tags})
}

func merge_headers(p *SamplePipeline) {
	p.Add(NewMultiHeaderMerger())
}

func normalize_min_max(p *SamplePipeline) {
	p.Batch(new(MinMaxScaling))
}

func normalize_standardize(p *SamplePipeline) {
	p.Batch(new(StandardizationScaling))
}

func pick_x_percent(p *SamplePipeline, params string) {
	pick_percentage, err := strconv.ParseFloat(params, 64)
	if err != nil {
		log.Fatalln("Failed to parse parameter for -e pick:", err)
	}
	counter := float64(0)
	p.Add(&SampleFilter{
		Description: fmt.Sprintf("Pick %.2f%%", pick_percentage*100),
		IncludeFilter: func(inSample *bitflow.Sample) bool {
			counter += pick_percentage
			if counter > 1.0 {
				counter -= 1.0
				return true
			}
			return false
		},
	})
}

func filter_metrics(p *SamplePipeline) {
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
	sign := "!="
	if equals {
		sign = "=="
	}
	p.Add(&SampleFilter{
		Description: fmt.Sprintf("Filter tag %v %s %v", tag, sign, val),
		IncludeFilter: func(inSample *bitflow.Sample) bool {
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
	if params == "" {
		log.Fatalln("-e plot needs parameters (-e plot,[<tag>,]<filename>)")
	}
	index := strings.IndexRune(params, ',')
	tag := ""
	filename := params
	if index == -1 {
		log.Warnln("-e plot got no tag parameter, not coloring plot (-e plot,[<tag>,]<filename>)")
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
		log.Warnln("No parameter for -e decouple, default channel buffer:", buf)
	}
	pipe.Add(&DecouplingProcessor{ChannelBuffer: buf})
}

func feature_stats(pipe *SamplePipeline, params string) {
	if params == "" {
		log.Fatalln("-e stats needs parameter: file to store feature statistics")
	} else {
		pipe.Add(NewStoreStats(params))
	}
}

func remap_features(pipe *SamplePipeline, params string) {
	var metrics []string
	if params != "" {
		metrics = strings.Split(params, ",")
	}
	pipe.Add(NewMetricMapper(metrics))
}

func filter_variance(pipe *SamplePipeline, params string) {
	variance, err := strconv.ParseFloat(params, 64)
	if err != nil {
		log.Fatalln("Error parsing parameter for -e filter_variance:", err)
	}
	pipe.Batch(NewMetricVarianceFilter(variance))
}

func pick_head(pipe *SamplePipeline, params string) {
	num, err := strconv.Atoi(params)
	if err != nil {
		log.Fatalln("Error parsing parameter for -e head:", err)
	}
	pipe.Add(&PickHead{Num: num})
}

type PickHead struct {
	bitflow.AbstractProcessor
	Num       int // parameter
	processed int // internal variable
}

func (head *PickHead) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := head.Check(sample, header); err != nil {
		return err
	}
	if head.Num > head.processed {
		head.processed++
		return head.OutgoingSink.Sample(sample, header)
	} else {
		return nil
	}
}

func (head *PickHead) String() string {
	return "Pick first " + strconv.Itoa(head.Num) + " samples"
}

type SampleTagger struct {
	SourceTags    []string
	DontOverwrite bool
}

func (h *SampleTagger) HandleHeader(header *bitflow.Header, source string) {
	header.HasTags = true
}

func (h *SampleTagger) HandleSample(sample *bitflow.Sample, source string) {
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

func set_tags(pipe *SamplePipeline, params string) {
	pairs := strings.Split(params, ",")
	keys := make([]string, len(pairs))
	values := make([]string, len(pairs))
	for i, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			log.Fatalln("Parameter for -e tags must be comma-separated key-value pairs: -e tags,KEY=VAL,KEY2=VAL2")
		}
		keys[i] = parts[0]
		values[i] = parts[1]
	}
	pipe.Add(&SampleTagProcessor{
		Keys:   keys,
		Values: values,
	})
}

type SampleTagProcessor struct {
	bitflow.AbstractProcessor
	Keys   []string
	Values []string
}

func (p *SampleTagProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := p.Check(sample, header); err != nil {
		return err
	}
	for i, key := range p.Keys {
		sample.SetTag(key, p.Values[i])
	}
	return p.OutgoingSink.Sample(sample, header)
}

type metricRenamer struct {
	bitflow.AbstractProcessor
	Regexes map[*regexp.Regexp]string
}

func (r *metricRenamer) String() string {
	return fmt.Sprintf("Metric renamer (%v regexes)", len(r.Regexes))
}

func (r *metricRenamer) Header(header *bitflow.Header) error {
	if err := r.CheckSink(); err != nil {
		return err
	} else {
		for i, field := range header.Fields {
			for regex, replace := range r.Regexes {
				field = regex.ReplaceAllString(field, replace)
				header.Fields[i] = field
			}
		}
		return r.OutgoingSink.Header(header)
	}
}

func rename_metrics(p *SamplePipeline, params string) {
	regexes := make(map[*regexp.Regexp]string)
	for i, part := range strings.Split(params, ",") {
		keyVal := strings.SplitN(part, "=", 2)
		if len(keyVal) != 2 {
			log.Fatalf("Parmameter %v for -e rename is not regex=replace: %v", i, part)
		}
		regexCode := keyVal[0]
		replace := keyVal[1]
		r, err := regexp.Compile(regexCode)
		if err != nil {
			log.Fatalf("Error compiling regex %v: %v", regexCode, err)
		}
		regexes[r] = replace
	}
	if len(regexes) == 0 {
		log.Fatalln("-e rename needs at least one regex=replace parameter (comma-separated)")
	}
	p.Add(&metricRenamer{
		Regexes: regexes,
	})
}
