package evaluation

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	onlinestats "github.com/antongulenko/go-onlinestats"
	"github.com/antongulenko/golib"
)

const EventEvaluationTsvHeader = BinaryEvaluationTsvHeader + "\tAnomalies\tDetected\tAvg Detection Time\tFalse Alarms\tAvg False Alarm Duration"

type EventEvaluationProcessor struct {
	GroupedEvaluation

	stateStart     time.Time
	anomalyState   bool
	stateCounter   int // Number of times we switched between anomaly and normal
	lastSampleTime time.Time
}

func (p *EventEvaluationProcessor) String() string {
	return "event-based evaluation"
}

func (p *EventEvaluationProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	isAnomaly := sample.Tag(EvalExpectedTag) == EvalAnomaly
	if p.anomalyState != isAnomaly || p.stateStart.IsZero() {
		p.flushGroups(sample.Time)
		p.stateStart = sample.Time
		p.anomalyState = isAnomaly
		p.stateCounter++
	}
	p.lastSampleTime = sample.Time
	return p.GroupedEvaluation.Sample(sample, header)
}

func (p *EventEvaluationProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
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
	p.flushGroups(p.lastSampleTime)
	p.GroupedEvaluation.Close()
}

func (p *EventEvaluationProcessor) newGroup(groupName string) EvaluationStats {
	return &EventEvaluationStats{
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
}

func (s *EventEvaluationStats) TSV() string {
	str := s.BinaryEvaluationStats.TSV()
	detected := s.DetectedAnomalies.Len()
	detectedStr := "-"
	if s.AnomalyEvents > 0 {
		detectedStr = strconv.Itoa(detected)
	}
	falseAlarmDuration := time.Duration(s.FalseAlarms.Mean())
	falseAlarmDurationStddev := time.Duration(s.FalseAlarms.Stddev())
	detectionTime := time.Duration(s.DetectedAnomalies.Mean())
	detectionTimeStddev := time.Duration(s.DetectedAnomalies.Stddev())
	detectionRate := 0.0
	if s.AnomalyEvents > 0 {
		detectionRate = float64(detected) / float64(s.AnomalyEvents) * 100
	}
	str += fmt.Sprintf("\t%v\t%v (%.1f%%)\t%v ±%v\t%v\t%v ±%v",
		s.AnomalyEvents, detectedStr, detectionRate, detectionTime, detectionTimeStddev, s.FalseAlarms.Len(), falseAlarmDuration, falseAlarmDurationStddev)
	return str
}

func (s *EventEvaluationStats) Evaluate(sample *bitflow.Sample, header *bitflow.Header) {
	s.BinaryEvaluationStats.Evaluate(sample, header)

	predicted := sample.Tag(EvalPredictedTag) == EvalAnomaly
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
	if !s.falseAlarmStart.IsZero() && !t.IsZero() {
		falseAlarmDuration := t.Sub(s.falseAlarmStart)
		s.FalseAlarms.Push(float64(falseAlarmDuration))
	}
	s.falseAlarmStart = time.Time{}
}
