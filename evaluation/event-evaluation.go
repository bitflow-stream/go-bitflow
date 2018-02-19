package evaluation

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

const EventEvaluationTsvHeader = BinaryEvaluationTsvHeader + "\tAnomalies\tDetected\tAvg Detection Time\tFalse Alarms\tAvg False Alarm Duration"

type EventEvaluationProcessor struct {
	GroupedEvaluation

	stateStart   time.Time
	anomalyState bool
	stateCounter int // Number of times we switched between anomaly and normal
}

func (p *EventEvaluationProcessor) String() string {
	return "event-based evaluation"
}

func (p *EventEvaluationProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	isAnomaly := sample.Tag(EvalExpectedTag) == EvalAnomaly
	if p.anomalyState != isAnomaly || p.stateStart.IsZero() {
		p.flushGroups()
		p.stateStart = sample.Time
		p.anomalyState = isAnomaly
		p.stateCounter++
	}
	return p.GroupedEvaluation.Sample(sample, header)
}

func (p *EventEvaluationProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.TsvHeader = EventEvaluationTsvHeader
	p.NewGroup = p.newGroup
	return p.GroupedEvaluation.Start(wg)
}

func (p *EventEvaluationProcessor) flushGroups() {
	for _, group := range p.GroupedEvaluation.groups {
		group.(*EventEvaluationStats).flushState()
	}
}

func (p *EventEvaluationProcessor) Close() {
	p.flushGroups()
	p.GroupedEvaluation.Close()
}

func (p *EventEvaluationProcessor) newGroup(groupName string) EvaluationStats {
	return &EventEvaluationStats{
		processor: p,
	}
}

type EventEvaluationStats struct {
	BinaryEvaluationStats
	AnomalyEvents       int
	UndetectedAnomalies int           // Anomaly events without any detection
	DetectionTime       time.Duration // Time from the beginning of each anomaly to the first detection. Summed up, to make an average
	FalseAlarms         int           // FalsePositive occurrences

	processor                *EventEvaluationProcessor
	stateHandled             bool
	anomalyDetected          bool
	previousAnomalyPredicted bool // To determine FalseAlarms
}

func (s *EventEvaluationStats) TSV() string {
	str := s.BinaryEvaluationStats.TSV()
	detected := s.AnomalyEvents - s.UndetectedAnomalies
	detectedStr := "-"
	if s.AnomalyEvents > 0 {
		detectedStr = strconv.Itoa(detected)
	}
	samplesPerFalseAlarm := "-"
	if s.FalseAlarms > 0 {
		samplesPerFalseAlarm = fmt.Sprintf("%v", float64(s.FP)/float64(s.FalseAlarms))
	}
	timePerDetection := "-"
	if detected > 0 {
		timePerDetection = (s.DetectionTime / time.Duration(detected)).String()
	}
	detectionRate := 0.0
	if s.AnomalyEvents > 0 {
		detectionRate = float64(detected) / float64(s.AnomalyEvents) * 100
	}
	str += fmt.Sprintf("\t%v\t%v (%.1f%%)\t%v\t%v\t%v",
		s.AnomalyEvents, detectedStr, detectionRate, timePerDetection, s.FalseAlarms, samplesPerFalseAlarm)
	return str
}

func (s *EventEvaluationStats) Evaluate(sample *bitflow.Sample, header *bitflow.Header) {
	s.BinaryEvaluationStats.Evaluate(sample, header)

	predicted := sample.Tag(EvalPredictedTag) == EvalAnomaly
	if predicted && s.processor.anomalyState {
		if !s.anomalyDetected {
			s.DetectionTime += sample.Time.Sub(s.processor.stateStart)
		}
		s.anomalyDetected = true
	}
	if !s.processor.anomalyState && predicted && !s.previousAnomalyPredicted {
		s.FalseAlarms++
	}
	s.previousAnomalyPredicted = predicted
	s.stateHandled = true
}

func (s *EventEvaluationStats) flushState() {
	if s.stateHandled && s.processor.anomalyState {
		s.AnomalyEvents++
		if !s.anomalyDetected {
			s.UndetectedAnomalies++
		}
	}
	s.stateHandled = false
	s.anomalyDetected = false
}
