package main

import (
	"errors"
	"path/filepath"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
)

func init() {
	RegisterAnalysisParams("set_filename", set_filename_tag,
		"Number of levels to go up the directory. 0 means filename, etc. Requires file input and -e source_tag")

	RegisterAnalysisParams("source_tag",
		func(p *SamplePipeline, param string) {
			set_sample_tagger(p, param, false)
		}, "the tag to set as the data source")
	RegisterAnalysisParams("source_tag_append",
		func(p *SamplePipeline, param string) {
			set_sample_tagger(p, param, true)
		}, "the tag to set as the data source (will create new tag if already present)")
}

type SampleTagger struct {
	SourceTags    []string
	DontOverwrite bool
}

func (h *SampleTagger) HandleHeader(header *bitflow.Header, source string) {
	header.HasTags = true
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
	if tag == "" {
		log.Fatalln("Sample tagger needs a parameter")
	}
	if source, ok := p.Source.(bitflow.UnmarshallingMetricSource); ok {
		source.SetSampleHandler(&SampleTagger{SourceTags: []string{tag}, DontOverwrite: dontOverwrite})
	}
}

func set_filename_tag(p *SamplePipeline, param string) {
	num, err := strconv.Atoi(param)
	if err == nil && num < 0 {
		err = errors.New("Number must be >= 0")
	}
	if err != nil {
		log.Fatalf("Failed to parse parameter for -e set_filename: %v", err)
	}

	if filesource, ok := p.Source.(*bitflow.FileSource); ok {
		filesource.ConvertFilename = func(filename string) string {
			for i := 0; i < num; i++ {
				filename = filepath.Dir(filename)
			}
			return filepath.Base(filename)
		}
	} else {
		log.Warnf("Cannot apply set_filename: data source is not *bitflow.FileSource but %T", p.Source)
	}
}
