package fork

import (
	"fmt"
	"path/filepath"

	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
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
	res.Sink = output
	if b.Build != nil {
		for _, processor := range b.Build() {
			res.Add(processor)
		}
	}
	return &res
}

type MultiplexPipelineBuilder []*pipeline.SamplePipeline

func (b MultiplexPipelineBuilder) BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline {
	// Type of key must be int, and index must be in range. Otherwise panic!
	pipe := &(b[key.(int)].SamplePipeline)
	if pipe.Sink == nil {
		pipe.Sink = output
	}
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
	SimplePipelineBuilder
	NewFile     func(originalFile string, key interface{}) string
	Description string
}

func (b *MultiFilePipelineBuilder) String() string {
	_ = b.SimplePipelineBuilder.String() // Fill the examplePipeline field
	if len(b.examplePipeline) == 0 {
		return fmt.Sprintf("MultiFiles %v", b.Description)
	} else {
		return fmt.Sprintf("MultiFiles %v (subpipeline: %v)", b.Description, b.examplePipeline)
	}
}

func (b *MultiFilePipelineBuilder) BuildPipeline(key interface{}, output *ForkMerger) *bitflow.SamplePipeline {
	simple := b.SimplePipelineBuilder.BuildPipeline(key, output)
	files, ok := output.GetOriginalSink().(*bitflow.FileSink)
	if ok {
		newFilename := b.NewFile(files.Filename, key)
		newFiles := &bitflow.FileSink{
			AbstractMarshallingMetricSink: files.AbstractMarshallingMetricSink,
			Filename:                      newFilename,
			CleanFiles:                    files.CleanFiles,
			IoBuffer:                      files.IoBuffer,
		}
		simple.Sink = newFiles
	} else {
		log.Warnf("[%v]: Cannot assign new files, did not find *bitflow.FileSink as my direct output", b)
	}
	return simple
}

func MultiFileSuffixBuilder(buildPipeline func() []bitflow.SampleProcessor) *MultiFilePipelineBuilder {
	builder := &MultiFilePipelineBuilder{
		Description: "files suffixed with subpipeline key",
		NewFile: func(oldFile string, key interface{}) string {
			suffix := fmt.Sprintf("%v", key)
			group := bitflow.NewFileGroup(oldFile)
			return group.BuildFilenameStr(suffix)
		},
	}
	builder.Build = buildPipeline
	return builder
}

func MultiFileDirectoryBuilder(replaceFilename bool, buildPipeline func() []bitflow.SampleProcessor) *MultiFilePipelineBuilder {
	builder := &MultiFilePipelineBuilder{
		Description: "directory tree built from subpipeline key",
		NewFile: func(oldFile string, key interface{}) string {
			path := fmt.Sprintf("%v", key)
			if path == "" {
				return oldFile
			}
			if replaceFilename {
				path += filepath.Ext(oldFile)
			} else {
				path = filepath.Join(path, filepath.Base(oldFile))
			}
			return filepath.Join(filepath.Dir(oldFile), path)
		},
	}
	builder.Build = buildPipeline
	return builder
}

type StringPipelineBuilder struct {
	Pipelines map[string]*pipeline.SamplePipeline
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
		log.Warnf("No subpipeline defined for key '%v' (type %T). Using empty pipeline (Have pipelines: %v)", strKey, key, keys)
		pipe = new(pipeline.SamplePipeline)
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
