package steps

import (
	"fmt"
	"io"
	"strings"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

func RegisterGraphiteOutput(b *query.PipelineBuilder) {
	factory := &SimpleTextMarshallerFactory{
		Description: "graphite",
		NameFixer:   strings.NewReplacer("/", ".", " ", "_", "\t", "_", "\n", "_"),
		WriteValue: func(name string, val float64, sample *bitflow.Sample, writer io.Writer) error {
			_, err := fmt.Fprintf(writer, "%v %v %v\n", name, val, sample.Time.Unix())
			return err
		},
	}
	b.RegisterAnalysisParamsErr("graphite", factory.createTcpOutput, "Send metrics and/or tags to the given Graphite endpoint. Required parameter: 'target'. Optional: 'prefix'", nil)
}

func RegisterOpentsdbOutput(b *query.PipelineBuilder) {
	fixer := strings.NewReplacer("=", "_", "/", ".", " ", "_", "\t", "_", "\n", "_")
	factory := &SimpleTextMarshallerFactory{
		Description: "opentsdb",
		NameFixer:   fixer,
		WriteValue: func(name string, val float64, sample *bitflow.Sample, writer io.Writer) error {
			_, err := fmt.Fprintf(writer, "put %v %v %v", name, sample.Time.Unix(), val)
			for key, val := range sample.TagMap() {
				key = fixer.Replace(key)
				name = fixer.Replace(name)
				_, err = fmt.Fprintf(writer, " %s=%s", key, val)
				if err != nil {
					break
				}
			}
			if err == nil {
				_, err = fmt.Fprintf(writer, " bitflow=true") // Add an artificial tag, because at least one tag is required
			}
			if err == nil {
				_, err = writer.Write([]byte{'\n'})
			}
			return err
		},
	}
	b.RegisterAnalysisParamsErr("opentsdb", factory.createTcpOutput, "Send metrics and/or tags to the given OpenTSDB endpoint. Required parameter: 'target'. Optional: 'prefix'", nil)
}

var _ bitflow.Marshaller = new(SimpleTextMarshaller)

type SimpleTextMarshallerFactory struct {
	Description string
	NameFixer   *strings.Replacer
	WriteValue  func(name string, val float64, sample *bitflow.Sample, writer io.Writer) error
}

func (f *SimpleTextMarshallerFactory) createTcpOutput(p *pipeline.SamplePipeline, params map[string]string) error {
	target, hasTarget := params["target"]
	if !hasTarget {
		return query.ParameterError("target", fmt.Errorf("Missing required parameter"))
	}
	prefix := params["prefix"]
	delete(params, "target")
	delete(params, "prefix")

	sink, err := _make_tcp_output(params)
	sink.Endpoint = target
	sink.SetMarshaller(&SimpleTextMarshaller{
		MetricPrefix: prefix,
		Description:  f.Description,
		NameFixer:    f.NameFixer,
		WriteValue:   f.WriteValue,
	})
	if err == nil {
		p.Add(sink)
	}
	return err
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

type SimpleTextMarshaller struct {
	Description  string
	MetricPrefix string
	NameFixer    *strings.Replacer
	WriteValue   func(name string, val float64, sample *bitflow.Sample, writer io.Writer) error
}

func (o *SimpleTextMarshaller) String() string {
	return fmt.Sprintf("%s(prefix: %v)", o.Description, o.MetricPrefix)
}

func (o *SimpleTextMarshaller) WriteHeader(header *bitflow.Header, hasTags bool, writer io.Writer) error {
	// No separate header
	return nil
}

func (o *SimpleTextMarshaller) WriteSample(sample *bitflow.Sample, header *bitflow.Header, hasTags bool, writer io.Writer) error {
	prefix := o.MetricPrefix
	if prefix != "" {
		prefix = pipeline.ResolveTagTemplate(prefix, "_", sample)
	}

	for i, value := range sample.Values {
		name := o.NameFixer.Replace(prefix + header.Fields[i])
		if err := o.WriteValue(name, float64(value), sample, writer); err != nil {
			return err
		}
	}
	return nil
}
