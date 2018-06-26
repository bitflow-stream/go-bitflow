package fork

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/antongulenko/go-bitflow"
	log "github.com/sirupsen/logrus"
)

type RoundRobinDistributor struct {
	NumSubPipelines int
	current         int
}

func (rr *RoundRobinDistributor) Distribute(_ *bitflow.Sample, _ *bitflow.Header) []interface{} {
	cur := rr.current % rr.NumSubPipelines
	rr.current++
	return []interface{}{cur}
}

func (rr *RoundRobinDistributor) String() string {
	return fmt.Sprintf("round robin (%v)", rr.NumSubPipelines)
}

type MultiplexDistributor struct {
	numSubPipelines int
	keys            []interface{}
}

func NewMultiplexDistributor(numSubPipelines int) *MultiplexDistributor {
	multi := &MultiplexDistributor{
		numSubPipelines: numSubPipelines,
		keys:            make([]interface{}, numSubPipelines),
	}
	for i := 0; i < numSubPipelines; i++ {
		multi.keys[i] = i
	}
	return multi
}

func (d *MultiplexDistributor) Distribute(_ *bitflow.Sample, _ *bitflow.Header) []interface{} {
	return d.keys
}

func (d *MultiplexDistributor) String() string {
	return fmt.Sprintf("multiplex (%v)", d.numSubPipelines)
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

func (d *TagsDistributor) String() (res string) {
	if len(d.Tags) == 1 {
		res = "tag " + d.Tags[0]
	} else {
		fmt.Sprintf("tags %v", d.Tags)
	}
	if d.Separator != "" {
		res += ", separated by " + d.Separator
	}
	return
}

type TagTemplateDistributor struct {
	Template string // Placeholders like ${xxx} will be replaced by tag values (left empty if tag missing)
}

var templateRegex = regexp.MustCompile("\\$\\{[^\\{]*\\}") // Example: ${hello}

func (d *TagTemplateDistributor) Distribute(sample *bitflow.Sample, _ *bitflow.Header) []interface{} {
	key := templateRegex.ReplaceAllStringFunc(d.Template, func(placeholder string) string {
		if strings.HasPrefix(placeholder, "${") && strings.HasSuffix(placeholder, "}") {
			return sample.Tag(placeholder[2 : len(placeholder)-1])
		}
		return ""
	})
	return []interface{}{key}
}

func (d *TagTemplateDistributor) String() string {
	return fmt.Sprintf("tag template: %v", d.Template)
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
