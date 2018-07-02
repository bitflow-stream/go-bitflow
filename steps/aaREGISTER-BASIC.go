package steps

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/fork"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

func REGISTER_BASIC(b *query.PipelineBuilder) {

	b.RegisterAnalysisParamsErr("output_files", split_files, "Output samples to multiple files, filenames are built from the given template, where placeholders like ${xxx} will be replaced with tag values", nil)
	b.RegisterAnalysisParamsErr("do", general_expression, "Execute the given expression on every sample", []string{"expr"})

	b.RegisterAnalysisParamsErr("subprocess", run_subprocess, "Start a subprocess for processing samples. Samples will be sent/received over std in/out in the given format (default: binary)", []string{"cmd"}, "format")

	// Set metadata
	b.RegisterAnalysis("set_time", set_time_processor, "Set the timestamp on every processed sample to the current time")
	b.RegisterAnalysis("append_latency", append_time_difference, "Append the time difference to the previous sample as a metric")

	// Select
	b.RegisterAnalysisParamsErr("pick", pick_x_percent, "Forward only a percentage of samples, parameter is in the range 0..1", []string{"percent"})
	b.RegisterAnalysisParamsErr("head", pick_head, "Forward only a number of the first processed samples. If close=true is given as parameter, close the whole pipeline afterwards.", []string{"num"}, "close")
	b.RegisterAnalysisParamsErr("skip", skip_samples, "Drop a number of samples in the beginning", []string{"num"})
	b.RegisterAnalysisParamsErr("filter", filter_expression, "Filter the samples based on a boolean expression", []string{"expr"})

	// Reorder
	b.RegisterAnalysis("shuffle", shuffle_data, "Shuffle a batch of samples to a random ordering")
	b.RegisterAnalysisParams("sort", sort_data, "Sort a batch of samples based on the values of the given comma-separated tags. The default criterion is the timestmap", []string{}, "tags")

	// Change values
	b.RegisterAnalysis("scale_min_max", normalize_min_max, "Normalize a batch of samples using a min-max scale")
	b.RegisterAnalysis("standardize", normalize_standardize, "Normalize a batch of samples based on the mean and std-deviation")

	// Change header/metrics
	b.RegisterAnalysisParams("parse_tags", parse_tags_to_metrics, "Append metrics based on tag values. Keys are new metric names, values are tag names", nil)
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

func shuffle_data(p *pipeline.SamplePipeline) {
	p.Batch(NewSampleShuffler())
}

func sort_data(p *pipeline.SamplePipeline, params map[string]string) {
	var tags []string
	if tags_param, ok := params["tags"]; ok {
		tags = strings.Split(tags_param, ",")
	}
	p.Batch(&SampleSorter{tags})
}

func merge_headers(p *pipeline.SamplePipeline) {
	p.Add(NewMultiHeaderMerger())
}

func normalize_min_max(p *pipeline.SamplePipeline) {
	p.Batch(new(MinMaxScaling))
}

func normalize_standardize(p *pipeline.SamplePipeline) {
	p.Batch(new(StandardizationScaling))
}

func pick_x_percent(p *pipeline.SamplePipeline, params map[string]string) error {
	pick_percentage, err := strconv.ParseFloat(params["percent"], 64)
	if err != nil {
		return query.ParameterError("percent", err)
	}
	counter := float64(0)
	p.Add(&SampleFilter{
		Description: pipeline.String(fmt.Sprintf("Pick %.2f%%", pick_percentage*100)),
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

func filter_metrics_include(p *pipeline.SamplePipeline, params map[string]string) error {
	filter, err := NewMetricFilter().IncludeRegex(params["m"])
	if err == nil {
		p.Add(filter)
	}
	return err
}

func filter_metrics_exclude(p *pipeline.SamplePipeline, params map[string]string) error {
	filter, err := NewMetricFilter().ExcludeRegex(params["m"])
	if err == nil {
		p.Add(filter)
	}
	return err
}

func filter_expression(p *pipeline.SamplePipeline, params map[string]string) error {
	return add_expression(p, params, true)
}

func general_expression(p *pipeline.SamplePipeline, params map[string]string) error {
	return add_expression(p, params, false)
}

func add_expression(p *pipeline.SamplePipeline, params map[string]string, filter bool) error {
	proc := &ExpressionProcessor{Filter: filter}
	err := proc.AddExpression(params["expr"])
	if err == nil {
		p.Add(proc)
	}
	return err
}

func remap_metrics(p *pipeline.SamplePipeline, params map[string]string) {
	metrics := strings.Split(params["header"], ",")
	p.Add(NewMetricMapper(metrics))
}

func filter_variance(p *pipeline.SamplePipeline, params map[string]string) error {
	variance, err := strconv.ParseFloat(params["min"], 64)
	if err != nil {
		err = query.ParameterError("min", err)
	} else {
		p.Batch(NewMetricVarianceFilter(variance))
	}
	return err
}

func pick_head(p *pipeline.SamplePipeline, params map[string]string) error {
	doClose := params["close"] == "true"
	num, err := strconv.Atoi(params["num"])
	if err != nil {
		err = query.ParameterError("num", err)
	} else {
		processed := 0
		proc := &pipeline.SimpleProcessor{
			Description: "Pick first " + strconv.Itoa(num) + " samples",
		}
		proc.Process = func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if num > processed {
				processed++
				return sample, header, nil
			} else {
				if doClose {
					proc.Error(nil) // Stop processing without an error
				}
				return nil, nil, nil
			}
		}
		p.Add(proc)
	}
	return err
}

func skip_samples(p *pipeline.SamplePipeline, params map[string]string) error {
	num, err := strconv.Atoi(params["num"])
	if err != nil {
		err = query.ParameterError("num", err)
	} else {
		dropped := 0
		proc := &pipeline.SimpleProcessor{
			Description: "Drop first " + strconv.Itoa(num) + " samples",
		}
		proc.Process = func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if dropped >= num {
				return sample, header, nil
			} else {
				dropped++
				return nil, nil, nil
			}
		}
		p.Add(proc)
	}
	return err
}

func split_files(p *pipeline.SamplePipeline, params map[string]string) error {
	filename := params["file"]
	if filename == "" {
		return query.ParameterError("file", errors.New("Missing required parameter"))
	}
	delete(params, "file")

	distributor, err := make_multi_file_pipeline_builder(params)
	distributor.Template = filename
	if err == nil {
		p.Add(&fork.SampleFork{Distributor: distributor})
	}
	return err
}

func make_multi_file_pipeline_builder(params map[string]string) (*fork.MultiFileDistributor, error) {
	var endpointFactory bitflow.EndpointFactory
	if err := endpointFactory.ParseParameters(params); err != nil {
		return nil, fmt.Errorf("Error parsing parameters: %v", err)
	}
	output, err := endpointFactory.CreateOutput("file://-") // Create empty file output, will only be used as template with configuration values
	if err != nil {
		return nil, fmt.Errorf("Error creating template file output: %v", err)
	}
	fileOutput, ok := output.(*bitflow.FileSink)
	if !ok {
		return nil, fmt.Errorf("Error creating template file output, received wrong type: %T", output)
	}
	return &fork.MultiFileDistributor{Config: *fileOutput}, nil
}

func rename_metrics(p *pipeline.SamplePipeline, params map[string]string) error {
	if len(params) == 0 {
		return errors.New("Need at least one regex=replacement parameter")
	}

	var regexes []*regexp.Regexp
	var replacements []string
	for regex, replacement := range params {
		r, err := regexp.Compile(regex)
		if err != nil {
			return query.ParameterError(regex, err)
		}
		regexes = append(regexes, r)
		replacements = append(replacements, replacement)
	}
	p.Add(NewMetricRenamer(regexes, replacements))
	return nil
}

func strip_metrics(p *pipeline.SamplePipeline) {
	p.Add(&pipeline.SimpleProcessor{
		Description: "remove metric values, keep timestamp and tags",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			return sample.Metadata().NewSample(nil), header.Clone(nil), nil
		},
	})
}

func set_time_processor(p *pipeline.SamplePipeline) {
	p.Add(&pipeline.SimpleProcessor{
		Description: "reset time to now",
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			sample.Time = time.Now()
			return sample, header, nil
		},
	})
}

func aggregate_avg(p *pipeline.SamplePipeline, params map[string]string) error {
	agg, err := create_aggregator(params)
	if err != nil {
		return err
	}
	p.Add(agg.AddAvg("_avg"))
	return nil
}

func aggregate_slope(p *pipeline.SamplePipeline, params map[string]string) error {
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
	return nil, query.ParameterError("window", golib.MultiError{err1, err2})
}

func run_subprocess(p *pipeline.SamplePipeline, params map[string]string) error {
	cmd := pipeline.SplitShellCommand(params["cmd"])
	format, ok := params["format"]
	if !ok {
		format = "bin"
	}
	delete(params, "cmd")
	delete(params, "format")

	var endpointFactory bitflow.EndpointFactory
	if err := endpointFactory.ParseParameters(params); err != nil {
		return fmt.Errorf("Error parsing parameters: %v", err)
	}

	runner := &SubprocessRunner{
		Cmd:  cmd[0],
		Args: cmd[1:],
	}
	if err := runner.Configure(format, &endpointFactory); err != nil {
		return err
	}
	p.Add(runner)
	return nil
}

func parse_tags_to_metrics(p *pipeline.SamplePipeline, params map[string]string) {
	var checker bitflow.HeaderChecker
	var outHeader *bitflow.Header
	var sorted pipeline.SortedStringPairs
	warnedMissingTags := make(map[string]bool)
	sorted.FillFromMap(params)
	sort.Sort(&sorted)

	p.Add(&pipeline.SimpleProcessor{
		Description: "Convert tags to metrics: " + sorted.String(),
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if checker.HeaderChanged(header) {
				outHeader = header.Clone(append(header.Fields, sorted.Keys...))
			}
			values := make([]float64, len(sorted.Values))
			for i, tag := range sorted.Values {
				var value float64
				if !sample.HasTag(tag) {
					if !warnedMissingTags[tag] {
						warnedMissingTags[tag] = true
						log.Warnf("Encountered sample missing tag '%v'. Using metric value 0 instead. This warning is printed once per tag.", tag)
					}
				} else {
					var err error
					value, err = strconv.ParseFloat(sample.Tag(tag), 64)
					if err != nil {
						return nil, nil, fmt.Errorf("Cloud not convert '%v' tag to float64: %v", tag, err)
					}
				}
				values[i] = value
			}
			pipeline.AppendToSample(sample, values)
			return sample, outHeader, nil
		},
	})
}

func append_time_difference(p *pipeline.SamplePipeline) {
	fieldName := "time-difference"
	var checker bitflow.HeaderChecker
	var outHeader *bitflow.Header
	var lastTime time.Time

	p.Add(&pipeline.SimpleProcessor{
		Description: "Append time difference as metric " + fieldName,
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			if checker.HeaderChanged(header) {
				outHeader = header.Clone(append(header.Fields, fieldName))
			}
			var diff float64
			if !lastTime.IsZero() {
				diff = float64(sample.Time.Sub(lastTime))
			}
			lastTime = sample.Time
			pipeline.AppendToSample(sample, []float64{diff})
			return sample, outHeader, nil
		},
	})
}
