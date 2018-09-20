package recovery

import (
	"bytes"
	"fmt"
	"time"

	"github.com/antongulenko/go-bitflow"
)

type History interface {
	StoreAnomaly(anomaly *AnomalyEvent, executions []*ExecutionEvent)
	GetAnomalies(node string) []*AnomalyEvent
	GetExecutions(anomaly *AnomalyEvent) []*ExecutionEvent
}

type AnomalyEvent struct {
	Node     string
	Features []AnomalyFeature
	Start    time.Time
	End      time.Time
}

func (e *AnomalyEvent) IsResolved() bool {
	return e != nil && !e.Start.IsZero() && !e.End.IsZero() && e.End.After(e.Start)
}

type ExecutionEvent struct {
	Node       string
	Recovery   string
	Started    time.Time
	Duration   time.Duration
	Ended      time.Time // Time when the anomaly was reverted or when this recovery timed out, depending on the Successful flag
	Successful bool
	Error      error // If the recovery failed to execute
}

type AnomalyFeature struct {
	Name  string
	Value float64
}

func SampleToAnomalyFeatures(sample *bitflow.Sample, header *bitflow.Header) []AnomalyFeature {
	res := make([]AnomalyFeature, len(header.Fields))
	for i, field := range header.Fields {
		res[i] = AnomalyFeature{
			Name:  field,
			Value: float64(sample.Values[i]),
		}
	}
	return res
}

type VolatileHistory struct {
	anomalies  map[string][]*AnomalyEvent
	executions map[*AnomalyEvent][]*ExecutionEvent
}

func (h *VolatileHistory) StoreAnomaly(anomaly *AnomalyEvent, executions []*ExecutionEvent) {
	if h.anomalies == nil {
		h.anomalies = make(map[string][]*AnomalyEvent)
		h.executions = make(map[*AnomalyEvent][]*ExecutionEvent)
	}
	h.anomalies[anomaly.Node] = append(h.anomalies[anomaly.Node], anomaly)
	h.executions[anomaly] = executions
}

func (h *VolatileHistory) GetAnomalies(node string) []*AnomalyEvent {
	return h.anomalies[node]
}

func (h *VolatileHistory) GetExecutions(anomaly *AnomalyEvent) []*ExecutionEvent {
	return h.executions[anomaly]
}

func (h *VolatileHistory) String() string {
	var buf bytes.Buffer
	for node, events := range h.anomalies {
		fmt.Fprintf(&buf, " - Node %v\n", node)
		for eventNr, event := range events {
			fmt.Fprintf(&buf, "   - %v. state (%v - %v)\n", eventNr, event.Start, event.End)
			for executionNr, execution := range h.executions[event] {
				fmt.Fprintf(&buf, "     - %v. execution: %v (success %v, %v - %v, duration %v, error: %v)\n",
					executionNr, execution.Recovery, execution.Successful,
					execution.Started.Format(bitflow.TextMarshallerDateFormat), execution.Ended.Format(bitflow.TextMarshallerDateFormat),
					execution.Duration, execution.Error)
			}
		}
	}
	return buf.String()
}
