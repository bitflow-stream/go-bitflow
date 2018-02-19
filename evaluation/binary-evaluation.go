package evaluation

import (
	"fmt"
	"sync"

	bitflow "github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

const (
	EvalExpectedTag  = "evalExpected"
	EvalPredictedTag = "evalPredicted"
	EvalNormal       = "normal"
	EvalAnomaly      = "anomaly"

	BinaryEvaluationTsvHeader = "F1\tAccuracy\tPrecision\tRecall\tTotal\tTP\tTN\tFP\tFN"
)

type BinaryEvaluationProcessor struct {
	GroupedEvaluation
}

func (p *BinaryEvaluationProcessor) String() string {
	return "binary classification evaluation"
}

func (p *BinaryEvaluationProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.TsvHeader = BinaryEvaluationTsvHeader
	p.NewGroup = p.newGroup
	return p.GroupedEvaluation.Start(wg)
}

func (p *BinaryEvaluationProcessor) newGroup(groupName string) EvaluationStats {
	return new(BinaryEvaluationStats)
}

type BinaryEvaluationStats struct {
	TP int
	TN int
	FP int
	FN int
}

func (s *BinaryEvaluationStats) String() string {
	return fmt.Sprintf("F1 = %.2v, Accuracy = %.2v, Precision = %.2v, Recall = %.2v, Total = %v (TP = %v, TN = %v, FP = %v, FN = %v)",
		s.F1(), s.Accuracy(), s.Precision(), s.Recall(), s.Total(), s.TP, s.TN, s.FP, s.FN)
}

func (s *BinaryEvaluationStats) TSV() string {
	tot := s.Total()
	totF := float64(tot) / 100.0
	return fmt.Sprintf("%.2f%%\t%.2f%%\t%.2f%%\t%.2f%%\t%v\t%v (%.1f%%)\t%v (%.1f%%)\t%v (%.1f%%)\t%v (%.1f%%)",
		s.F1()*100, s.Accuracy()*100, s.Precision()*100, s.Recall()*100,
		tot, s.TP, float64(s.TP)/totF, s.TN, float64(s.TN)/totF, s.FP, float64(s.FP)/totF, s.FN, float64(s.FN)/totF)
}

func (s *BinaryEvaluationStats) Evaluate(sample *bitflow.Sample, header *bitflow.Header) {
	expected := sample.Tag(EvalExpectedTag) == EvalAnomaly
	predicted := sample.Tag(EvalPredictedTag) == EvalAnomaly
	s.EvaluateClassification(expected, predicted)
}

func (s *BinaryEvaluationStats) EvaluateClassification(expected, predicted bool) {
	switch {
	case expected && predicted:
		s.TP++
	case !expected && !predicted:
		s.TN++
	case expected && !predicted:
		s.FN++
	case !expected && predicted:
		s.FP++
	}
}

func (s *BinaryEvaluationStats) Accuracy() float64 {
	return float64(s.TP+s.TN) / float64(s.Total())
}

func (s *BinaryEvaluationStats) Precision() float64 {
	return float64(s.TP) / float64(s.TP+s.FP)
}

func (s *BinaryEvaluationStats) Recall() float64 {
	return float64(s.TP) / float64(s.TP+s.FN)
}

func (s *BinaryEvaluationStats) F1() float64 {
	return float64(2*s.TP) / float64(2*s.TP+s.FP+s.FN)
}

func (s *BinaryEvaluationStats) Total() int {
	return s.TP + s.TN + s.FP + s.FN
}
