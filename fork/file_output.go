package fork

import (
	"fmt"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
)

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
		Description: fmt.Sprintf("directory tree built from subpipeline key"),
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
