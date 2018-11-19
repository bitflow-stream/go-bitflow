package recovery

import (
	"bytes"
	"fmt"
	"time"

	"github.com/bitflow-stream/go-bitflow"
)

type History interface {
	StoreExecution(execution *ExecutionEvent)
	GetExecutions(node string) []*ExecutionEvent
}

type ExecutionEvent struct {
	Node string

	// Anomaly
	AnomalyFeatures   AnomalyData
	AnomalyStarted    time.Time // Copied from PreviousExecution, if PreviousAttempts > 0
	PreviousAttempts  int
	PreviousExecution *ExecutionEvent

	// This Recovery Execution
	Recovery          string
	Started           time.Time
	ExecutionDuration time.Duration // As reported by the ExecutionEngine
	Ended             time.Time     // Time when the anomaly was recovered or when this recovery timed out, depending on the Successful flag
	Error             error         // If the recovery failed to execute
	Successful        bool          // True, if the anomaly was resolved
}

type AnomalyFeature struct {
	Name  string
	Value float64
}

type AnomalyData []AnomalyFeature

func SampleToAnomalyFeatures(sample *bitflow.Sample, header *bitflow.Header) AnomalyData {
	res := make(AnomalyData, len(header.Fields))
	for i, field := range header.Fields {
		res[i] = AnomalyFeature{
			Name:  field,
			Value: float64(sample.Values[i]),
		}
	}
	return res
}

type VolatileHistory struct {
	executions map[string][]*ExecutionEvent
}

func (h *VolatileHistory) StoreExecution(execution *ExecutionEvent) {
	if h.executions == nil {
		h.executions = make(map[string][]*ExecutionEvent)
	}
	h.executions[execution.Node] = append(h.executions[execution.Node], execution)
}

func (h *VolatileHistory) GetExecutions(node string) []*ExecutionEvent {
	return h.executions[node]
}

func (h *VolatileHistory) String() string {
	dateFmt := "2006-01-02 15:04:05.999"
	var buf bytes.Buffer
	for node, events := range h.executions {
		fmt.Fprintf(&buf, " - Node %v\n", node)
		anomalyCounter := 0
		for eventNr, event := range events {
			if event.PreviousExecution == nil {
				fmt.Fprintf(&buf, "   - %v. anomaly (start %v)\n", anomalyCounter, event.AnomalyStarted.Format(dateFmt))
				anomalyCounter++
			}
			fmt.Fprintf(&buf, "     - %v. attempt (%v total): %v (success %v, %v - %v, duration %v, error: %v)\n",
				event.PreviousAttempts+1, eventNr+1, event.Recovery, event.Successful,
				event.Started.Format(dateFmt), event.Ended.Format(dateFmt),
				event.ExecutionDuration, event.Error)
			if event.Successful {
				fmt.Fprintf(&buf, "  - Recovered after %v", event.Ended.Sub(event.AnomalyStarted))
			}
		}
	}
	return buf.String()
}
