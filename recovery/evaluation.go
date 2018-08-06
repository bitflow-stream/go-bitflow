package recovery

import (
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	log "github.com/sirupsen/logrus"
)

var (
	evaluationFillerHeader = &bitflow.Header{Fields: []string{}} // Empty header for samples to progress the time in the DecisionMaker
)

type RecoveryEvaluationProcessor struct {
	bitflow.NoopProcessor
	RecoveryTags

	NormalSamplesBetweenAnomalies int
	FillerSamples                 int           // Number of samples to send between two real evaluation samples
	SampleRate                    time.Duration // Time progression between samples (both real and filler samples)

	data map[string]*nodeEvaluationData // Key: node name
	now  time.Time
}

type nodeEvaluationData struct {
	name      string
	anomalies map[string][]*bitflow.SampleAndHeader
	normal    *bitflow.SampleAndHeader
}

func RegisterRecoveryEvaluation(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("recovery-evaluation", func(p *pipeline.SamplePipeline, params map[string]string) error {
		step := &RecoveryEvaluationProcessor{
			data: make(map[string]*nodeEvaluationData),
		}
		var err error
		step.SampleRate = query.DurationParam(params, "sample-rate", 0, false, &err)
		step.FillerSamples = query.IntParam(params, "filler-samples", 0, true, &err)
		step.NormalSamplesBetweenAnomalies = query.IntParam(params, "normal-fillers", 0, true, &err)
		if err != nil {
			return err
		}

		step.ParseRecoveryTags(params)
		p.Add(step)
		return nil
	}, "Evaluation procedure for the recovery() step. Fully reads all input data before performing the evaluation.",
		append([]string{
			"sample-rate",
		}, RecoveryTagParams...),
		"filler-samples", "normal-fillers")
}

func (p *RecoveryEvaluationProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	node, state := p.GetRecoveryTags(sample)
	if node != "" && state != "" {
		p.storeEvaluationSample(node, state, sample, header)
	}
	return p.NoopProcessor.Sample(sample, header)
}

func (p *RecoveryEvaluationProcessor) storeEvaluationSample(node, state string, sample *bitflow.Sample, header *bitflow.Header) {
	data, ok := p.data[node]
	if !ok {
		data = &nodeEvaluationData{
			name:      node,
			anomalies: make(map[string][]*bitflow.SampleAndHeader),
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
		data.anomalies[state] = append(data.anomalies[state], &bitflow.SampleAndHeader{
			Sample: sample,
			Header: header,
		})
	}
}

func (p *RecoveryEvaluationProcessor) Close() {
	p.now = time.Now()
	p.runEvaluation()
	p.NoopProcessor.Close()
}

func (p *RecoveryEvaluationProcessor) runEvaluation() {
	for node, data := range p.data {
		if data.normal == nil {
			log.Errorf("Cannot evaluate node %v: no normal data sample available", node)
			continue
		}
		if len(data.anomalies) == 0 {
			log.Errorf("Cannot evaluate node %v: no anomaly data available", node)
			continue
		}

		for anomalyName, anomaly := range data.anomalies {
			log.Printf("Evaluating node %v anomaly %v (%v sample(s))...", node, anomalyName, len(anomaly))
			for _, sample := range anomaly {
				p.sendSample(sample)
				// TODO continue sending this sample until the anomaly is resolved
			}
			for i := 0; i < p.NormalSamplesBetweenAnomalies; i++ {
				p.sendSample(data.normal)
			}
		}
	}
}

func (p *RecoveryEvaluationProcessor) sendSample(sample *bitflow.SampleAndHeader) {
	sample.Sample.Time = p.progressTime()
	err := p.NoopProcessor.Sample(sample.Sample, sample.Header)
	if err != nil {
		log.Errorf("DecisionMaker evaluation: error sending evaluation sample: %v", err)
		return
	}
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

func (p *RecoveryEvaluationProcessor) progressTime() time.Time {
	res := p.now
	p.now = res.Add(p.SampleRate)
	return res
}
