package evaluation

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

type GroupedEvaluation struct {
	bitflow.NoopProcessor
	groups         map[string]Stats
	ignoredSamples map[string]int

	ConfigurableTags
	TsvHeader string
	NewGroup  func(name string) Stats
}

type ConfigurableTags struct {
	EvaluateTag        string // "evaluate"
	DoEvaluate         string // "true"
	EvalGroupsTag      string // "evalGroups"
	EvalGroupSeparator string // "|"
}

func (e *ConfigurableTags) SetEvaluationTags(params map[string]string) {
	e.EvaluateTag = params["evaluateTag"]
	if e.EvaluateTag == "" {
		e.EvaluateTag = "evaluate"
	}
	e.DoEvaluate = params["evaluateValue"]
	if e.DoEvaluate == "" {
		e.DoEvaluate = "true"
	}
	e.EvalGroupsTag = params["groupsTag"]
	if e.EvalGroupsTag == "" {
		e.EvalGroupsTag = "evalGroups"
	}
	e.EvalGroupSeparator = params["groupsSeparator"]
	if e.EvalGroupSeparator == "" {
		e.EvalGroupSeparator = "|"
	}
}

func (e *ConfigurableTags) String() string {
	return fmt.Sprintf("evaluate tag: \"%v\", evaluate value: \"%v\", groups tag: \"%v\", groups separator: \"%v\"",
		e.EvaluateTag, e.DoEvaluate, e.EvalGroupsTag, e.EvalGroupSeparator)
}

type Stats interface {
	Evaluate(sample *bitflow.Sample, header *bitflow.Header)
	TSV() string
}

func (p *GroupedEvaluation) PrintResults() {
	var buf bytes.Buffer
	writer := new(tabwriter.Writer).Init(&buf, 5, 2, 3, ' ', 0)
	header := []byte("Group\t" + p.TsvHeader)
	writer.Write(append(header, '\n'))
	for _, b := range header {
		if b != '\t' {
			b = '='
		}
		writer.Write([]byte{b})
	}
	writer.Write([]byte{'\n'})
	groupNames := make([]string, 0, len(p.groups))
	for group := range p.groups {
		groupNames = append(groupNames, group)
	}
	sort.Strings(groupNames)

	for _, name := range groupNames {
		writer.Write([]byte(name + "\t" + p.groups[name].TSV() + "\n"))
	}
	writer.Flush()

	log.Println("------------ Evaluation Results")
	log.Printf("------------ Ignored samples: %v", p.ignoredSamples)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		log.Println(line)
	}
}

func (p *GroupedEvaluation) Start(wg *sync.WaitGroup) golib.StopChan {
	p.groups = make(map[string]Stats)
	p.ignoredSamples = make(map[string]int)
	return p.NoopProcessor.Start(wg)
}

func (p *GroupedEvaluation) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	eval := sample.Tag(p.EvaluateTag)
	if eval == p.DoEvaluate {
		groups := strings.Split(sample.Tag(p.EvalGroupsTag), p.EvalGroupSeparator)
		for _, group := range groups {
			stats, ok := p.groups[group]
			if !ok {
				stats = p.NewGroup(group)
				p.groups[group] = stats
			}
			stats.Evaluate(sample, header)
		}
	} else if eval == "" {
		eval = "ignored"
	}
	if eval != p.DoEvaluate {
		ignoredCounter := p.ignoredSamples[eval]
		p.ignoredSamples[eval] = ignoredCounter + 1
	}
	return p.NoopProcessor.Sample(sample, header)
}

func (p *GroupedEvaluation) Close() {
	p.PrintResults()
	p.NoopProcessor.Close()
}
