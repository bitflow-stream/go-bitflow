package evaluation

import (
	"fmt"
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/golib"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
)

const (
	BinaryEvaluationTsvHeader = "Total\tTP\tTN\tFP\tFN\tF1\tAccuracy\tPrecision\tSpecificity (TNR)\tRecall (TPR)"
)

type BinaryEvaluationProcessor struct {
	GroupedEvaluation
	BinaryEvaluationTags
}

func (p *BinaryEvaluationProcessor) String() string {
	return fmt.Sprintf("binary classification evaluation (evaluation: [%v], binary evaluation: [%v])",
		&p.ConfigurableTags, &p.BinaryEvaluationTags)
}

func (p *BinaryEvaluationProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	p.TsvHeader = BinaryEvaluationTsvHeader
	p.NewGroup = p.newGroup
	return p.GroupedEvaluation.Start(wg)
}

func (p *BinaryEvaluationProcessor) newGroup(groupName string) Stats {
	return &BinaryEvaluationStats{
		Tags: &p.BinaryEvaluationTags,
	}
}

type BinaryEvaluationTags struct {
	Expected     string // "expected"
	Predicted    string // "predicted"
	AnomalyValue string // "anomaly", All other values (or missing values) considered not anomaly
}

func RegisterBinaryEvaluation(b builder.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) {
		eval := new(BinaryEvaluationProcessor)
		eval.SetBinaryEvaluationTags(params)
		eval.SetEvaluationTags(params)
		p.Add(eval)
	}
	b.RegisterAnalysisParams("binary_evaluation", create, "Evaluate 'expected' and 'predicted' tags, separate evaluation by |-separated fields in 'evalGroups' tag", builder.OptionalParams("expectedTag", "predictedTag", "anomalyValue", "evaluateTag", "evaluateValue", "groupsTag", "groupsSeparator"))
}

func (e *BinaryEvaluationTags) SetBinaryEvaluationTags(params map[string]string) {
	e.Expected = params["expectedTag"]
	if e.Expected == "" {
		e.Expected = "expected"
	}
	e.Predicted = params["predictedTag"]
	if e.Predicted == "" {
		e.Predicted = "predicted"
	}
	e.AnomalyValue = params["anomalyValue"]
	if e.AnomalyValue == "" {
		e.AnomalyValue = "anomaly"
	}
}

func (e *BinaryEvaluationTags) String() string {
	return fmt.Sprintf("expected tag: \"%v\", predicted tag: \"%v\", anomaly value: \"%v\"",
		e.Expected, e.Predicted, e.AnomalyValue)
}

type BinaryEvaluationStats struct {
	Tags *BinaryEvaluationTags
	TP   int
	TN   int
	FP   int
	FN   int
}

func (s *BinaryEvaluationStats) TSV() string {
	tot := s.Total()
	totF := float64(tot) / 100.0
	return fmt.Sprintf("%v\t%v (%.1f%%)\t%v (%.1f%%)\t%v (%.1f%%)\t%v (%.1f%%)\t%.2f%%\t%.2f%%\t%.2f%%\t%.2f%%\t%.2f%%",
		tot, s.TP, float64(s.TP)/totF, s.TN, float64(s.TN)/totF, s.FP, float64(s.FP)/totF, s.FN, float64(s.FN)/totF,
		s.F1()*100, s.Accuracy()*100, s.Precision()*100, s.Specificity()*100, s.Recall()*100)
}

func (s *BinaryEvaluationStats) Evaluate(sample *bitflow.Sample, header *bitflow.Header) {
	expected := sample.Tag(s.Tags.Expected) == s.Tags.AnomalyValue
	predicted := sample.Tag(s.Tags.Predicted) == s.Tags.AnomalyValue
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
