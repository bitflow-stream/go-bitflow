package main

import (
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
)

func init() {
	RegisterSampleHandler("source_tag", func(param string) bitflow.ReadSampleHandler {
		if param == "" {
			log.Fatalln("Sample handler source_tag needs a parameter")
		}
		return &SampleTagger{SourceTags: []string{param}}
	})
	RegisterSampleHandler("source_tag_append", func(param string) bitflow.ReadSampleHandler {
		if param == "" {
			log.Fatalln("Sample handler source_tag needs a parameter")
		}
		return &SampleTagger{SourceTags: []string{param}, DontOverwrite: true}
	})
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
