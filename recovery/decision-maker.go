package recovery

import (
	"fmt"
	"sync"
	"time"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/bitflow-script/reg"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

type State string

const (
	StateUnknown    = State("unknown")
	StateNormal     = State("normal")
	StateAnomaly    = State("anomaly")
	StateNoData     = State("no-data")
	StateRecovering = State("recovering")
)

type DecisionMaker struct {
	bitflow.NoopProcessor
	state              map[string]*NodeState
	warnedUnknownNodes map[string]bool

	Graph     *SimilarityGraph
	Execution ExecutionEngine
	History   History
	Selection Selection

	RecoverNoDataState    bool
	NoDataTimeout         time.Duration
	RecoveryFailedTimeout time.Duration

	NodeStateChangeCallback func(nodeName string, oldState, newState State, timestamp time.Time)

	ConfigurableTags

	now          time.Time
	shutdown     bool
	progressCond *sync.Cond
}

func RegisterRecoveryEngine(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("recovery", func(p *pipeline.SamplePipeline, params map[string]string) error {
		var err error

		noDataTimeout := reg.DurationParam(params, "no-data", 0, false, &err)
		recoveryFailedTimeout := reg.DurationParam(params, "recovery-failed", 0, false, &err)
		layerSimilarity := reg.FloatParam(params, "layer-simil", 0, false, &err)
		groupSimilarity := reg.FloatParam(params, "group-simil", 0, false, &err)
		evaluate := reg.BoolParam(params, "evaluate", false, true, &err)
		recoverNoDataState := reg.BoolParam(params, "recover-no-data", false, false, &err)
		randomSelection := reg.BoolParam(params, "random-selection", false, true, &err)
		if err != nil {
			return err
		}

		dependencyModelFile := params["model"]
		dependencyModel, err := LoadDependencyModel(dependencyModelFile)
		if err != nil {
			return reg.ParameterError("model", err)
		}
		graph := dependencyModel.BuildSimilarityGraph(groupSimilarity, layerSimilarity)

		execution, err := NewMockExecution(params)
		if err != nil {
			return err
		}
		var selection Selection
		if randomSelection {
			selection, err = NewRandomSelectionParams(params)
		} else {
			selection = new(RecommendingSelection)
		}
		if err != nil {
			return err
		}

		history := new(VolatileHistory)

		var tags ConfigurableTags
		tags.ParseRecoveryTags(params)

		engine := &DecisionMaker{
			Graph:                 graph,
			Execution:             execution,
			History:               history,
			Selection:             selection,
			NoDataTimeout:         noDataTimeout,
			RecoveryFailedTimeout: recoveryFailedTimeout,
			ConfigurableTags:      tags,
			RecoverNoDataState:    recoverNoDataState,
		}

		if evaluate {
			collector := &EvaluationDataCollector{
				data:               make(map[string]*nodeEvaluationData),
				StoreNormalSamples: reg.IntParam(params, "store-normal-samples", 1000, true, &err),
				ConfigurableTags:   tags,
			}
			evalStep := &EvaluationProcessor{
				Execution:           execution,
				SampleRate:          reg.DurationParam(params, "sample-rate", 0, true, &err),
				FillerSamples:       reg.IntParam(params, "filler-samples", 0, true, &err),
				RecoveriesPerState:  reg.FloatParam(params, "recoveries-per-state", 1, true, &err),
				MaxRecoveryAttempts: reg.IntParam(params, "max-recoveries", 500, true, &err),
				collector:           collector,
			}
			if err != nil {
				return err
			}

			engine.NodeStateChangeCallback = evalStep.nodeStateChangeCallback
			p.Add((&pipeline.BatchProcessor{
				FlushTags:                   []string{collector.NodeNameTag, collector.StateTag, "anomaly"},
				SampleTimestampFlushTimeout: 5 * time.Second,
			}).Add(collector))
			p.Add(evalStep)
		}

		p.Add(engine)
		return nil
	}, "Recovery Engine based on recommendation system",
		reg.RequiredParams(append([]string{
			"model", "layer-simil", "group-simil", // Dependency/Similarity Graph
			"no-data", "recovery-failed", // Timeouts
			"recover-no-data",
		}, TagParameterNames...)...),
		reg.OptionalParams(
			"avg-recovery-time", "recovery-error-percentage", "num-mock-recoveries", "rand-seed", // Mock execution engine
			"evaluate", "max-recoveries", "sample-rate", "filler-samples", "recoveries-per-state", "store-normal-samples", // Evaluation
			"random-selection", // Selection
		))

}

func (d *DecisionMaker) String() string {
	return fmt.Sprintf("Recovery-Engine Decision Maker (%v, no-data-timeout: %v, recovery-failed-timeout: %v, recover-no-data: %v)",
		d.ConfigurableTags, d.NoDataTimeout, d.RecoveryFailedTimeout, d.RecoverNoDataState)
}

func (d *DecisionMaker) Start(wg *sync.WaitGroup) golib.StopChan {
	d.state = make(map[string]*NodeState)
	d.warnedUnknownNodes = make(map[string]bool)
	d.progressCond = sync.NewCond(new(sync.Mutex))
	for nodeName, graphNode := range d.Graph.Nodes {
		d.state[nodeName] = &NodeState{
			engine:         d,
			SimilarityNode: graphNode,
			LastState:      StateUnknown,
			state:          StateUnknown,
			Name:           nodeName,
		}
	}
	wg.Add(1)
	go d.loopHandleUpdates(wg)
	return d.NoopProcessor.Start(wg)
}

func (d *DecisionMaker) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	now := sample.Time // Use the sample time to make evaluation experiments easier to control
	node, state := d.GetRecoveryTags(sample)
	if node != "" && state != "" {
		nodeState, ok := d.state[node]
		if !ok {
			if !d.warnedUnknownNodes[node] {
				log.Warnf("Ignoring data for unknown node with %v=%v and %v=%v", d.NodeNameTag, node, d.StateTag, state)
				d.warnedUnknownNodes[node] = true
			}
		} else {
			nodeState.LastSample = sample
			nodeState.LastHeader = header
			nodeState.LastUpdate = now
			if state == d.NormalStateValue {
				nodeState.LastState = StateNormal
			} else {
				nodeState.LastState = StateAnomaly
			}
		}
	}
	d.progressTime(now)
	return d.NoopProcessor.Sample(sample, header)
}

func (d *DecisionMaker) Close() {
	d.shutdown = true
	d.progressTime(d.now) // Wakeup and shutdown all parallel goroutines
	d.NoopProcessor.Close()
}

func (d *DecisionMaker) progressTime(now time.Time) {
	d.progressCond.L.Lock()
	defer d.progressCond.L.Unlock()
	d.now = now

	for _, node := range d.state {
		if now.Sub(node.LastUpdate) > node.engine.NoDataTimeout {
			// Not receiving any data from node
			node.LastState = StateNoData
		}
	}
	d.progressCond.Broadcast()
}

func (d *DecisionMaker) loopHandleUpdates(wg *sync.WaitGroup) {
	defer wg.Done()
	var now time.Time
	for !d.shutdown {
		now = d.waitForUpdate(now)
		if d.shutdown {
			break
		}
		for _, node := range d.state {
			node.stateUpdated(now)
		}
	}
}

func (d *DecisionMaker) waitForUpdate(previousTime time.Time) time.Time {
	d.progressCond.L.Lock()
	defer d.progressCond.L.Unlock()
	for !previousTime.Before(d.now) && !d.shutdown {
		d.progressCond.Wait()
	}
	return d.now
}

type NodeState struct {
	*SimilarityNode
	engine *DecisionMaker
	Name   string

	// Data updated from received samples
	LastUpdate time.Time
	LastSample *bitflow.Sample
	LastHeader *bitflow.Header
	LastState  State

	// Internal state
	state        State
	stateChanged time.Time

	anomaly    *AnomalyEvent
	recoveries []*ExecutionEvent
}

func (node *NodeState) stateUpdated(now time.Time) {
	newState := node.LastState
	if node.state == StateRecovering && newState != StateNormal {
		// When recovering, stay in that state until the anomaly is resolved or times out
		if now.Sub(node.stateChanged) < node.engine.RecoveryFailedTimeout {
			newState = StateRecovering // Recovery has not yet timed out
		}
	}
	node.setState(newState, now)
}

func (node *NodeState) setState(newState State, now time.Time) {
	if node.state != newState {
		oldState := node.state
		node.state = newState
		node.stateChanged = now
		log.Debugf("Node %v switched state %v -> %v", node.Name, oldState, newState)
		node.handleStateChanged(oldState, now)
		if callback := node.engine.NodeStateChangeCallback; callback != nil {
			callback(node.Name, oldState, newState, now)
		}
	}
}

func (node *NodeState) handleStateChanged(oldState State, now time.Time) {
	newState := node.state
	switch {
	case oldState == StateRecovering && newState == StateNormal:
		// Recovery successful
		recovery := node.recoveries[len(node.recoveries)-1]
		recovery.Ended = now
		recovery.Successful = true
		node.anomaly.End = now
		node.engine.History.StoreAnomaly(node.anomaly, node.recoveries)
		node.anomaly = nil
		node.recoveries = nil
	case oldState == StateRecovering && (newState == StateAnomaly || newState == StateNoData):
		// Recovery timed out. Restart recovery procedure.
		recovery := node.recoveries[len(node.recoveries)-1]
		recovery.Ended = now
		fallthrough
	case newState == StateAnomaly || newState == StateNoData:
		if newState == StateAnomaly || node.engine.RecoverNoDataState {
			if node.anomaly == nil {
				node.anomaly = &AnomalyEvent{
					Node:     node.Name,
					Features: SampleToAnomalyFeatures(node.LastSample, node.LastHeader),
					Start:    now,
				}
			}
			node.setState(StateRecovering, now)
		}
	case newState == StateRecovering:
		node.runRecovery(now)
	}
}

func (node *NodeState) runRecovery(now time.Time) {
	recoveryName := node.selectRecovery()
	if recoveryName == "" {
		log.Errorf("No recovery available for node %v, state %v", node.Name, node.state)
		return
	}
	duration, err := node.engine.Execution.RunRecovery(node.Name, recoveryName)
	node.recoveries = append(node.recoveries, &ExecutionEvent{
		Node:     node.Name,
		Recovery: recoveryName,
		Started:  now,
		Duration: duration,
		Error:    err,
	})
}

func (node *NodeState) selectRecovery() string {
	possible := node.engine.Execution.PossibleRecoveries(node.Name)
	if len(possible) == 0 {
		return ""
	}
	return node.engine.Selection.SelectRecovery(node.SimilarityNode, node.anomaly.Features, possible, node.engine.History)
}
