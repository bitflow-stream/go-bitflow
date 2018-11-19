package recovery

import (
	"fmt"

	"github.com/bitflow-stream/go-bitflow"
	log "github.com/sirupsen/logrus"
)

var TagParameterNames = []string{"node-name", "state", "normal-state"}

type ConfigurableTags struct {
	NodeNameTag      string
	StateTag         string
	NormalStateValue string

	warned bool
}

func (t ConfigurableTags) String() string {
	return fmt.Sprintf("node-name: %v, normal-state: %v=%v", t.NodeNameTag, t.StateTag, t.NormalStateValue)
}

func (t *ConfigurableTags) ParseRecoveryTags(params map[string]string) {
	t.NodeNameTag = params["node-name"]
	t.StateTag = params["state"]
	t.NormalStateValue = params["normal-state"]
}

func (t *ConfigurableTags) GetRecoveryTags(sample *bitflow.Sample) (name string, state string) {
	if !sample.HasTag(t.NodeNameTag) || !sample.HasTag(t.StateTag) {
		if !t.warned {
			log.Warnf("Ignoring samples without tag '%v' and/or '%v'", t.NodeNameTag, t.StateTag)
			t.warned = true
		}
	} else {
		name = sample.Tag(t.NodeNameTag)
		state = sample.Tag(t.StateTag)
	}
	return
}
