package recovery

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/antongulenko/go-bitflow"
	log "github.com/sirupsen/logrus"
)

var (
	evaluationFillerHeader = &bitflow.Header{Fields: []string{}} // Empty header for samples to progress the time in the DecisionMaker
)

type EvaluationDataCollector struct {
	ConfigurableTags
	data               map[string]*nodeEvaluationData // Key: node name
	StoreNormalSamples int
}

func (p *EvaluationDataCollector) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	var node, state string
	for _, sample := range samples {
		newNode, newState := p.GetRecoveryTags(sample)
		if node == "" || state == "" {
			node, state = newNode, newState
		} else if newNode != node || newState != state {
			log.Warnf("Dropping batch, which contains multiple values for tags %v (%v and %v) and %v (%v and %v)",
				p.NodeNameTag, node, newNode, p.StateTag, state, newState)
			node = ""
			state = ""
			break
		}
	}
	if node != "" && state != "" {
		p.storeEvaluationEvent(node, state, samples, header)
	}
	return header, nil, nil
}

func (p *EvaluationDataCollector) String() string {
	return fmt.Sprintf("Collect evaluation data (%v, store-normal-samples: %v)", p.ConfigurableTags, p.StoreNormalSamples)
}

func (p *EvaluationDataCollector) storeEvaluationEvent(node, state string, samples []*bitflow.Sample, header *bitflow.Header) {
	data, ok := p.data[node]
	if !ok {
		data = &nodeEvaluationData{
			name: node,
		}
		p.data[node] = data
	}
	if state == p.NormalStateValue {
		for _, sample := range samples {
			if len(data.normal) >= p.StoreNormalSamples {
				break
			}
			data.normal = append(data.normal, &bitflow.SampleAndHeader{
				Sample: sample,
				Header: header,
			})
		}
	} else {
		anomaly := &EvaluatedAnomalyEvent{
			node:  node,
			state: state,
		}
		for _, sample := range samples {
			anomaly.samples = append(anomaly.samples, &bitflow.SampleAndHeader{
				Sample: sample,
				Header: header,
			})
		}
		data.anomalies = append(data.anomalies, anomaly)
	}
}

type nodeEvaluationData struct {
	name      string
	anomalies []*EvaluatedAnomalyEvent
	normal    []*bitflow.SampleAndHeader

	normalIndex int // When sending normal data, continuously loop through the slice of normal samples
}

type EvaluatedAnomalyEvent struct {
	samples          []*bitflow.SampleAndHeader
	node             string
	state            string
	expectedRecovery string

	resolved           bool
	history            []RecoveringAttempt
	start              time.Time
	end                time.Time
	sentAnomalySamples int
}

type RecoveringAttempt struct {
	recovery string
	success  bool
	duration time.Duration
}

type EvaluationProcessor struct {
	bitflow.NoopProcessor
	Execution *MockExecutionEngine
	collector *EvaluationDataCollector

	NormalSamplesBetweenAnomalies int
	FillerSamples                 int           // Number of samples to send between two real evaluation samples
	SampleRate                    time.Duration // Time progression between samples (both real and filler samples)
	RecoveriesPerState            float64       // >1 means there are "non-functioning" recoveries, <1 means some recoveries handle multiple states

	now            time.Time
	currentAnomaly *EvaluatedAnomalyEvent
}

func (p *EvaluationProcessor) String() string {
	return fmt.Sprintf("Evaluate decision maker (sample-rate %v, filler-samples %v, normal-samples %v, recoveries-per-state %v)",
		p.SampleRate, p.FillerSamples, p.NormalSamplesBetweenAnomalies, p.RecoveriesPerState)
}

func (p *EvaluationProcessor) Close() {
	p.now = time.Now()
	p.runEvaluation()
	p.outputResults()
	p.NoopProcessor.Close()
}

func (p *EvaluationProcessor) runEvaluation() {
	p.Execution.Events = p.executionEventCallback

	log.Printf("Received evaluation data for %v node(s):", len(p.collector.data))
	p.iterate(func(nodeName string, node *nodeEvaluationData) {
		log.Printf(" - %v: %v anomalies (normal samples: %v)", nodeName, len(node.anomalies), len(node.normal))
	})
	states, numRecoveries := p.assignExpectedRecoveries()
	log.Printf("Running evaluation of %v total states and %v total recoveries:", len(states), numRecoveries)
	for state, recovery := range states {
		log.Printf(" - %v recovered by %v", state, recovery)
	}

	p.iterate(func(nodeName string, node *nodeEvaluationData) {
		if len(node.normal) == 0 {
			log.Warnf("Cannot evaluate node %v: no normal data sample available", nodeName)
			return
		}
		if len(node.anomalies) == 0 {
			log.Warnf("Cannot evaluate node %v: no anomaly data available", nodeName)
			return
		}

		for i, anomaly := range node.anomalies {
			if len(anomaly.samples) == 0 {
				log.Warnf("Cannot evaluate event %v of %v for node %v (state %v): no anomaly events", i+1, len(node.anomalies), nodeName, anomaly.state)
				continue
			}

			log.Printf("Evaluating node %v event %v of %v (%v samples, state %v)...", nodeName, i+1, len(node.anomalies), len(anomaly.samples), anomaly.state)
			p.currentAnomaly = anomaly
			sampleIndex := 0
			anomaly.start = p.now
			for !anomaly.resolved {
				// Loop through all anomaly samples until the anomaly is resolved.
				// Not accurate for evolving anomalies like memory leaks...
				p.sendSample(anomaly.samples[sampleIndex%len(anomaly.samples)], node)
				sampleIndex++
			}
			anomaly.end = p.now
			anomaly.sentAnomalySamples = sampleIndex
			for i := 0; i < p.NormalSamplesBetweenAnomalies; i++ {
				p.sendNormalSample(node)
			}
		}
	})
}

func (p *EvaluationProcessor) assignExpectedRecoveries() (map[string]string, int) {
	allStates := make(map[string]bool)
	p.iterate(func(nodeName string, node *nodeEvaluationData) {
		for _, anomaly := range node.anomalies {
			allStates[anomaly.state] = true
		}
	})
	numStates := len(allStates)
	numRecoveries := int(p.RecoveriesPerState * float64(numStates))
	p.Execution.SetNumRecoveries(numRecoveries)
	allRecoveries := p.Execution.PossibleRecoveries("some-node") // TODO different nodes might have different recoveries
	if len(allRecoveries) != numRecoveries {
		panic(fmt.Sprintf("Execution engine delivered %v recoveries instead of %v", len(allRecoveries), numRecoveries))
	}
	allRecoveries = allRecoveries[:numRecoveries]

	// TODO allow different recoveries for different node layers/groups. Requires access to similarity or dependency model
	stateRecoveries := make(map[string]string)
	p.iterate(func(nodeName string, node *nodeEvaluationData) {
		for _, anomaly := range node.anomalies {
			state := anomaly.state
			recovery, ok := stateRecoveries[state]
			if !ok {
				recovery = allRecoveries[len(stateRecoveries)%len(allRecoveries)]
				stateRecoveries[state] = recovery
			}
			anomaly.expectedRecovery = recovery
		}
	})
	return stateRecoveries, numRecoveries
}

func (p *EvaluationProcessor) sendSample(sample *bitflow.SampleAndHeader, node *nodeEvaluationData) {
	sample.Sample.Time = p.progressTime()

	// Send the given sample for the given node
	err := p.NoopProcessor.Sample(sample.Sample, sample.Header)
	if err != nil {
		log.Errorf("DecisionMaker evaluation: error sending evaluation sample for node %v: %v", node.name, err)
		return
	}

	// Send normal-behavior samples for all other nodes
	p.iterate(func(nodeName string, otherNode *nodeEvaluationData) {
		if otherNode != node {
			p.sendNormalSample(otherNode)
		}
	})

	// Send some filler samples to progress the time between real samples
	for i := 0; i < p.FillerSamples; i++ {
		fillerSample := &bitflow.Sample{
			Time:   p.progressTime(),
			Values: []bitflow.Value{}, // No values in filler samples
		}
		err := p.NoopProcessor.Sample(fillerSample, evaluationFillerHeader)
		if err != nil {
			log.Errorf("DecisionMaker evaluation: error sending filler sample %v of %v: %v", i, p.FillerSamples, err)
			return
		}
	}
}

func (p *EvaluationProcessor) sendNormalSample(node *nodeEvaluationData) {
	if len(node.normal) == 0 {
		return
	}
	normal := node.normal[node.normalIndex%len(node.normal)]
	node.normalIndex++
	err := p.NoopProcessor.Sample(normal.Sample, normal.Header)
	if err != nil {
		log.Errorf("DecisionMaker evaluation: error sending normal-behavior sample nr %v for node %v: %v", node.normal, node.name, err)
	}
}

func (p *EvaluationProcessor) progressTime() time.Time {
	res := p.now
	p.now = res.Add(p.SampleRate)
	return res
}

func (p *EvaluationProcessor) executionEventCallback(node string, recovery string, success bool, duration time.Duration) {
	log.Debugf("Executed recovery %v for node %v, success: %v, duration: %v (expected recovery: %v)", recovery, node, success, duration, p.currentAnomaly.expectedRecovery)
	if success && p.currentAnomaly.expectedRecovery == recovery {
		p.currentAnomaly.resolved = true
	}
	p.currentAnomaly.history = append(p.currentAnomaly.history, RecoveringAttempt{
		recovery: recovery,
		duration: duration,
		success:  success,
	})
}

func (p *EvaluationProcessor) outputResults() {
	log.Println("Evaluation finished, now outputting results")
	header := &bitflow.Header{Fields: []string{"node_event", "total_node_events", "state_event", "total_state_events", "resolved", "recovery_attempts", "anomaly_samples", "recovery_duration_seconds", "recovery_sample_time_seconds"}}
	now := time.Now()

	p.iterate(func(nodeName string, node *nodeEvaluationData) {
		stateCounters := make(map[string]int)
		totalStateCounters := make(map[string]int)
		for _, anomaly := range node.anomalies {
			totalStateCounters[anomaly.state] = totalStateCounters[anomaly.state] + 1
		}

		for i, anomaly := range node.anomalies {
			resolved := 1
			if !anomaly.resolved {
				resolved = 0
			}
			var totalDuration time.Duration
			for _, recovery := range anomaly.history {
				totalDuration += recovery.duration
			}
			stateCounters[anomaly.state] = stateCounters[anomaly.state] + 1

			sample := &bitflow.Sample{
				Time: now,
				Values: []bitflow.Value{
					bitflow.Value(i + 1),
					bitflow.Value(len(node.anomalies)),
					bitflow.Value(stateCounters[anomaly.state]),
					bitflow.Value(totalStateCounters[anomaly.state]),
					bitflow.Value(resolved),
					bitflow.Value(len(anomaly.history)),
					bitflow.Value(anomaly.sentAnomalySamples),

					// TODO an exact recovery time needs some additional synchronization with the asynchronous recovery procedure
					bitflow.Value(anomaly.end.Sub(anomaly.start).Seconds()),
					bitflow.Value(totalDuration.Seconds()),
				},
			}
			sample.SetTag("node", nodeName)
			sample.SetTag("state", anomaly.state)
			sample.SetTag("node-state", nodeName+"-"+anomaly.state)
			sample.SetTag("resolved", strconv.FormatBool(anomaly.resolved))
			sample.SetTag("evaluation-results", "true")
			if err := p.NoopProcessor.Sample(sample, header); err != nil {
				log.Errorf("Error sending evaluation result sample for node %v, state %v (nr %v of %v): %v", nodeName, anomaly.state, i, len(node.anomalies), err)
			}
		}
	})
}

func (p *EvaluationProcessor) iterate(do func(nodeName string, node *nodeEvaluationData)) {
	nodes := make([]string, 0, len(p.collector.data))
	for name := range p.collector.data {
		nodes = append(nodes, name)
	}
	sort.Strings(nodes)
	for _, nodeName := range nodes {
		do(nodeName, p.collector.data[nodeName])
	}
}
