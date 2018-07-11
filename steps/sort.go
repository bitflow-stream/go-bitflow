package steps

import (
	"log"
	"sort"
	"strings"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

// Sort based on given Tags, use Timestamp as last sort criterion
type SampleSorter struct {
	Tags []string
}

type SampleSlice struct {
	samples []*bitflow.Sample
	sorter  *SampleSorter
}

func (s SampleSlice) Len() int {
	return len(s.samples)
}

func (s SampleSlice) Less(i, j int) bool {
	a := s.samples[i]
	b := s.samples[j]
	for _, tag := range s.sorter.Tags {
		tagA := a.Tag(tag)
		tagB := b.Tag(tag)
		if tagA == tagB {
			continue
		}
		return tagA < tagB
	}
	return a.Time.Before(b.Time)
}

func (s SampleSlice) Swap(i, j int) {
	s.samples[i], s.samples[j] = s.samples[j], s.samples[i]
}

func (sorter *SampleSorter) ProcessBatch(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
	log.Println("Sorting", len(samples), "samples")
	sort.Sort(SampleSlice{samples, sorter})
	return header, samples, nil
}

func (sorter *SampleSorter) String() string {
	all := make([]string, len(sorter.Tags)+1)
	copy(all, sorter.Tags)
	all[len(all)-1] = "Timestamp"
	return "Sort: " + strings.Join(all, ", ")
}

func RegisterSampleSorter(b *query.PipelineBuilder) {
	b.RegisterAnalysisParams("sort",
		func(p *pipeline.SamplePipeline, params map[string]string) {
			var tags []string
			if tags_param, ok := params["tags"]; ok {
				tags = strings.Split(tags_param, ",")
			}
			p.Batch(&SampleSorter{tags})
		},
		"Sort a batch of samples based on the values of the given comma-separated tags. The default criterion is the timestmap", []string{}, "tags")
}
