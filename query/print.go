package query

import (
	"bytes"
	"fmt"
	"sort"
)

func (p *SamplePipeline) String() string {
	return "Pipeline"
}

func (p *SamplePipeline) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, 0, len(p.Processors)+2)
	if p.Source != nil {
		res = append(res, p.Source)
	}
	for _, proc := range p.Processors {
		res = append(res, proc)
	}
	if p.Sink != nil {
		res = append(res, p.Sink)
	}
	return res
}

func (builder PipelineBuilder) PrintAllAnalyses() string {
	all := make(SortedAnalyses, 0, len(builder.analysis_registry))
	for _, analysis := range builder.analysis_registry {
		all = append(all, analysis)
	}
	sort.Sort(all)
	var buf bytes.Buffer
	for i, analysis := range all {
		if analysis.Func == nil {
			continue
		}
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(" - ")
		buf.WriteString(analysis.Name)
		buf.WriteString(":\n")
		buf.WriteString("      ")
		buf.WriteString(analysis.Description)
	}
	return buf.String()
}

type SortedAnalyses []registeredAnalysis

func (slice SortedAnalyses) Len() int {
	return len(slice)
}

func (slice SortedAnalyses) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice SortedAnalyses) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
