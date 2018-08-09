package query

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
)

type ProcessingSteps []JsonProcessingStep

func (slice ProcessingSteps) Len() int {
	return len(slice)
}

func (slice ProcessingSteps) Less(i, j int) bool {
	// Sort the forks after the regular processing steps
	return (!slice[i].IsFork && slice[j].IsFork) || slice[i].Name < slice[j].Name
}

func (slice ProcessingSteps) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type JsonProcessingStep struct {
	Name           string
	IsFork         bool
	Description    string
	RequiredParams []string
	OptionalParams []string
}

func (b PipelineBuilder) getSortedProcessingSteps() ProcessingSteps {
	all := make(ProcessingSteps, 0, len(b.analysis_registry))
	for _, step := range b.analysis_registry {
		if step.Func == nil {
			continue
		}
		all = append(all, JsonProcessingStep{
			Name:           step.Name,
			IsFork:         false,
			Description:    step.Description,
			RequiredParams: step.Params.required,
			OptionalParams: step.Params.optional,
		})
	}
	for _, fork := range b.fork_registry {
		if fork.Func == nil {
			continue
		}
		all = append(all, JsonProcessingStep{
			Name:           fork.Name,
			IsFork:         true,
			Description:    fork.Description,
			RequiredParams: fork.Params.required,
			OptionalParams: fork.Params.optional,
		})
	}
	sort.Sort(all)
	return all
}

func (b PipelineBuilder) PrintAllAnalyses() string {
	all := b.getSortedProcessingSteps()
	var buf bytes.Buffer
	for i, analysis := range all {
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

func (b PipelineBuilder) PrintJsonCapabilities(out io.Writer) error {
	all := b.getSortedProcessingSteps()
	data, err := json.Marshal(all)
	if err == nil {
		_, err = out.Write(data)
	}
	return err
}
