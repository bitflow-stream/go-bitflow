package steps

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterGraphiteOutput(b reg.ProcessorRegistry) {
	factory := &SimpleTextMarshallerFactory{
		Description: "graphite",
		NameFixer:   strings.NewReplacer("/", ".", " ", "_", "\t", "_", "\n", "_").Replace,
		WriteValue: func(name string, val float64, sample *bitflow.Sample, writer io.Writer) error {
			_, err := fmt.Fprintf(writer, "%v %v %v\n", name, val, sample.Time.Unix())
			return err
		},
	}
	b.RegisterStep("graphite", factory.createTcpOutput, "Send metrics and/or tags to the given Graphite endpoint.").
		Required("target", reg.String()).
		Optional("prefix", reg.String(), "").
		Optional("endpoint-config", reg.Map(reg.String()), map[string]string{})
}

func RegisterOpentsdbOutput(b reg.ProcessorRegistry) {
	const max_opentsdb_tags = 8

	nameReplacer := strings.NewReplacer("/", ".")          // Convention for bitflow metric names uses slashes, while OpenTSDB uses dots
	illegalChars := regexp.MustCompile("[^\\p{L}\\d-_./]") // \p{L} matches Unicode letters, \d matches digits. The listed characters are legal, and the entire set is negated.
	replacementString := "_"

	factory := &SimpleTextMarshallerFactory{
		Endpoints:   b.Endpoints,
		Description: "opentsdb",
		NameFixer: func(in string) string {
			in = nameReplacer.Replace(in)
			return illegalChars.ReplaceAllLiteralString(in, replacementString)
		},
		WriteValue: func(name string, val float64, sample *bitflow.Sample, writer io.Writer) error {
			_, err := fmt.Fprintf(writer, "put %v %v %f", name, sample.Time.Unix(), val)
			addedTags := 0
			for _, tag := range sample.SortedTags() {
				key := illegalChars.ReplaceAllLiteralString(tag.Key, replacementString)
				val := illegalChars.ReplaceAllLiteralString(tag.Value, replacementString)
				_, err = fmt.Fprintf(writer, " %s=%s", key, val)
				addedTags++
				if err != nil || addedTags >= max_opentsdb_tags {
					break
				}
			}
			if err == nil && addedTags == 0 {
				_, err = fmt.Fprintf(writer, " bitflow=true") // Add an artificial tag, because at least one tag is required
			}
			if err == nil {
				_, err = writer.Write([]byte{'\n'})
			}
			return err
		},
	}
	b.RegisterStep("opentsdb", factory.createTcpOutput, "Send metrics and/or tags to the given OpenTSDB endpoint.").
		Required("target", reg.String()).
		Optional("prefix", reg.String(), "").
		Optional("endpoint-config", reg.Map(reg.String()), map[string]string{})
}

var _ bitflow.Marshaller = new(SimpleTextMarshaller)

type SimpleTextMarshallerFactory struct {
	Description string
	NameFixer   func(string) string
	WriteValue  func(name string, val float64, sample *bitflow.Sample, writer io.Writer) error

	Endpoints *bitflow.EndpointFactory
}

func (f *SimpleTextMarshallerFactory) createTcpOutput(p *bitflow.SamplePipeline, params map[string]interface{}) error {
	sink, err := _make_tcp_output(params["endpoint-config"].(map[string]string), f.Endpoints)
	if err == nil {
		sink.Endpoint = params["target"].(string)
		sink.SetMarshaller(&SimpleTextMarshaller{
			MetricPrefix: params["prefix"].(string),
			Description:  f.Description,
			NameFixer:    f.NameFixer,
			WriteValue:   f.WriteValue,
		})
		p.Add(sink)
	}
	return err
}

func _make_tcp_output(endpointParams map[string]string, endpoints *bitflow.EndpointFactory) (*bitflow.TCPSink, error) {
	endpoints, err := endpoints.CloneWithParams(endpointParams)
	if err != nil {
		return nil, fmt.Errorf("Error parsing parameters: %v", err)
	}
	output, err := endpoints.CreateOutput("tcp://-") // Create empty TCP output, will only be used as template with configuration values
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
	NameFixer    func(string) string
	WriteValue   func(name string, val float64, sample *bitflow.Sample, writer io.Writer) error
}

// ShouldCloseAfterFirstSample defines that text streams can stream without closing
func (SimpleTextMarshaller) ShouldCloseAfterFirstSample() bool {
	return false
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
		prefix = bitflow.ResolveTagTemplate(prefix, "_", sample)
	}

	for i, value := range sample.Values {
		name := o.NameFixer(prefix + header.Fields[i])
		if err := o.WriteValue(name, float64(value), sample, writer); err != nil {
			return err
		}
	}
	return nil
}
