package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
)

func init() {
	// Control execution
	RegisterAnalysis("noop", noop_processor)
	RegisterAnalysis("sleep", sleep_samples)
	RegisterAnalysisParams("batch", generic_batch, "tag for triggering flush of batch")
	RegisterAnalysisParams("decouple", decouple_samples, "number of buffered samples")
	RegisterAnalysisParams("split_files", split_files, "tag to use for separating the data")
	RegisterAnalysisParams("do", general_expression, "expression to execute for each sample")

	// Set metadata
	RegisterAnalysisParams("tags", set_tags, "comma-separated list of key-value tags")
	RegisterAnalysis("set_time", set_time_processor)

	// Select
	RegisterAnalysisParams("pick", pick_x_percent, "samples to keep 0..1")
	RegisterAnalysisParams("head", pick_head, "number of first samples to keep")
	RegisterAnalysisParams("filter", filter_expression, "Filter expression including metrics as variables. Tags accessed via tag() and has_tag().")

	// Reorder
	RegisterAnalysis("shuffle", shuffle_data)
	RegisterAnalysisParams("sort", sort_data, "comma-separated list of tags. Default criterion is the timestamp.")

	// Change values
	RegisterAnalysis("scale_min_max", normalize_min_max)
	RegisterAnalysis("standardize", normalize_standardize)

	// Change header/metrics
	RegisterAnalysisParams("remap", remap_metrics, "comma-separated list of metrics")
	RegisterAnalysisParams("rename", rename_metrics, "comma-separated list of regex=replace pairs")
	RegisterAnalysisParams("include", filter_metrics_include, "Regex to match metrics to be included")
	RegisterAnalysisParams("exclude", filter_metrics_exclude, "Regex to match metrics to be excluded")
	RegisterAnalysisParams("filter_variance", filter_variance, "minimum weighted stddev of the population (stddev / mean)")
	RegisterAnalysisParams("avg", aggregate_avg, "Optional parameter: duration or number of samples.")
	RegisterAnalysisParams("slope", aggregate_slope, "Optional parameter: duration or number of samples.")
	RegisterAnalysis("merge_headers", merge_headers)
	RegisterAnalysis("strip", strip_metrics)
}

func noop_processor(p *SamplePipeline) {
	p.Add(new(bitflow.AbstractProcessor))
}

func shuffle_data(p *SamplePipeline) {
	p.Batch(NewSampleShuffler())
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
		Description: String(fmt.Sprintf("Pick %.2f%%", pick_percentage*100)),
		IncludeFilter: func(_ *bitflow.Sample, _ *bitflow.Header) (bool, error) {
			counter += pick_percentage
			if counter > 1.0 {
				counter -= 1.0
				return true, nil
			}
			return false, nil
		},
	})
}

func filter_metrics_include(p *SamplePipeline, param string) {
	p.Add(NewMetricFilter().IncludeRegex(param))
}

func filter_metrics_exclude(p *SamplePipeline, param string) {
	p.Add(NewMetricFilter().ExcludeRegex(param))
}

func filter_expression(pipe *SamplePipeline, params string) {
	add_expression(pipe, params, true)
}

func general_expression(pipe *SamplePipeline, params string) {
	add_expression(pipe, params, false)
}

func add_expression(pipe *SamplePipeline, expression string, filter bool) {
	proc := &ExpressionProcessor{Filter: filter}
	err := proc.AddExpression(expression)
	if err != nil {
		log.Fatalln(err)
	}
	pipe.Add(proc)
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

func remap_metrics(pipe *SamplePipeline, params string) {
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
	processed := 0
	pipe.Add(&SimpleProcessor{
		Description: "Pick first " + strconv.Itoa(num) + " samples",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if num > processed {
				processed++
				return sample, header, nil
			} else {
				return nil, nil, nil
			}
		},
	})
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
	pipe.Add(&SimpleProcessor{
		Description: "",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			for i, key := range keys {
				sample.SetTag(key, values[i])
			}
			return sample, header, nil
		},
	})
}

func split_files(p *SamplePipeline, params string) {
	distributor := &TagsDistributor{
		Tags:        []string{params},
		Separator:   "-",
		Replacement: "_empty_",
	}
	p.Add(NewMetricFork(distributor, MultiFileSuffixBuilder(nil)))
}

func rename_metrics(p *SamplePipeline, params string) {
	var regexes []*regexp.Regexp
	var replacements []string
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
		regexes = append(regexes, r)
		replacements = append(replacements, replace)
	}
	if len(regexes) == 0 {
		log.Fatalln("-e rename needs at least one regex=replace parameter (comma-separated)")
	}
	p.Add(NewMetricRenamer(regexes, replacements))
}

func strip_metrics(p *SamplePipeline) {
	p.Add(&SimpleProcessor{
		Description: "remove metric values, keep timestamp and tags",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			return sample.Metadata().NewSample(nil), header.Clone(nil), nil
		},
	})
}

func sleep_samples(p *SamplePipeline) {
	var lastTimestamp time.Time
	p.Add(&SimpleProcessor{
		Description: "sleep between samples",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			last := lastTimestamp
			if !last.IsZero() {
				diff := sample.Time.Sub(last)
				if diff > 0 {
					time.Sleep(diff)
				}
			}
			lastTimestamp = sample.Time
			return sample, header, nil
		},
	})
}

func set_time_processor(p *SamplePipeline) {
	p.Add(&SimpleProcessor{
		Description: "reset time to now",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			sample.Time = time.Now()
			return sample, header, nil
		},
	})
}

func aggregate_avg(p *SamplePipeline, param string) {
	p.Add(create_aggregator(param).AddAvg("_avg"))
}

func aggregate_slope(p *SamplePipeline, param string) {
	p.Add(create_aggregator(param).AddSlope("_slope"))
}

func create_aggregator(param string) *FeatureAggregator {
	if param == "" {
		return &FeatureAggregator{}
	}
	dur, err := time.ParseDuration(param)
	if err == nil {
		return &FeatureAggregator{WindowDuration: dur}
	}
	num, err := strconv.Atoi(param)
	if err == nil {
		return &FeatureAggregator{WindowSize: num}
	}
	log.Fatalf("Failed to parse aggregation parameter %v: Need either a number, or a duration", param)
	return nil
}

func generic_batch(p *SamplePipeline, param string) {
	p.Add(&BatchProcessor{
		FlushTag: param,
	})
}
