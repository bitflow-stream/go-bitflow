package fork

import (
	"bytes"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
)

type RoundRobinDistributor struct {
	NumSubpipelines int
	current         int
}

func (rr *RoundRobinDistributor) Distribute(_ *bitflow.Sample, _ *bitflow.Header) []interface{} {
	cur := rr.current % rr.NumSubpipelines
	rr.current++
	return []interface{}{cur}
}

func (rr *RoundRobinDistributor) String() string {
	return fmt.Sprintf("round robin (%v)", rr.NumSubpipelines)
}

type MultiplexDistributor struct {
	numSubpipelines int
	keys            []interface{}
}

func NewMultiplexDistributor(numSubpipelines int) *MultiplexDistributor {
	multi := &MultiplexDistributor{
		numSubpipelines: numSubpipelines,
		keys:            make([]interface{}, numSubpipelines),
	}
	for i := 0; i < numSubpipelines; i++ {
		multi.keys[i] = i
	}
	return multi
}

func (d *MultiplexDistributor) Distribute(_ *bitflow.Sample, _ *bitflow.Header) []interface{} {
	return d.keys
}

func (d *MultiplexDistributor) String() string {
	return fmt.Sprintf("multiplex (%v)", d.numSubpipelines)
}

type TagsDistributor struct {
	Tags        []string
	Separator   string
	Replacement string // For missing/empty tags
}

func (d *TagsDistributor) Distribute(sample *bitflow.Sample, _ *bitflow.Header) []interface{} {
	var key bytes.Buffer
	for i, tag := range d.Tags {
		if i > 0 {
			key.WriteString(d.Separator)
		}
		value := sample.Tag(tag)
		if value == "" {
			value = d.Replacement
		}
		key.WriteString(value)
	}
	return []interface{}{key.String()}
}

func (d *TagsDistributor) String() string {
	return fmt.Sprintf("tags %v, separated by %v", d.Tags, d.Separator)
}

type StringRemapDistributor struct {
	Mapping map[string]string
}

func (d *StringRemapDistributor) Distribute(forkPath []interface{}) []interface{} {
	input := ""
	for i, path := range forkPath {
		if i > 0 {
			input += " "
		}
		input += fmt.Sprintf("%v", path)
	}
	result, ok := d.Mapping[input]
	if !ok {
		result = ""
		d.Mapping[input] = result
		log.Warnf("[%v]: No mapping found for fork path '%v', mapping to default output", d, input)
	}
	return []interface{}{result}
}

func (d *StringRemapDistributor) String() string {
	return fmt.Sprintf("String remapper (len %v)", len(d.Mapping))
}
