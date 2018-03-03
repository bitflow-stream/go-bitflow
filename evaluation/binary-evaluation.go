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

	BinaryEvaluationTsvHeader = "Total\tTP\tTN\tFP\tFN\tF1\tAccuracy\tPrecision\tSpecificity (TNR)\tRecall (TPR)"
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

func (s *BinaryEvaluationStats) TSV() string {
	tot := s.Total()
	totF := float64(tot) / 100.0
	return fmt.Sprintf("%v\t%v (%.1f%%)\t%v (%.1f%%)\t%v (%.1f%%)\t%v (%.1f%%)\t%.2f%%\t%.2f%%\t%.2f%%\t%.2f%%\t%.2f%%",
		tot, s.TP, float64(s.TP)/totF, s.TN, float64(s.TN)/totF, s.FP, float64(s.FP)/totF, s.FN, float64(s.FN)/totF,
		s.F1()*100, s.Accuracy()*100, s.Precision()*100, s.Specificity()*100, s.Recall()*100)
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

func (s *BinaryEvaluationStats) Specificity() float64 {
	return float64(s.TN) / float64(s.FP+s.TN)
}

func (s *BinaryEvaluationStats) F1() float64 {
	return float64(2*s.TP) / float64(2*s.TP+s.FP+s.FN)
}

func (s *BinaryEvaluationStats) Total() int {
	return s.TP + s.TN + s.FP + s.FN
}
