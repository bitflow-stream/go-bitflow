package evaluation

import (
	"bytes"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"

	bitflow "github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

const (
	EvaluateTag        = "evaluate"
	DoEvaluate         = "true"
	EvalGroupsTag      = "evalGroups"
	EvalGroupSeparator = "|"
)

type GroupedEvaluation struct {
	bitflow.AbstractProcessor
	groups         map[string]EvaluationStats
	ignoredSamples map[string]int

	TsvHeader string
	NewGroup  func(name string) EvaluationStats
}

type EvaluationStats interface {
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
	p.groups = make(map[string]EvaluationStats)
	p.ignoredSamples = make(map[string]int)
	return p.AbstractProcessor.Start(wg)
}

func (p *GroupedEvaluation) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	eval := sample.Tag(EvaluateTag)
	if eval == DoEvaluate {
		groups := strings.Split(sample.Tag(EvalGroupsTag), EvalGroupSeparator)
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
	if eval != DoEvaluate {
		ignoredCounter := p.ignoredSamples[eval]
		p.ignoredSamples[eval] = ignoredCounter + 1
	}
	return p.OutgoingSink.Sample(sample, header)
}

func (p *GroupedEvaluation) Close() {
	p.PrintResults()
	p.AbstractProcessor.Close()
}
