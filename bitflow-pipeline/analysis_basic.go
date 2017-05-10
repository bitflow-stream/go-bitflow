package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
	. "github.com/antongulenko/go-bitflow-pipeline/fork"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
)

func RegisterBasicAnalyses(b *query.PipelineBuilder) {
	// Control execution
	b.RegisterAnalysis("noop", noop_processor, "Pass samples through without modification")
	b.RegisterAnalysisParamsErr("sleep", sleep_samples, "Between every two samples, sleep the time difference between their timestamps", []string{}, "time")
	b.RegisterAnalysisParams("batch", generic_batch, "Collect samples and flush them when the given tag changes its value. Affects the follow-up analysis step, if it is also a batch analysis", []string{"tag"})
	b.RegisterAnalysisParamsErr("decouple", decouple_samples, "Start a new concurrent routine for handling samples. The parameter is the size of the FIFO-buffer for handing over the samples", []string{"batch"})

	b.RegisterAnalysisParams("split_files", split_files, "Split the samples into multiple files, one file per value of the given tag. Must be used as last step before a file output", []string{"tag"})
	b.RegisterAnalysisParamsErr("do", general_expression, "Execute the given expression on every sample", []string{"expr"})

	b.RegisterAnalysisParamsErr("subprocess", run_subprocess, "Start a subprocess for processing samples. Samples will be sent/received over std in/out in the given format (default: binary).", []string{"cmd"}, "format")

	// Forks
	b.RegisterFork("multiplex", fork_multiplex, "Multiplex fork copies each incoming sample to a fixed number of forked sub-pipelines", []string{"num"})
	b.RegisterFork("rr", fork_round_robin, "The round-robin fork distributes the samples equally to a fixed number of sub-pipelines", []string{"num"})
	b.RegisterFork("remap", fork_remap, "The remap-fork can be used after another fork to remap the incoming sub-pipelines to new outgoing sub-pipelines", nil)

	// Set metadata
	b.RegisterAnalysisParamsErr("listen_tags", add_listen_tags, "Listen for HTTP requests on the given port at /tag to configure tags. URL parameters are tag key-value pairs. URL parameter 'timeout' is in seconds.", []string{"port"})
	b.RegisterAnalysisParams("tags", set_tags, "Set the given tags on every sample", nil)
	b.RegisterAnalysis("set_time", set_time_processor, "Set the timestamp on every processed sample to the current time")

	// Select
	b.RegisterAnalysisParamsErr("pick", pick_x_percent, "Forward only a percentage of samples, parameter is in the range 0..1", []string{"percent"})
	b.RegisterAnalysisParamsErr("head", pick_head, "Forward only a number of the first processed samples", []string{"num"})
	b.RegisterAnalysisParamsErr("filter", filter_expression, "Filter the samples based on a boolean expression", []string{"expr"})

	// Reorder
	b.RegisterAnalysis("shuffle", shuffle_data, "Shuffle a batch of samples to a random ordering")
	b.RegisterAnalysisParams("sort", sort_data, "Sort a batch of samples based on the values of the given comma-separated tags. The default criterion is the timestmap", []string{}, "tags")

	// Change values
	b.RegisterAnalysis("scale_min_max", normalize_min_max, "Normalize a batch of samples using a min-max scale")
	b.RegisterAnalysis("standardize", normalize_standardize, "Normalize a batch of samples based on the mean and std-deviation")

	// Change header/metrics
	b.RegisterAnalysisParams("remap", remap_metrics, "Change (reorder) the header to the given comma-separated list of metrics", []string{"header"})
	b.RegisterAnalysisParamsErr("rename", rename_metrics, "Find the keys (regexes) in every metric name and replace the matched parts with the given values", nil)
	b.RegisterAnalysisParamsErr("include", filter_metrics_include, "Match every metric with the given regex and only include the matched metrics", []string{"m"})
	b.RegisterAnalysisParamsErr("exclude", filter_metrics_exclude, "Match every metric with the given regex and exclude the matched metrics", []string{"m"})
	b.RegisterAnalysisParamsErr("filter_variance", filter_variance, "In a batch of samples, filter out the metrics with a variance lower than the given theshold (based on the weighted stdev of the population, stddev/mean)", []string{"min"})
	b.RegisterAnalysisParamsErr("avg", aggregate_avg, "Add an average metric for every incoming metric. Optional parameter: duration or number of samples", []string{}, "window")
	b.RegisterAnalysisParamsErr("slope", aggregate_slope, "Add a slope metric for every incoming metric. Optional parameter: duration or number of samples", []string{}, "window")
	b.RegisterAnalysis("merge_headers", merge_headers, "Accept any number of changing headers and merge them into one output header when flushing the results")
	b.RegisterAnalysis("strip", strip_metrics, "Remove all metrics, only keeping the timestamp and the tags of eacy sample")
}

func parameterError(name string, err error) error {
	return fmt.Errorf("Failed to parse '%v' parameter: %v", name, err)
}

func noop_processor(p *SamplePipeline) {
	p.Add(new(NoopProcessor))
}

func shuffle_data(p *SamplePipeline) {
	p.Batch(NewSampleShuffler())
}

func sort_data(p *SamplePipeline, params map[string]string) {
	var tags []string
	if tags_param, ok := params["tags"]; ok {
		tags = strings.Split(tags_param, ",")
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

func pick_x_percent(p *SamplePipeline, params map[string]string) error {
	pick_percentage, err := strconv.ParseFloat(params["percent"], 64)
	if err != nil {
		return parameterError("percent", err)
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
	return nil
}

func filter_metrics_include(p *SamplePipeline, params map[string]string) error {
	filter, err := NewMetricFilter().IncludeRegex(params["m"])
	if err == nil {
		p.Add(filter)
	}
	return err
}

func filter_metrics_exclude(p *SamplePipeline, params map[string]string) error {
	filter, err := NewMetricFilter().ExcludeRegex(params["m"])
	if err == nil {
		p.Add(filter)
	}
	return err
}

func filter_expression(p *SamplePipeline, params map[string]string) error {
	return add_expression(p, params, true)
}

func general_expression(p *SamplePipeline, params map[string]string) error {
	return add_expression(p, params, false)
}

func add_expression(p *SamplePipeline, params map[string]string, filter bool) error {
	proc := &ExpressionProcessor{Filter: filter}
	err := proc.AddExpression(params["expr"])
	if err == nil {
		p.Add(proc)
	}
	return err
}

func decouple_samples(p *SamplePipeline, params map[string]string) error {
	buf, err := strconv.Atoi(params["batch"])
	if err != nil {
		err = parameterError("batch", err)
	} else {
		p.Add(&DecouplingProcessor{ChannelBuffer: buf})
	}
	return err
}

func remap_metrics(p *SamplePipeline, params map[string]string) {
	metrics := strings.Split(params["header"], ",")
	p.Add(NewMetricMapper(metrics))
}

func filter_variance(p *SamplePipeline, params map[string]string) error {
	variance, err := strconv.ParseFloat(params["min"], 64)
	if err != nil {
		err = parameterError("min", err)
	} else {
		p.Batch(NewMetricVarianceFilter(variance))
	}
	return err
}

func pick_head(p *SamplePipeline, params map[string]string) error {
	num, err := strconv.Atoi(params["num"])
	if err != nil {
		err = parameterError("num", err)
	} else {
		processed := 0
		p.Add(&SimpleProcessor{
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
	return err
}

func add_listen_tags(p *SamplePipeline, params map[string]string) error {
	port, err := strconv.Atoi(params["port"])
	if err != nil {
		return parameterError("port", err)
	}
	p.Add(&HttpTagger{Port: port})
	return nil
}

func set_tags(p *SamplePipeline, params map[string]string) {
	p.Add(NewTaggingProcessor(params))
}

func split_files(p *SamplePipeline, params map[string]string) {
	distributor := &TagsDistributor{
		Tags:        []string{params["tag"]},
		Separator:   "-",
		Replacement: "_empty_",
	}
	p.Add(&MetricFork{
		ParallelClose: true,
		Distributor:   distributor,
		Builder:       MultiFileSuffixBuilder(nil),
	})
}

func rename_metrics(p *SamplePipeline, params map[string]string) error {
	if len(params) == 0 {
		return errors.New("Need at least one regex=replacement parameter")
	}

	var regexes []*regexp.Regexp
	var replacements []string
	for regex, replacement := range params {
		r, err := regexp.Compile(regex)
		if err != nil {
			return parameterError(regex, err)
		}
		regexes = append(regexes, r)
		replacements = append(replacements, replacement)
	}
	p.Add(NewMetricRenamer(regexes, replacements))
	return nil
}

func strip_metrics(p *SamplePipeline) {
	p.Add(&SimpleProcessor{
		Description: "remove metric values, keep timestamp and tags",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			return sample.Metadata().NewSample(nil), header.Clone(nil), nil
		},
	})
}

func sleep_samples(p *SamplePipeline, params map[string]string) error {
	var timeout time.Duration
	hasTimeout := false
	if timeoutStr, ok := params["time"]; ok {
		hasTimeout = true
		var err error
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			return parameterError("time", err)
		}
	}

	desc := "sleep between samples"
	if hasTimeout {
		desc += fmt.Sprintf(" (%v)", timeout)
	} else {
		desc += " (timestamp difference)"
	}

	// TODO make this sleep interruptible

	var lastTimestamp time.Time
	p.Add(&SimpleProcessor{
		Description: desc,
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if hasTimeout {
				time.Sleep(timeout)
			} else {
				last := lastTimestamp
				if !last.IsZero() {
					diff := sample.Time.Sub(last)
					if diff > 0 {
						time.Sleep(diff)
					}
				}
				lastTimestamp = sample.Time
			}
			return sample, header, nil
		},
	})
	return nil
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

func aggregate_avg(p *SamplePipeline, params map[string]string) error {
	agg, err := create_aggregator(params)
	if err != nil {
		return err
	}
	p.Add(agg.AddAvg("_avg"))
	return nil
}

func aggregate_slope(p *SamplePipeline, params map[string]string) error {
	agg, err := create_aggregator(params)
	if err != nil {
		return err
	}
	p.Add(agg.AddSlope("_slope"))
	return nil
}

func create_aggregator(params map[string]string) (*FeatureAggregator, error) {
	window, haveWindow := params["window"]
	if !haveWindow {
		return &FeatureAggregator{}, nil
	}
	dur, err1 := time.ParseDuration(window)
	if err1 == nil {
		return &FeatureAggregator{WindowDuration: dur}, nil
	}
	num, err2 := strconv.Atoi(window)
	if err2 == nil {
		return &FeatureAggregator{WindowSize: num}, nil
	}
	return nil, parameterError("window", golib.MultiError{err1, err2})
}

func generic_batch(p *SamplePipeline, params map[string]string) {
	p.Add(&BatchProcessor{
		FlushTag: params["tag"],
	})
}

func run_subprocess(p *SamplePipeline, params map[string]string) error {
	cmd := SplitShellCommand(params["cmd"])
	format, ok := params["format"]
	if !ok {
		format = "bin"
	}
	runner := &SubprocessRunner{
		Cmd:  cmd[0],
		Args: cmd[1:],
	}
	if err := runner.Configure(format, &builder.Endpoints); err != nil {
		return err
	}
	p.Add(runner)
	return nil
}

func fork_multiplex(params map[string]string) (fmt.Stringer, error) {
	num, err := strconv.Atoi(params["num"])
	if err != nil {
		return nil, err
	}
	return NewMultiplexDistributor(num), nil
}

func fork_round_robin(params map[string]string) (fmt.Stringer, error) {
	num, err := strconv.Atoi(params["num"])
	if err != nil {
		return nil, err
	}
	return &RoundRobinDistributor{
		NumSubPipelines: num,
	}, nil
}

func fork_remap(params map[string]string) (fmt.Stringer, error) {
	return &StringRemapDistributor{
		Mapping: params,
	}, nil
}
