package recovery

import (
	"sync"
	"time"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

var (
	evaluationFillerHeader = &bitflow.Header{Fields: []string{}} // Empty header for samples to progress the time in the DecisionMaker
)

type EvaluationProcessor struct {
	bitflow.NoopProcessor
	ConfigurableTags
	Execution *MockExecutionEngine

	NormalSamplesBetweenAnomalies int
	FillerSamples                 int           // Number of samples to send between two real evaluation samples
	SampleRate                    time.Duration // Time progression between samples (both real and filler samples)

	RecoveriesPerState float64 // >1 means there are "non-functioning" recoveries, <1 means some recoveries handle multiple states

	data map[string]*nodeEvaluationData // Key: node name
	now  time.Time

	currentAnomaly *EvaluationEvent
}

type nodeEvaluationData struct {
	name      string
	anomalies []*EvaluationEvent
	normal    *bitflow.SampleAndHeader
}

func (p *EvaluationProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.Execution.Events = p.executionEventCallback
	return p.NoopProcessor.Start(wg)
}

func (p *EvaluationProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	node, state := p.GetRecoveryTags(sample)
	if node != "" && state != "" {
		p.storeEvaluationSample(node, state, sample, header)
	}
	return p.NoopProcessor.Sample(sample, header)
}

func (p *EvaluationProcessor) storeEvaluationSample(node, state string, sample *bitflow.Sample, header *bitflow.Header) {
	data, ok := p.data[node]
	if !ok {
		data = &nodeEvaluationData{
			name: node,
		}
		p.data[node] = data
	}
	if state == p.NormalStateValue {
		if data.normal == nil {
			data.normal = &bitflow.SampleAndHeader{
				Sample: sample,
				Header: header,
			}
		}
	} else {
		data.anomalies = append(data.anomalies, &EvaluationEvent{
			node:  node,
			state: state,
			sample: &bitflow.SampleAndHeader{
				Sample: sample,
				Header: header,
			},
		})
	}
}

func (p *EvaluationProcessor) assignExpectedRecoveries() (map[string]string, int) {
	allStates := make(map[string]bool)
	for _, node := range p.data {
		for _, anomaly := range node.anomalies {
			allStates[anomaly.state] = true
		}
	}
	numStates := len(allStates)
	numRecoveries := int(p.RecoveriesPerState * float64(numStates))
	p.Execution.SetNumRecoveries(numRecoveries)
	allRecoveries := p.Execution.PossibleRecoveries("some-node") // TODO different nodes might have different stateRecoveries
	allRecoveries = allRecoveries[:numRecoveries]

	// TODO allow different recoveries for different node layers/groups. Requires access to similarity or dependency model
	stateRecoveries := make(map[string]string)
	for _, node := range p.data {
		for _, anomaly := range node.anomalies {
			state := anomaly.state
			recovery, ok := stateRecoveries[state]
			if !ok {
				recovery = allRecoveries[len(stateRecoveries)%len(allRecoveries)]
			}
			anomaly.expectedRecovery = recovery
		}
	}
	return stateRecoveries, numRecoveries
}

func (p *EvaluationProcessor) Close() {
	p.now = time.Now()
	p.runEvaluation()
	p.NoopProcessor.Close()
}

func (p *EvaluationProcessor) runEvaluation() {
	states, numRecoveries := p.assignExpectedRecoveries()

	log.Println("Running evaluation of %v total states and %v total recoveries:", len(states), numRecoveries)
	for state, recovery := range states {
		log.Println(" - %v recovered by %v", state, recovery)
	}

	for nodeName, node := range p.data {
		if node.normal == nil {
			log.Errorf("Cannot evaluate node %v: no normal data sample available", nodeName)
			continue
		}
		if len(node.anomalies) == 0 {
			log.Errorf("Cannot evaluate node %v: no anomaly data available", nodeName)
			continue
		}

		for i, anomaly := range node.anomalies {
			log.Printf("Evaluating node %v event %v of %v (state %v)...", nodeName, i+1, len(node.anomalies), anomaly.state)
			p.currentAnomaly = anomaly
			i := 0
			for !anomaly.resolved {
				p.sendSample(anomaly.sample, node)
				// TODO add timeout?
				i++
				if i%50 == 0 {
					log.Warnf("Sending anomaly sample %v for node %v anomaly %v", i, nodeName, anomaly.state)
				}
			}
			for i := 0; i < p.NormalSamplesBetweenAnomalies; i++ {
				p.sendSample(node.normal, node)
			}
		}
	}

	// TODO output results
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
	for name, data := range p.data {
		if data != node {
			err := p.NoopProcessor.Sample(sample.Sample, sample.Header)
			if err != nil {
				log.Errorf("DecisionMaker evaluation: error sending normal-behavior sample for node %v: %v", name, err)
				return
			}
		}
	}

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

func (p *EvaluationProcessor) progressTime() time.Time {
	res := p.now
	p.now = res.Add(p.SampleRate)
	return res
}

func (p *EvaluationProcessor) executionEventCallback(node string, recovery string, success bool, duration time.Duration) {
	p.currentAnomaly.history = append(p.currentAnomaly.history, RecoveryEvent{
		recovery: recovery,
		duration: duration,
		success:  success,
	})
	if success && p.currentAnomaly.expectedRecovery == recovery {
		p.currentAnomaly.resolved = true
	}
}

type EvaluationEvent struct {
	// For now, store only one state per anomaly. TODO allow complete time series
	sample           *bitflow.SampleAndHeader
	node             string
	state            string
	expectedRecovery string

	resolved bool
	history  []RecoveryEvent
}

type RecoveryEvent struct {
	recovery string
	success  bool
	duration time.Duration
}
