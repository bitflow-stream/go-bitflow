package recovery

import (
	"fmt"

	"github.com/antongulenko/go-bitflow"
	log "github.com/sirupsen/logrus"
)

type EvaluationDataCollector struct {
	ConfigurableTags
	StoreEvaluationEvent func(node, state string, samples []*bitflow.Sample, header *bitflow.Header)
}

func (p *EvaluationDataCollector) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	var node, state string
	for _, sample := range samples {
		newNode, newState := p.GetRecoveryTags(sample)
		if node == "" || state == "" {
			node, state = newNode, newState
		} else if newNode != node || newState != state {
			log.Warnf("Dropping batch, which contains multiple values for tags %v (%v and %v) and %v (%v and %v)",
				p.NodeNameTag, node, newNode, p.StateTag, state, newState)
			node = ""
			state = ""
			break
		}
	}
	if node != "" && state != "" {
		p.StoreEvaluationEvent(node, state, samples, header)
	}
	return header, nil, nil
}

func (p *EvaluationDataCollector) String() string {
	return fmt.Sprintf("Collect evaluation data (%v)", p.ConfigurableTags)
}
