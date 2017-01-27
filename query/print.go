package query

import (
	"bytes"
	"sort"
)

func (p *SamplePipeline) Print() []string {
	return p.print("")
}

func (p *SamplePipeline) print(indent string) []string {
	processors := p.Processors
	if len(processors) == 0 {
		return []string{"Empty analysis pipeline"}
	} else if len(processors) == 1 {
		return []string{"Analysis: " + processors[0].String()}
	} else {
		res := make([]string, 0, len(processors)+1)
		res = append(res, "Analysis pipeline:")
		for i, proc := range processors {
			indent := "├─"
			if i == len(processors)-1 {
				indent = "└─"
			}
			res = append(res, indent+proc.String())
		}
		return res
	}
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
