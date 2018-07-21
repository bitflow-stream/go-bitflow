package evaluation

import (
	"fmt"
	"strings"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

var (
	balancerHypervisors = map[string]bool{
		"wally183": true,
		"wally192": true,
	}
	backendHypervisors = map[string]bool{
		"wally193": true,
		"wally194": true,
		"wally195": true,
		"wally197": true,
		"wally198": true,
	}
)

type CitTagsPreprocessor struct {
	bitflow.NoopProcessor

	EvaluationTags
	BinaryEvaluationTags
	NormalTagValue string
	TrainingEnd    time.Time
}

func RegisterCitTagsPreprocessor(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) error {
		proc := &CitTagsPreprocessor{
			NormalTagValue: params["normalValue"],
		}
		if trainingEndStr, ok := params["trainingEnd"]; ok {
			trainingEnd, err := time.Parse(golib.SimpleTimeLayout, trainingEndStr)
			if err != nil {
				return query.ParameterError("trainingEnd", err)
			}
			proc.TrainingEnd = trainingEnd
		}
		proc.SetBinaryEvaluationTags(params)
		proc.SetEvaluationTags(params)
		if proc.NormalTagValue == "" {
			proc.NormalTagValue = "normal"
		}
		p.Add(proc)
		return nil
	}
	b.RegisterAnalysisParamsErr("preprocess_cit_tags", create, "Process 'host', 'cls' and 'target' tags into more useful information.", []string{}, "expectedTag", "predictedTag", "anomalyValue", "evaluateTag", "evaluateValue", "groupsTag", "groupsSeparator", "trainingEnd", "normalValue")
}

func (p *CitTagsPreprocessor) String() string {
	return fmt.Sprintf("tags preprocessor (training end: %v, normal tag value: \"%v\", evaluation: [%v], binary evaluation: [%v])",
		p.TrainingEnd, p.NormalTagValue, &p.EvaluationTags, &p.BinaryEvaluationTags)
}

func (p *CitTagsPreprocessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	host := sample.Tag("host")

	// Infer the layer and component from the host name
	layer := ""
	component := ""
	if strings.HasPrefix(host, "wally") {
		layer = "physical"
		if balancerHypervisors[host] {
			component = "balancer"
		} else if backendHypervisors[host] {
			component = "backend"
		}
	} else if strings.HasPrefix(host, "vod-") {
		layer = "virtual"
		if strings.HasPrefix(host, "vod-balancer-") {
			component = "balancer"
		} else if strings.HasPrefix(host, "vod-video-") {
			component = "backend"
		}
	}
	if layer == "" || component == "" {
		return fmt.Errorf("Unexpected host: " + host)
	}
	sample.SetTag("layer", layer)
	sample.SetTag("component", component)

	// Evaluate the 'cls' and 'target' tags, infer information required for the evaluation
	cls := sample.Tag("cls")
	target := sample.Tag("target")
	groups := []string{"all", "host_" + host, "layer_" + layer, "component_" + component, "layer/component_" + layer + "_" + component}
	if sample.HasTag("cls") {
		groups = append(groups, "all_normal")
		sample.SetTag("anomaly", "normal")
		if sample.HasTag("target") {
			// At one point, the experiment controller failed to reset the target tag on a few hosts
			log.Warnf("Sample has both 'cls' and 'target' tags. Removing 'target' (cls=%v, target=%v).", cls, target)
			sample.DeleteTag("target")
		}
		if !p.TrainingEnd.IsZero() && sample.Time.Before(p.TrainingEnd) {
			sample.SetTag(p.EvaluateTag, "training")
		} else {
			sample.SetTag(p.EvaluateTag, p.DoEvaluate)
			sample.SetTag(p.Expected, p.NormalTagValue)
		}
	} else if sample.HasTag("target") {
		targetParts := strings.SplitN(target, "|", 2)
		targetHost := targetParts[0]
		anomaly := targetParts[1]
		sample.SetTag("anomaly", anomaly)
		sample.SetTag("targetHost", targetHost)
		groups = append(groups, "anomaly_"+anomaly, "all_anomalies")

		// Only evaluate hosts that are target of the current injection
		if targetHost == host {
			sample.SetTag(p.EvaluateTag, p.DoEvaluate)
			sample.SetTag(p.Expected, p.AnomalyValue)
		}
	}
	sample.SetTag(p.EvalGroupsTag, strings.Join(groups, p.EvalGroupSeparator))

	return p.NoopProcessor.Sample(sample, header)
}
