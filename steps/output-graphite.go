package steps

import (
	"fmt"
	"io"
	"strings"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/fork"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

var graphiteMetricNameCleaner = strings.NewReplacer("/", ".", " ", "_", "\t", "_", "\n", "_")

func RegisterGraphiteOutput(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) error {
		target, hasTarget := params["target"]
		if !hasTarget {
			return query.ParameterError("target", fmt.Errorf("Missing required parameter"))
		}
		prefix := params["prefix"]
		delete(params, "target")
		delete(params, "prefix")

		sink, err := _make_tcp_output(params)
		sink.Endpoint = target
		sink.SetMarshaller(&GraphiteMarshaller{
			MetricPrefix: prefix,
		})
		if err == nil {
			p.Add(sink)
		}
		return err
	}

	b.RegisterAnalysisParamsErr("graphite", create, "Send metrics and/or tags to the given Graphite endpoint. Required parameter: 'target'. Optional: 'prefix'", nil)
}

func _make_tcp_output(params map[string]string) (*bitflow.TCPSink, error) {
	var endpointFactory bitflow.EndpointFactory
	if err := endpointFactory.ParseParameters(params); err != nil {
		return nil, fmt.Errorf("Error parsing parameters: %v", err)
	}
	output, err := endpointFactory.CreateOutput("tcp://-") // Create empty TCP output, will only be used as template with configuration values
	if err != nil {
		return nil, fmt.Errorf("Error creating template TCP output: %v", err)
	}
	tcpOutput, ok := output.(*bitflow.TCPSink)
	if !ok {
		return nil, fmt.Errorf("Error creating template file output, received wrong type: %T", output)
	}
	return tcpOutput, nil
}

type GraphiteMarshaller struct {
	MetricPrefix string
}

var _ bitflow.Marshaller = new(GraphiteMarshaller)

func (o *GraphiteMarshaller) String() string {
	return fmt.Sprintf("graphite(prefix: %v)", o.MetricPrefix)
}

func (o *GraphiteMarshaller) WriteHeader(header *bitflow.Header, hasTags bool, writer io.Writer) error {
	// No separate header
	return nil
}

func (o *GraphiteMarshaller) WriteSample(sample *bitflow.Sample, header *bitflow.Header, hasTags bool, writer io.Writer) error {
	prefix := o.MetricPrefix
	if prefix != "" {
		templater := fork.TagTemplate{
			Template:     prefix,
			MissingValue: "_",
		}
		prefix = templater.BuildKey(sample)
	}

	for i, value := range sample.Values {
		name := header.Fields[i]
		name = prefix + name
		name = graphiteMetricNameCleaner.Replace(name)
		if err := o.writeValue(name, float64(value), sample.Time, writer); err != nil {
			return err
		}
	}
	return nil
}

func (o *GraphiteMarshaller) writeValue(name string, val float64, ts time.Time, writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "%v %v %v\n", name, val, ts.Unix())
	return err
}
