package recovery

import (
	"fmt"
	"sync"
	"time"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

type State string

const (
	StateUnknown    = State("unknown")
	StateNormal     = State("normal")
	StateAnomaly    = State("anomaly")
	StateRecovering = State("recovering")

	// Indicates that the entire engine is shutting down
	StateShutdown = State("shutdown")
)

type NodeState struct {
	*SimilarityNode
	Name         string
	State        State
	StateString  string
	LastUpdate   time.Time
	LastSample   *bitflow.Sample
	LastHeader   *bitflow.Header
	StateChanged time.Time

	anomaly          *Anomaly
	recoveries       []*Execution
	stateChangedCond *sync.Cond
}

type DecisionMaker struct {
	bitflow.NoopProcessor
	state              map[string]*NodeState
	warnedUnknownNodes map[string]bool

	Graph     *SimilarityGraph
	Execution ExecutionEngine
	History   History
	Selection Selection

	NoDataTimeout         time.Duration
	RecoveryFailedTimeout time.Duration

	ConfigurableTags
}

func RegisterRecoveryEngine(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("recovery", func(pipeline *pipeline.SamplePipeline, params map[string]string) error {
		var err error
		noDataTimeout := query.DurationParam(params, "no-data", 0, false, &err)
		recoveryFailedTimeout := query.DurationParam(params, "recovery-failed", 0, false, &err)
		layerSimilarity := query.FloatParam(params, "layer-simil", 0, false, &err)
		groupSimilarity := query.FloatParam(params, "group-simil", 0, false, &err)
		evaluate := query.BoolParam(params, "evaluate", false, true, &err)
		if err != nil {
			return err
		}

		dependencyModelFile := params["model"]
		dependencyModel, err := LoadDependencyModel(dependencyModelFile)
		if err != nil {
			return query.ParameterError("model", err)
		}
		graph := dependencyModel.BuildSimilarityGraph(groupSimilarity, layerSimilarity)

		execution, err := NewMockExecution(params)
		if err != nil {
			return err
		}

		if evaluate {
			evalStep := &EvaluationProcessor{
				data:      make(map[string]*nodeEvaluationData),
				Execution: execution,
			}
			evalStep.SampleRate = query.DurationParam(params, "sample-rate", 0, true, &err)
			evalStep.FillerSamples = query.IntParam(params, "filler-samples", 0, true, &err)
			evalStep.NormalSamplesBetweenAnomalies = query.IntParam(params, "normal-fillers", 0, true, &err)
			evalStep.RecoveriesPerState = query.FloatParam(params, "recoveries-per-state", 1, true, &err)
			if err != nil {
				return err
			}
			evalStep.ParseRecoveryTags(params)
			pipeline.Add(evalStep)
		}

		history := new(VolatileHistory)
		selection := new(RandomSelection)

		var tags ConfigurableTags
		tags.ParseRecoveryTags(params)
		pipeline.Add(&DecisionMaker{
			Graph:                 graph,
			Execution:             execution,
			History:               history,
			Selection:             selection,
			NoDataTimeout:         noDataTimeout,
			RecoveryFailedTimeout: recoveryFailedTimeout,
			ConfigurableTags:      tags,
		})
		return nil
	}, "Recovery Engine based on recommendation system",
		append([]string{
			"model", "layer-simil", "group-simil", // Dependency/Similarity Graph
			"no-data", "recovery-failed", // Timeouts
		}, TagParameterNames...),
		"avg-recovery-time", "recovery-error-percentage", "num-mock-recoveries", // Mock execution engine
		"evaluate", "sample-rate", "filler-samples", "normal-fillers", "recoveries-per-state", // Evaluation
	)
}

func (d *DecisionMaker) String() string {
	return fmt.Sprintf("Recovery-Engine Decision Maker (node-name: %v, normal-state: %v=%v, no-data-timeout: %v, recovery-failed-timeout: %v)",
		d.NodeNameTag, d.StateTag, d.NormalStateValue, d.NoDataTimeout, d.RecoveryFailedTimeout)
}

func (d *DecisionMaker) Start(wg *sync.WaitGroup) golib.StopChan {
	d.state = make(map[string]*NodeState)
	d.warnedUnknownNodes = make(map[string]bool)
	for nodeName, node := range d.Graph.Nodes {
		state := &NodeState{
			SimilarityNode:   node,
			State:            StateUnknown,
			Name:             nodeName,
			stateChangedCond: sync.NewCond(new(sync.Mutex)),
		}
		wg.Add(1)
		go d.handleStateChangeLoop(wg, state)
		d.state[nodeName] = state
	}
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
			nodeState.LastUpdate = now
			nodeState.StateString = state
			nodeState.LastSample = sample
			nodeState.LastHeader = header
			if state == d.NormalStateValue {
				d.setState(now, nodeState, StateNormal)
			} else {
				if nodeState.State != StateRecovering {
					// Stay in recovering state until timeout
					d.setState(now, nodeState, StateAnomaly)
				}
			}
		}
	}
	d.timeTick(now)
	return d.NoopProcessor.Sample(sample, header)
}

func (d *DecisionMaker) Close() {
	for _, state := range d.state {
		d.setState(state.StateChanged, state, StateShutdown)
	}
}

func (d *DecisionMaker) timeTick(now time.Time) {
	for _, state := range d.state {
		if !state.LastUpdate.IsZero() && !d.hasData(now, state) {
			// Receiving no data is counted as anomaly
			d.setState(now, state, StateAnomaly)
		} else if state.State == StateRecovering && now.Sub(state.StateChanged) >= d.RecoveryFailedTimeout {
			d.setState(now, state, StateAnomaly)
		}
	}
}

func (d *DecisionMaker) hasData(now time.Time, state *NodeState) bool {
	return now.Sub(state.LastUpdate) < d.NoDataTimeout
}

func (d *DecisionMaker) setState(now time.Time, state *NodeState, newState State) {
	if newState != state.State {
		state.State = newState
		state.StateChanged = now
		state.stateChangedCond.L.Lock()
		defer state.stateChangedCond.L.Unlock()
		state.stateChangedCond.Broadcast()
	}
}

func (d *DecisionMaker) handleStateChangeLoop(wg *sync.WaitGroup, node *NodeState) {
	defer wg.Done()
	oldState := node.State
	for {
		// Wait for state change
		for oldState == node.State {
			node.stateChangedCond.L.Lock()
			node.stateChangedCond.Wait()
			node.stateChangedCond.L.Unlock()
		}
		newState := node.State
		if newState == StateShutdown || d.StopChan.Stopped() {
			break
		}

		d.stateChanged(node.StateChanged, oldState, node)
		oldState = newState
	}
}

func (d *DecisionMaker) stateChanged(now time.Time, oldState State, state *NodeState) {
	log.Printf("Node %v switched state %v -> %v (have data: %v)", state.Name, oldState, state.State, d.hasData(now, state))
	switch {
	case oldState == StateRecovering && state.State == StateNormal:
		d.recoverySuccessful(now, state)
	case oldState == StateRecovering && state.State == StateAnomaly:
		d.recoveryTimedOut(now, state)
	case state.State == StateAnomaly:
		if state.anomaly != nil {
			log.Errorf("State of node %v changed to anomaly, although an anomaly was already active... Discarding old anomaly data.", state.Name)
		}
		state.recoveries = nil
		state.anomaly = &Anomaly{
			Node:     state.Name,
			Features: SampleToAnomalyFeatures(state.LastSample, state.LastHeader),
			Start:    now,
		}
		d.runRecovery(now, state)
	}
}

func (d *DecisionMaker) getAnomalyFeatures(now time.Time, state *NodeState) []AnomalyFeature {
	if state.LastHeader == nil || state.LastSample == nil {
		// No data available... Should usually not happen, but the recovery-action selection will have to deal with that.
		log.Warnf("No data received yet for node '%v', cannot compute features of anomaly situation", state.Name)
		return nil
	}
	return SampleToAnomalyFeatures(state.LastSample, state.LastHeader)
}

func (d *DecisionMaker) runRecovery(now time.Time, state *NodeState) {
	for {
		recoveryName := d.selectRecovery(now, state)
		duration, err := d.Execution.RunRecovery(state.Name, recoveryName)
		finished := now.Add(duration)
		recovery := &Execution{
			Node:              state.Name,
			Recovery:          recoveryName,
			Started:           now,
			ExecutionFinished: finished,
		}
		state.recoveries = append(state.recoveries, recovery)
		if err != nil {
			recovery.Error = err.Error()
			now = finished
			d.runRecovery(finished, state)
		} else {
			// TODO limit the number of retries
			break
		}
	}
}

func (d *DecisionMaker) recoverySuccessful(now time.Time, state *NodeState) {
	recovery := state.recoveries[len(state.recoveries)-1]
	recovery.Ended = now
	recovery.Successful = true
	d.History.StoreAnomaly(state.anomaly, state.recoveries)
	state.anomaly = nil
	state.recoveries = nil
}

func (d *DecisionMaker) recoveryTimedOut(now time.Time, state *NodeState) {
	recovery := state.recoveries[len(state.recoveries)-1]
	recovery.Ended = now
	d.runRecovery(now, state)
}

func (d *DecisionMaker) selectRecovery(now time.Time, state *NodeState) string {
	possible := d.Execution.PossibleRecoveries(state.Name)
	return d.Selection.SelectRecovery(state.SimilarityNode, state.anomaly.Features, possible, d.History)
}
