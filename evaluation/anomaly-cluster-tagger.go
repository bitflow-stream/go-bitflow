package evaluation

import (
	"fmt"
	"strconv"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/clustering/denstream"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

type AnomalyClusterTagger struct {
	bitflow.NoopProcessor
	BinaryEvaluationTags
	NormalTagValue        string
	TreatOutliersAsNormal bool
}

func (p *AnomalyClusterTagger) String() string {
	return fmt.Sprintf("cluster tagger (normal tag value: \"%v\", binary evaluation: [%v])",
		p.NormalTagValue, &p.BinaryEvaluationTags)
}

func (p *AnomalyClusterTagger) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	clusterId := sample.Tag("cluster")
	predicted := ""
	switch clusterId {
	case denstream.NewOutlierStr:
		predicted = p.AnomalyValue
	case denstream.OutlierStr:
		if p.TreatOutliersAsNormal {
			predicted = p.NormalTagValue
		} else {
			predicted = p.AnomalyValue
		}
	default:
		predicted = p.NormalTagValue
	}
	sample.SetTag(p.Predicted, predicted)
	return p.NoopProcessor.Sample(sample, header)
}

func RegisterAnomalyClusterTagger(b *query.PipelineBuilder) {
	create := func(p *pipeline.SamplePipeline, params map[string]string) error {
		proc := &AnomalyClusterTagger{
			NormalTagValue: params["normalValue"],
		}
		if treatNormalStr, ok := params["treatOutliersAsNormal"]; ok {
			treatNormal, err := strconv.ParseBool(treatNormalStr)
			if err != nil {
				return query.ParameterError("treatOutliersAsNormal", err)
			}
			proc.TreatOutliersAsNormal = treatNormal
		}
		proc.SetBinaryEvaluationTags(params)
		if proc.NormalTagValue == "" {
			proc.NormalTagValue = "normal"
		}
		p.Add(proc)
		return nil
	}
	b.RegisterAnalysisParamsErr("cluster_tag", create, "Translate 'cluster' tag into predicted=normal (for cluster >= 0) or predicted=anomaly (for cluster == -1 and cluster == -2)", []string{}, "predictedTag", "anomalyValue", "normalValue", "treatOutliersAsNormal")
}
