package recovery

import (
	bitflow "github.com/antongulenko/go-bitflow"
	log "github.com/sirupsen/logrus"
)

var RecoveryTagParams = []string{"node-name", "state", "normal-state"}

type RecoveryTags struct {
	NodeNameTag      string
	StateTag         string
	NormalStateValue string
}

func (t *RecoveryTags) ParseRecoveryTags(params map[string]string) {
	t.NodeNameTag = params["node-name"]
	t.StateTag = params["state"]
	t.NormalStateValue = params["normal-state"]
}

func (t *RecoveryTags) GetRecoveryTags(sample *bitflow.Sample) (name string, state string) {
	if !sample.HasTag(t.NodeNameTag) || !sample.HasTag(t.StateTag) {
		log.Warnf("Ignoring Sample without tag '%v' and/or '%v'", t.NodeNameTag, t.StateTag)
	} else {
		name = sample.Tag(t.NodeNameTag)
		state = sample.Tag(t.StateTag)
	}
	return
}
