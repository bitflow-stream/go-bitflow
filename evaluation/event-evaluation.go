package evaluation

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	onlinestats "github.com/antongulenko/go-onlinestats"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

const EventEvaluationTsvHeader = BinaryEvaluationTsvHeader + "\tAnomalies\tDetected\tAvg Detection Time\tFalse Alarms\tAvg False Alarm Duration"

type AbstractAnomalyEventProcessor struct {
	BinaryEvaluationTags
	BatchKeyTag string

	StateChanged func(time.Time)

	previousKey    string
	stateStart     time.Time
	anomalyState   bool
	stateCounter   int // Number of times we switched between anomaly and normal
	lastSampleTime time.Time
}

func (p *AbstractAnomalyEventProcessor) Sample(sample *bitflow.Sample) {
	isNewBatch := false
	flushTime := sample.Time
	if p.BatchKeyTag != "" {
		newKey := sample.Tag(p.BatchKeyTag)
		isNewBatch = p.previousKey != newKey
		if isNewBatch {
			log.Debugf("Flushing event evaluation batch because \"%v\" tag changed from \"%v\" to \"%v\"", p.BatchKeyTag, p.previousKey, newKey)
		}
		p.previousKey = newKey
		if isNewBatch {
			flushTime = time.Time{} // The new sample is not related to the previous ones
		}
	}
	isAnomaly := sample.Tag(p.Expected) == p.AnomalyValue
	if isNewBatch || p.stateStart.IsZero() || p.anomalyState != isAnomaly {
		p.StateChanged(flushTime)
		p.stateStart = sample.Time
		p.anomalyState = isAnomaly
		p.stateCounter++
	}
}

func (p *AbstractAnomalyEventProcessor) Close() {
	p.StateChanged(time.Time{})
}

type EventEvaluationProcessor struct {
	GroupedEvaluation
	AbstractAnomalyEventProcessor
}

func (p *EventEvaluationProcessor) String() string {
	return fmt.Sprintf("event-based evaluation (batch-key-tag: \"%v\", evaluation: [%v], binary evaluation: [%v])",
		p.BatchKeyTag, &p.EvaluationTags, &p.BinaryEvaluationTags)
}

func (p *EventEvaluationProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.AbstractAnomalyEventProcessor.Sample(sample)
	return p.GroupedEvaluation.Sample(sample, header)
}

func (p *EventEvaluationProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.StateChanged = p.flushGroups
	p.TsvHeader = EventEvaluationTsvHeader
	p.NewGroup = p.newGroup
	return p.GroupedEvaluation.Start(wg)
}

func (p *EventEvaluationProcessor) flushGroups(t time.Time) {
	for _, group := range p.GroupedEvaluation.groups {
		group.(*EventEvaluationStats).flushState(t)
	}
}

func (p *EventEvaluationProcessor) Close() {
	p.AbstractAnomalyEventProcessor.Close()
	p.GroupedEvaluation.Close()
}

func (p *EventEvaluationProcessor) newGroup(groupName string) EvaluationStats {
	return &EventEvaluationStats{
		BinaryEvaluationStats: BinaryEvaluationStats{
			Tags: &p.BinaryEvaluationTags,
		},
		processor: p,
	}
}

type EventEvaluationStats struct {
	BinaryEvaluationStats
	AnomalyEvents     int
	FalseAlarms       onlinestats.Running // Count and duration of false alarms
	DetectedAnomalies onlinestats.Running // Count of detected anomalies, duration until detection

	processor       *EventEvaluationProcessor
	stateHandled    bool
	anomalyDetected bool
	falseAlarmStart time.Time
	lastSampleTime  time.Time
}

func (s *EventEvaluationStats) TSV() string {
	str := s.BinaryEvaluationStats.TSV()

	detected := s.DetectedAnomalies.Len()
	detectedStr := "-"
	if s.AnomalyEvents > 0 {
		detectedStr = strconv.Itoa(detected)
	}
	detectionRate := 0.0
	if s.AnomalyEvents > 0 {
		detectionRate = float64(detected) / float64(s.AnomalyEvents) * 100
	}

	falseAlarmDuration := time.Duration(s.FalseAlarms.Mean()).String()
	if s.FalseAlarms.Len() > 1 {
		falseAlarmDuration += " ±" + time.Duration(s.FalseAlarms.Stddev()).String()
	}
	detectionTime := time.Duration(s.DetectedAnomalies.Mean()).String()
	if s.DetectedAnomalies.Len() > 1 {
		detectionTime += " ±" + time.Duration(s.DetectedAnomalies.Stddev()).String()
	}

	str += fmt.Sprintf("\t%v\t%v (%.1f%%)\t%v\t%v\t%v",
		s.AnomalyEvents, detectedStr, detectionRate, detectionTime, s.FalseAlarms.Len(), falseAlarmDuration)
	return str
}

func (s *EventEvaluationStats) Evaluate(sample *bitflow.Sample, header *bitflow.Header) {
	s.BinaryEvaluationStats.Evaluate(sample, header)
	s.lastSampleTime = sample.Time

	predicted := sample.Tag(s.Tags.Predicted) == s.Tags.AnomalyValue
	if predicted && s.processor.anomalyState {
		if !s.anomalyDetected {
			detectionTime := sample.Time.Sub(s.processor.stateStart)
			s.DetectedAnomalies.Push(float64(detectionTime))
		}
		s.anomalyDetected = true
	}
	isFalseAlarm := predicted && !s.processor.anomalyState
	falseAlarmRunning := !s.falseAlarmStart.IsZero()
	if isFalseAlarm && !falseAlarmRunning {
		s.falseAlarmStart = sample.Time
	} else if !isFalseAlarm && falseAlarmRunning {
		s.flushFalseAlarm(sample.Time)
	}

	if !s.stateHandled && s.processor.anomalyState {
		s.AnomalyEvents++
	}
	s.stateHandled = true
}

func (s *EventEvaluationStats) flushState(t time.Time) {
	s.flushFalseAlarm(t)
	s.stateHandled = false
	s.anomalyDetected = false
}

func (s *EventEvaluationStats) flushFalseAlarm(t time.Time) {
	if t.IsZero() {
		t = s.lastSampleTime
	}
	if !s.falseAlarmStart.IsZero() && !t.IsZero() {
		if t.Before(s.falseAlarmStart) {
			log.Warnf("Ignoring false alarm that ends before it starts (Start: %v, End: %v)", s.falseAlarmStart, t)
		} else {
			falseAlarmDuration := t.Sub(s.falseAlarmStart)
			s.FalseAlarms.Push(float64(falseAlarmDuration))
		}
	}
	s.falseAlarmStart = time.Time{}
}

type AnomalySmoothing struct {
	bitflow.AbstractProcessor
	BinaryEvaluationTags
	AbstractAnomalyEventProcessor
	NormalTagValue    string
	SmoothingDuration time.Duration

	stateChanged         time.Time
	anomalyRunning       bool
	outputAnomalyRunning bool
	lastHost             string
}

func (p *AnomalySmoothing) String() string {
	return fmt.Sprintf("anomaly smoothing (smoothing duration: %v, normal tag value: %v, batch key: %v, tags: [%v])", p.SmoothingDuration, p.NormalTagValue, p.BatchKeyTag, &p.BinaryEvaluationTags)
}

func (p *AnomalySmoothing) Start(wg *sync.WaitGroup) golib.StopChan {
	p.StateChanged = p.anomalyStateChanged
	return p.AbstractProcessor.Start(wg)
}

func (p *AnomalySmoothing) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.AbstractAnomalyEventProcessor.Sample(sample)

	isAnomaly := sample.Tag(p.Predicted) == p.AnomalyValue
	if p.stateChanged.IsZero() || p.anomalyRunning != isAnomaly {
		p.stateChanged = sample.Time
		p.anomalyRunning = isAnomaly
	}
	stateDuration := sample.Time.Sub(p.stateChanged)
	if stateDuration > p.SmoothingDuration {
		p.outputAnomalyRunning = isAnomaly
	}

	if p.outputAnomalyRunning {
		sample.SetTag(p.Predicted, p.AnomalyValue)
	} else {
		sample.SetTag(p.Predicted, p.NormalTagValue)
	}
	return p.AbstractProcessor.Sample(sample, header)
}

func (p *AnomalySmoothing) Close() {
	p.AbstractAnomalyEventProcessor.Close()
	p.AbstractProcessor.Close()
}

func (p *AnomalySmoothing) anomalyStateChanged(t time.Time) {
	p.stateChanged = time.Time{}
}
