package main

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/antongulenko/go-bitflow"
	. "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

// TODO make this configurable as parameters of file inputs etc.

func RegisterTaggingAnalyses(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("set_filename", set_filename_tag,
		"When reading files, instead of using the entire path for source_tag, use only the given level in the directory tree (0 is the file, 1 the containing directory name, 2 the parent directory, ...)",
		[]string{"level"})

	b.RegisterAnalysisParams("source_tag",
		func(p *SamplePipeline, params map[string]string) {
			set_sample_tagger(p, params["tag"], false)
		}, "Set the name of the data source to the given tag. For files it is the file path (except see set_filename), for TCP connections it is the remote endpoint",
		[]string{"tag"})

	b.RegisterAnalysisParams("source_tag_append",
		func(p *SamplePipeline, params map[string]string) {
			set_sample_tagger(p, params["tag"], true)
		}, "Like source_tag, but don't override existing tag values. Instead, append a new tag with an incremented tag name",
		[]string{"tag"})
}

type SampleTagger struct {
	SourceTags    []string
	DontOverwrite bool
}

func (h *SampleTagger) HandleSample(sample *bitflow.Sample, source string) {
	for _, tag := range h.SourceTags {
		if h.DontOverwrite {
			base := tag
			tag = base
			for i := 0; sample.HasTag(tag); i++ {
				tag = base + strconv.Itoa(i)
			}
		}
		sample.SetTag(tag, source)
	}
}

func set_sample_tagger(p *SamplePipeline, tag string, dontOverwrite bool) {
	if source, ok := p.Source.(bitflow.UnmarshallingSampleSource); ok {
		source.SetSampleHandler(&SampleTagger{SourceTags: []string{tag}, DontOverwrite: dontOverwrite})
	}
}

func set_filename_tag(p *SamplePipeline, params map[string]string) error {
	num, err := strconv.ParseUint(params["level"], 10, 64)
	if err != nil {
		return query.ParameterError("level", err)
	}

	if fileSource, ok := p.Source.(*bitflow.FileSource); ok {
		fileSource.ConvertFilename = func(filename string) string {
			for i := uint64(0); i < num; i++ {
				filename = filepath.Dir(filename)
			}
			return filepath.Base(filename)
		}
		return nil
	} else {
		return fmt.Errorf("Data source must be *bitflow.FileSource, but was %T instead: %v", p.Source, p.Source)
	}
}
