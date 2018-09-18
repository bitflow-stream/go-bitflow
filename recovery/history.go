package recovery

import (
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
