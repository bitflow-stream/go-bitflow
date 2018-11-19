package recovery

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow"
	log "github.com/sirupsen/logrus"
)

type nodeEvaluationData struct {
	name      string
	anomalies []*EvaluatedAnomalyEvent
	normal    []*bitflow.SampleAndHeader

	normalIndex int // When sending normal data, continuously loop through the slice of normal samples
}

type EvaluatedAnomalyEvent struct {
	samples          []*bitflow.SampleAndHeader
	node             *nodeEvaluationData
	state            string
	expectedRecovery string

	resolved bool
	history  []RecoveringAttempt
	start    time.Time
	end      time.Time

	sentAnomalySamples    int
	waitingForRecovery    *golib.BoolCondition
	waitingForNormalState *golib.BoolCondition
}

func (e *EvaluatedAnomalyEvent) attemptedRecoveries() map[string]int {
	res := make(map[string]int)
	for _, attempt := range e.history {
		res[attempt.recovery] = res[attempt.recovery] + 1
	}
	return res
}

type RecoveringAttempt struct {
	recovery            string
	executionSuccessful bool
	duration            time.Duration
}

type EvaluationProcessor struct {
	bitflow.NoopProcessor
	ConfigurableTags
	Execution *MockExecutionEngine

	data map[string]*nodeEvaluationData // Key: node name

	StoreNormalSamples     int
	PauseBetweenAnomalies  time.Duration // Simulated time span before evaluating the next anomaly
	AnomalyRecoveryTimeout time.Duration // Should match the duration after which the decision engine assumes a recovery timed out
	RecoveriesPerState     float64       // >1 means there are "non-functioning" recoveries, <1 means some recoveries handle multiple states
	MaxRecoveryAttempts    int           // After this number of recoveries, give up waiting for the correct recovery

	now            time.Time
	currentAnomaly *EvaluatedAnomalyEvent
}

func (p *EvaluationProcessor) String() string {
	return fmt.Sprintf("Evaluate decision maker (time-between-anomalies %v, anomaly-recovery-timeout %v, recoveries-per-state %v, max-recovery-attempts %v)",
		p.PauseBetweenAnomalies, p.AnomalyRecoveryTimeout, p.RecoveriesPerState, p.MaxRecoveryAttempts)
}

func (p *EvaluationProcessor) Close() {
	start := time.Now()
	p.now = start
	p.Execution.Events = p.executionEventCallback
	p.prepareEvaluation()
	p.runEvaluation()
	p.outputResults()
	log.Println("Evaluation finished, now outputting results. Total simulated time:", p.now.Sub(start))
	p.NoopProcessor.Close()
}

func (p *EvaluationProcessor) storeEvaluationEvent(node, state string, samples []*bitflow.Sample, header *bitflow.Header) {
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
			node:                  data,
			state:                 state,
			waitingForRecovery:    golib.NewBoolCondition(),
			waitingForNormalState: golib.NewBoolCondition(),
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

func (p *EvaluationProcessor) prepareEvaluation() {
	log.Printf("Received evaluation data for %v node(s):", len(p.data))
	p.iterate(func(nodeName string, node *nodeEvaluationData) {
		log.Printf(" - %v: %v anomalies (normal samples: %v)", nodeName, len(node.anomalies), len(node.normal))
	})

	// Assign expected recoveries to the anomaly states
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

	log.Printf("Running evaluation of %v total states and %v total recoveries:", len(stateRecoveries), numRecoveries)
	for state, recovery := range stateRecoveries {
		log.Printf(" - %v recovered by %v", state, recovery)
	}
}

func (p *EvaluationProcessor) runEvaluation() {
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
				log.Warnf("Cannot evaluate event %v of %v for node %v (state %v): no anomaly samples", i+1, len(node.anomalies), nodeName, anomaly.state)
				continue
			}
			log.Printf("Evaluating node %v event %v of %v (%v samples, state %v)...", nodeName, i+1, len(node.anomalies), len(anomaly.samples), anomaly.state)

			// Run the anomaly and wait for a successful recovery (or timeout)
			p.currentAnomaly = anomaly
			anomaly.start = p.now
			p.runAnomaly()
			anomaly.end = p.now

			if len(anomaly.history) >= p.MaxRecoveryAttempts {
				// This is undesirable, since the recovery engine will think that the last recovery was successful
				log.Warnf("Recovery of anomaly %v (nr %v) of node %v timed out after %v attempts. Expected %v, but attempted recoveries: %v",
					anomaly.state, i+1, nodeName, len(anomaly.history), anomaly.expectedRecovery, anomaly.attemptedRecoveries())
			}

			// Send the node back to normal state and skip some time before the next anomaly
			p.sendNormalSample(anomaly.node)
			p.waitForNormalState()
			p.progress(p.PauseBetweenAnomalies / 2)
			p.sendNormalSample(anomaly.node)
			p.progress(p.PauseBetweenAnomalies / 2)
		}
	})
}

func (p *EvaluationProcessor) runAnomaly() {
	anomaly := p.currentAnomaly
	for len(anomaly.history) < p.MaxRecoveryAttempts {

		// Send one anomaly sample and wait for the recovery.
		// Not accurate for evolving anomalies like memory leaks (e.g. cannot really restart a memory leak)...
		selectionStart := time.Now()
		sampleDuration := p.sendAnomalySample()

		p.waitForRecovery()
		attempt := p.currentAnomaly.history[len(p.currentAnomaly.history)-1]
		success := attempt.executionSuccessful && p.currentAnomaly.expectedRecovery == attempt.recovery

		if success {
			selectionTime := time.Now().Sub(selectionStart)
			p.progress(attempt.duration + selectionTime)
			anomaly.resolved = true
			break
		} else {
			// Send so anomaly samples, until the next sample would time out the current recovery, leading to the next
			// recovery being executed.
			p.progress(sampleDuration) // Progress after the anomaly sample sent above
			anomalyTimeout := p.now.Add(p.AnomalyRecoveryTimeout)
			for p.now.Before(anomalyTimeout) {
				sampleDuration = p.sendAnomalySample()
				p.progress(sampleDuration)
			}
		}
	}
}

func (p *EvaluationProcessor) waitForRecovery() {
	p.currentAnomaly.waitingForRecovery.WaitAndUnset()
}

func (p *EvaluationProcessor) waitForNormalState() {
	p.currentAnomaly.waitingForNormalState.WaitAndUnset()
}

func (p *EvaluationProcessor) sendAnomalySample() time.Duration {
	index := p.currentAnomaly.sentAnomalySamples
	if index >= len(p.currentAnomaly.samples) {
		// After the anomaly ends, keep sending the last sample (do not revert evolving anomalies like memory leaks)
		index = len(p.currentAnomaly.samples) - 1
	}
	p.currentAnomaly.sentAnomalySamples++
	sample := p.currentAnomaly.samples[index]

	var time1, time2 time.Time
	if index == len(p.currentAnomaly.samples)-1 {
		// Keep progressing the duration between the last two samples
		time1, time2 = p.currentAnomaly.samples[index-1].Time, sample.Time
	} else {
		time1, time2 = sample.Time, p.currentAnomaly.samples[index+1].Time
	}
	if !time1.Before(time2) {
		// Sanity check
		panic(fmt.Sprintf("Samples have non-continuous timestamps: %v -> %v", time1, time2))
	}

	p.doSendSample(sample)

	// Send normal-behavior samples for all other nodes
	p.iterate(func(nodeName string, otherNode *nodeEvaluationData) {
		if otherNode != p.currentAnomaly.node && len(otherNode.normal) > 0 {
			p.sendNormalSample(otherNode)
		}
	})
	return time2.Sub(time1)
}

func (p *EvaluationProcessor) sendNormalSample(node *nodeEvaluationData) {
	if len(node.normal) == 0 {
		log.Errorf("Cannot send normal data for node %v: no normal samples available", node.name)
		return
	}
	normal := node.normal[node.normalIndex%len(node.normal)]
	node.normalIndex++
	p.doSendSample(normal)
}

func (p *EvaluationProcessor) doSendSample(sample *bitflow.SampleAndHeader) {
	sampleCopy := sample.Sample.Clone()
	sampleCopy.Time = p.now
	err := p.NoopProcessor.Sample(sampleCopy, sample.Header)
	if err != nil {
		log.Errorf("DecisionMaker evaluation: error sending sample: %v", err)
	}
}

func (p *EvaluationProcessor) progress(dur time.Duration) {
	p.now = p.now.Add(dur)
}

func (p *EvaluationProcessor) executionEventCallback(node string, recovery string, success bool, duration time.Duration) {
	log.Debugf("Executed recovery %v for node %v, success: %v, duration: %v (expected recovery: %v)", recovery, node, success, duration, p.currentAnomaly.expectedRecovery)
	p.currentAnomaly.history = append(p.currentAnomaly.history, RecoveringAttempt{
		recovery:            recovery,
		duration:            duration,
		executionSuccessful: success,
	})
	p.currentAnomaly.waitingForRecovery.Broadcast()
}

func (p *EvaluationProcessor) nodeStateChangeCallback(nodeName string, oldState, newState State, timestamp time.Time) {
	if p.currentAnomaly.node.name == nodeName && newState == StateNormal {
		p.currentAnomaly.waitingForNormalState.Broadcast()
	}
}

func (p *EvaluationProcessor) outputResults() {
	header := &bitflow.Header{Fields: []string{"node_event", "total_node_events", "state_event", "total_state_events",
		"resolved", "recovery_attempts", "anomaly_samples", "recovery_duration_seconds", "execution_time_seconds"}}
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
	nodes := make([]string, 0, len(p.data))
	for name := range p.data {
		nodes = append(nodes, name)
	}
	sort.Strings(nodes)
	for _, nodeName := range nodes {
		do(nodeName, p.data[nodeName])
	}
}
