package fork

import (
	"fmt"
	"sort"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	log "github.com/sirupsen/logrus"
)

type SimplePipelineBuilder struct {
	Build           func() []bitflow.SampleProcessor
	examplePipeline []fmt.Stringer
}

func (b *SimplePipelineBuilder) ContainedStringers() []fmt.Stringer {
	if b.examplePipeline == nil {
		if b.Build == nil {
			b.examplePipeline = make([]fmt.Stringer, 0)
		} else {
			pipe := b.Build()
			b.examplePipeline = make([]fmt.Stringer, len(pipe))
			for i, step := range pipe {
				b.examplePipeline[i] = step
			}
		}
	}
	return b.examplePipeline
}

func (b *SimplePipelineBuilder) String() string {
	return fmt.Sprintf("Simple Pipeline Builder len %v", len(b.ContainedStringers()))
}

func (b *SimplePipelineBuilder) BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline {
	var res bitflow.SamplePipeline
	if b.Build != nil {
		for _, processor := range b.Build() {
			res.Add(processor)
		}
	}
	res.Add(output)
	return &res
}

type MultiplexPipelineBuilder []*pipeline.SamplePipeline

func (b MultiplexPipelineBuilder) BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline {
	// Type of key must be int, and index must be in range. Otherwise panic!
	pipe := &(b[key.(int)].SamplePipeline)
	pipe.Add(output)
	return pipe
}

func (b MultiplexPipelineBuilder) String() string {
	return fmt.Sprintf("Multiplex builder, %v subpipelines", len(b))
}

func (b MultiplexPipelineBuilder) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, len(b))
	for i, pipe := range b {
		res[i] = pipe
	}
	return res
}

type MultiFilePipelineBuilder struct {
	Config bitflow.FileSink // Configuration parameters in this field will be used for file outputs
}

func (b *MultiFilePipelineBuilder) ContainedStringers() []fmt.Stringer {
	return []fmt.Stringer{pipeline.String(b.String())}
}

func (b *MultiFilePipelineBuilder) String() string {
	return "Output to files named by fork key"
}

func (b *MultiFilePipelineBuilder) BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline {
	var pipe bitflow.SamplePipeline
	fileOut := b.Config
	fileName := fmt.Sprintf("%v", key) // TODO this could be parameterized
	fileOut.Filename = fileName
	pipe.Add(&fileOut).Add(output)
	return &pipe
}

type StringPipelineBuilder struct {
	Pipelines            map[string]*pipeline.SamplePipeline
	BuildMissingPipeline func(string) (*pipeline.SamplePipeline, error)
}

func (b *StringPipelineBuilder) BuildPipeline(key interface{}, _ *ForkMerger) *bitflow.SamplePipeline {
	strKey := ""
	if key != nil {
		strKey = fmt.Sprintf("%v", key)
	}
	pipe, ok := b.Pipelines[strKey]
	if !ok {
		keys := make([]string, 0, len(b.Pipelines))
		for key := range b.Pipelines {
			keys = append(keys, key)
		}
		if b.BuildMissingPipeline != nil {
			var err error
			pipe, err = b.BuildMissingPipeline(strKey)
			log.Debugf("No subpipeline defined for key '%v' (type %T). Building default pipeline (Have pipelines: %v)", strKey, key, keys)
			if err != nil {
				log.Errorf("Failed to build default subpipeline for key '%v': %v", strKey, err)
				pipe = nil
			}
		} else {
			log.Warnf("No subpipeline defined for key '%v' (type %T). Using empty pipeline (Have pipelines: %v)", strKey, key, keys)
			pipe = new(pipeline.SamplePipeline)
		}
		b.Pipelines[strKey] = pipe
	}
	return &pipe.SamplePipeline
}

func (b *StringPipelineBuilder) String() string {
	return fmt.Sprintf("Pipeline builder, %v subpipelines", len(b.Pipelines))
}

func (b *StringPipelineBuilder) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, 0, len(b.Pipelines))
	for key, pipe := range b.Pipelines {
		var title string
		if key == "" {
			title = "Default pipeline"
		} else {
			title = fmt.Sprintf("Pipeline %v", key)
		}
		res = append(res, &pipeline.TitledSamplePipeline{
			SamplePipeline: pipe,
			Title:          title,
		})
	}
	sort.Sort(pipeline.SortedStringers(res))
	return res
}
