package steps

import (
	"fmt"
	"io"
	"strings"

	"github.com/bitflow-stream/go-bitflow/bitflow"
)

const PrometheusMarshallingFormat = bitflow.MarshallingFormat("prometheus")

func RegisterPrometheusMarshaller(endpoints *bitflow.EndpointFactory) {
	endpoints.Marshallers[PrometheusMarshallingFormat] = func() bitflow.Marshaller {
		return new(PrometheusMarshaller)
	}
}

// PrometheusMarshaller marshals Headers and Samples to the prometheus exposition format
type PrometheusMarshaller struct {
}

// String implements the Marshaller interface.
func (PrometheusMarshaller) String() string {
	return string(PrometheusMarshallingFormat)
}

// ShouldCloseAfterFirstSample defines that prometheus streams should close after first sent sample
func (PrometheusMarshaller) ShouldCloseAfterFirstSample() bool {
	return true
}

// WriteHeader implements the Marshaller interface. It is empty, because
// the prometheus exposition format doesn't need one
func (PrometheusMarshaller) WriteHeader(header *bitflow.Header, withTags bool, output io.Writer) error {
	return nil
}

// WriteSample implements the Marshaller interface. See the PrometheusMarshaller godoc
// for information about the format.
func (m PrometheusMarshaller) WriteSample(sample *bitflow.Sample, header *bitflow.Header, withTags bool, writer io.Writer) error {
	for i, value := range sample.Values {
		line := fmt.Sprintf("%s\t%f\n",
			m.renderMetricLine(header.Fields[i], "all"),
			value,
		)

		_, err := writer.Write([]byte(line))
		if err != nil {
			return err
		}
	}
	return nil
}

// renderMetricLine retrieves a sample field and renders a proper prometheus metric out of it
func (m PrometheusMarshaller) renderMetricLine(line string, group string) string {
	defaultLine := fmt.Sprintf(
		"%s{group=\"%s\"}",
		m.stripDashes(strings.Replace(line, "/", "_", -1)),
		group,
	)

	parts := strings.Split(line, "/")

	numParts := len(parts)
	if numParts == 1 {
		return defaultLine
	}

	switch parts[0] {
	case "disk-io", "disk-usage":
		return fmt.Sprintf("%s_%s{group=\"%s\"}", m.stripDashes(parts[0]), parts[2], group)
	case "disk":
		return fmt.Sprintf("disk_io_%s{group=\"%s\"}", parts[1], group)
	case "load":
		return fmt.Sprintf("load{minutes=\"%s\"}", parts[1])
	case "mem":
		return fmt.Sprintf("mem_%s{group=\"%s\"}", parts[1], group)
	case "net-io":
		if numParts == 2 {
			return fmt.Sprintf("%s_%s{group=\"%s\"}", m.stripDashes(parts[0]), parts[1], group)
		} else {
			nic := parts[2]
			return fmt.Sprintf("%s_%s{group=\"%s\", nic=\"%s\"}", m.stripDashes(parts[0]), parts[3], group, nic)
		}
	case "net-proto":
		return fmt.Sprintf("%s{group=\"%s\"}", m.stripDashes(strings.Join(parts, "_")), group)
	case "proc":
		newParts := parts[2:]
		return m.renderMetricLine(strings.Join(newParts, "/"), parts[1])
	}

	return defaultLine
}

func (PrometheusMarshaller) stripDashes(s string) string {
	return strings.Replace(s, "-", "_", -1)
}
