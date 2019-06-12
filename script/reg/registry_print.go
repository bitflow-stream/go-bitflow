package reg

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Name        string
	IsFork      bool
	IsBatch     bool
	Description string
	Params      []JsonParameter
}

type JsonParameter struct {
	Name     string
	Type     string
	Default  interface{}
	Required bool
}

func makeJsonProcessingStep(reg RegisteredStep, batch, fork bool) JsonProcessingStep {
	return JsonProcessingStep{
		Name:        reg.Name,
		Description: reg.Description,
		IsFork:      fork,
		IsBatch:     batch,
		Params:      makeJsonParameters(reg.Params),
	}
}

func makeJsonParameters(params RegisteredParameters) []JsonParameter {
	result := make([]JsonParameter, 0, len(params))
	for _, param := range params {
		result = append(result, JsonParameter{
			Name:     param.Name,
			Type:     param.Parser.String(),
			Default:  param.Default,
			Required: param.Required,
		})
	}
	return result
}

func (s *JsonProcessingStep) formatTo(buf *bytes.Buffer) {
	buf.WriteString("\n - ")
	buf.WriteString(s.Name)
	if s.Description != "" {
		buf.WriteString("\n      Description: ")
		buf.WriteString(s.Description)
	}
	if len(s.Params) > 0 {
		buf.WriteString("\n      Parameters:")
		for _, param := range s.Params {
			requiredOrOptional := "required"
			if !param.Required {
				requiredOrOptional = fmt.Sprintf("optional, default: %v", param.Default)
			}
			fmt.Fprintf(buf, "\n          %v (%v), %v", param.Name, param.Type, requiredOrOptional)
		}
	}
}

func (r ProcessorRegistry) getSortedProcessingSteps() (steps ProcessingSteps, batchSteps ProcessingSteps, forks ProcessingSteps) {
	steps = make(ProcessingSteps, 0, len(r.stepRegistry))
	batchSteps = make(ProcessingSteps, 0, len(r.batchRegistry))
	forks = make(ProcessingSteps, 0, len(r.forkRegistry))

	for _, step := range r.stepRegistry {
		if step.Func != nil {
			steps = append(steps, makeJsonProcessingStep(step.RegisteredStep, false, false))
		}
	}
	for _, step := range r.batchRegistry {
		if step.Func != nil {
			batchSteps = append(batchSteps, makeJsonProcessingStep(step.RegisteredStep, true, false))
		}
	}
	for _, step := range r.forkRegistry {
		if step.Func != nil {
			forks = append(forks, makeJsonProcessingStep(step.RegisteredStep, false, false))
		}
	}

	sort.Sort(steps)
	sort.Sort(batchSteps)
	sort.Sort(forks)
	return
}

func (r ProcessorRegistry) formatSection(buf *bytes.Buffer, steps ProcessingSteps, title string, started bool) bool {
	if len(steps) > 0 {
		if started {
			buf.WriteString("\n")
		}
		buf.WriteString(title)
		for _, step := range steps {
			step.formatTo(buf)
		}
		started = true
	}
	return started
}

func (r ProcessorRegistry) FormatCapabilities(out io.Writer) error {
	steps, batchSteps, forks := r.getSortedProcessingSteps()
	started := false
	var buf bytes.Buffer
	started = r.formatSection(&buf, steps, "Processing steps:", started)
	started = r.formatSection(&buf, batchSteps, "Batch processing steps:", started)
	started = r.formatSection(&buf, forks, "Forks:", started)
	buf.WriteString("\n\n")
	_, err := buf.WriteTo(out)
	return err
}

func (r ProcessorRegistry) FormatJsonCapabilities(out io.Writer) error {
	steps, batchSteps, forks := r.getSortedProcessingSteps()
	data, err := json.Marshal(map[string]interface{}{
		"steps":       steps,
		"batch_steps": batchSteps,
		"forks":       forks,
	})
	if err == nil {
		_, err = out.Write(data)
	}
	return err
}
