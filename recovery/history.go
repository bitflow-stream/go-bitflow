package recovery

import (
	"time"

	"github.com/antongulenko/go-bitflow"
)

type History interface {
	StoreAnomaly(anomaly *Anomaly, executions []*Execution)
	GetAnomalies(node string) []*Anomaly
	GetExecutions(anomaly *Anomaly) []*Execution
}

type Anomaly struct {
	Node     string
	Features []AnomalyFeature
	Start    time.Time
	End      time.Time
}

type Execution struct {
	Node              string
	Recovery          string
	Started           time.Time
	ExecutionFinished time.Time // 0 means it is still running
	Ended             time.Time // Time when the anomaly was reverted or when this recovery timed out, depening on the Successful flag
	Successful        bool
	Error             string // If the recovery failed to execute
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
	anomalies  map[string][]*Anomaly
	executions map[*Anomaly][]*Execution
}

func (h *VolatileHistory) StoreAnomaly(anomaly *Anomaly, executions []*Execution) {
	if h.anomalies == nil {
		h.anomalies = make(map[string][]*Anomaly)
		h.executions = make(map[*Anomaly][]*Execution)
	}
	h.anomalies[anomaly.Node] = append(h.anomalies[anomaly.Node], anomaly)
	h.executions[anomaly] = executions
}

func (h *VolatileHistory) GetAnomalies(node string) []*Anomaly {
	return h.anomalies[node]
}

func (h *VolatileHistory) GetExecutions(anomaly *Anomaly) []*Execution {
	return h.executions[anomaly]
}
