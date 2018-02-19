package evaluation

import (
	"fmt"
	"strings"

	bitflow "github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline/denstream"
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

	trainingEnd = golib.ParseTime(golib.SimpleTimeLayout, "2018-02-06 22:26:31")
)

type ClusterTagger struct {
	bitflow.AbstractProcessor
}

func (p *ClusterTagger) String() string {
	return "cluster tagger"
}

func (p *ClusterTagger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	clusterId := sample.Tag("cluster")
	predicted := ""
	if clusterId == denstream.OutlierStr || clusterId == denstream.NewOutlierStr {
		predicted = EvalAnomaly
	} else {
		predicted = EvalNormal
	}
	sample.SetTag(EvalPredictedTag, predicted)
	return p.OutgoingSink.Sample(sample, header)
}

type TagsPreprocessor struct {
	bitflow.AbstractProcessor
}

func (p *TagsPreprocessor) String() string {
	return "tags preprocessor"
}

func (p *TagsPreprocessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
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
		if sample.Time.Before(trainingEnd) {
			sample.SetTag(EvaluateTag, "training")
		} else {
			sample.SetTag(EvaluateTag, DoEvaluate)
			sample.SetTag(EvalExpectedTag, EvalNormal)
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
			sample.SetTag(EvaluateTag, DoEvaluate)
			sample.SetTag(EvalExpectedTag, EvalAnomaly)
		}
	}
	sample.SetTag(EvalGroupsTag, strings.Join(groups, "|"))

	return p.OutgoingSink.Sample(sample, header)
}

type EvaluationTagFixer struct {
	bitflow.AbstractProcessor
}

func (p *EvaluationTagFixer) String() string {
	return "evaluation tag fixer"
}

func (p *EvaluationTagFixer) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if sample.HasTag("evalTraining") {
		sample.DeleteTag("evalTraining")
		sample.SetTag(EvaluateTag, "training")
	}
	if sample.HasTag(EvalExpectedTag) {
		sample.SetTag(EvaluateTag, DoEvaluate)
	}
	return p.OutgoingSink.Sample(sample, header)
}
