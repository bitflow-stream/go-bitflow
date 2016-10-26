package pipeline

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type ForkDistributor interface {
	Distribute(sample *bitflow.Sample, header *bitflow.Header) []interface{}
	String() string
}

type PipelineBuilder interface {
	BuildPipeline(key interface{}, output *AggregatingSink) *bitflow.SamplePipeline
	String() string
}

type MetricFork struct {
	bitflow.AbstractProcessor

	Distributor   ForkDistributor
	Builder       PipelineBuilder
	ParallelClose bool

	pipelines        map[interface{}]bitflow.MetricSink
	lastHeaders      map[interface{}]*bitflow.Header
	runningPipelines int
	stopped          bool
	stoppedCond      sync.Cond
	subpipelineWg    sync.WaitGroup
	aggregatingSink  AggregatingSink
}

func NewMetricFork(distributor ForkDistributor, builder PipelineBuilder) *MetricFork {
	fork := &MetricFork{
		Builder:       builder,
		Distributor:   distributor,
		pipelines:     make(map[interface{}]bitflow.MetricSink),
		lastHeaders:   make(map[interface{}]*bitflow.Header),
		ParallelClose: true,
	}
	fork.stoppedCond.L = new(sync.Mutex)
	fork.aggregatingSink.fork = fork
	return fork
}

func (f *MetricFork) Start(wg *sync.WaitGroup) golib.StopChan {
	result := f.AbstractProcessor.Start(wg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		f.waitForSubpipelines()
		f.CloseSink(wg)
	}()
	return result
}

func (f *MetricFork) Header(header *bitflow.Header) error {
	// Drop header here, only send to respective subpipeline if changed.
	return nil
}

func (f *MetricFork) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := f.Check(sample, header); err != nil {
		return err
	}
	keys := f.Distributor.Distribute(sample, header)
	var errors golib.MultiError
	for _, key := range keys {
		pipeline, ok := f.pipelines[key]
		if !ok {
			pipeline = f.newPipeline(key)
		}
		if pipeline != nil {
			if lastHeader, ok := f.lastHeaders[key]; !ok || !lastHeader.Equals(header) {
				if err := pipeline.Header(header); err != nil {
					errors.Add(err)
					continue
				}
				f.lastHeaders[key] = header
			}
			err := pipeline.Sample(sample, header)
			errors.Add(err)
		}
	}
	return errors.NilOrError()
}

func (f *MetricFork) Close() {
	f.stoppedCond.L.Lock()
	defer f.stoppedCond.L.Unlock()
	f.stopped = true

	var wg sync.WaitGroup
	for i, pipeline := range f.pipelines {
		f.pipelines[i] = nil // Enable GC
		if pipeline != nil {
			wg.Add(1)
			go func(pipeline bitflow.MetricSink) {
				defer wg.Done()
				pipeline.Close()
			}(pipeline)
			if !f.ParallelClose {
				wg.Wait()
			}
		}
	}
	wg.Wait()
	f.stoppedCond.Broadcast()
}

func (f *MetricFork) String() string {
	return fmt.Sprintf("Fork. Distributor: %v, Builder: %v", f.Distributor, f.Builder)
}

func (f *MetricFork) newPipeline(key interface{}) bitflow.MetricSink {
	log.Debugf("[%v]: Starting forked subpipeline %v", f, key)
	pipeline := f.Builder.BuildPipeline(key, &f.aggregatingSink)
	group := golib.NewTaskGroup()
	pipeline.Source = nil // The source is already started
	pipeline.Construct(group)
	f.runningPipelines++
	waitingTasks, channels := group.StartTasks(&f.subpipelineWg)
	f.subpipelineWg.Add(1)

	go func() {
		defer f.subpipelineWg.Done()

		idx, err := golib.WaitForAny(channels)
		if err != nil {
			log.Errorln(err)
		}
		group.ReverseStop(golib.DefaultPrintTaskStopWait) // The Stop() calls are actually ignored because group contains only MetricSinks
		_ = golib.PrintErrors(channels, waitingTasks, golib.DefaultPrintTaskStopWait)

		if idx == -1 {
			// Inactive subpipeline can occur when all processors and the sink return nil from Start().
			// This means that they simply react on Header()/Sample() and wait for the final Close() call.
			log.Debugf("[%v]: Subpipeline inactive: %v", f, key)
		} else {
			log.Debugf("[%v]: Finished forked subpipeline %v", f, key)
		}

		f.stoppedCond.L.Lock()
		defer f.stoppedCond.L.Unlock()
		f.runningPipelines--
		f.stoppedCond.Broadcast()
	}()

	var first bitflow.MetricSink
	if len(pipeline.Processors) == 0 {
		first = pipeline.Sink
	} else {
		first = pipeline.Processors[0]
	}
	f.pipelines[key] = first
	return first
}

func (f *MetricFork) waitForSubpipelines() {
	f.stoppedCond.L.Lock()
	for !f.stopped || f.runningPipelines > 0 {
		f.stoppedCond.Wait()
	}
	f.stoppedCond.L.Unlock()
	f.subpipelineWg.Wait()
}

type AggregatingSink struct {
	bitflow.AbstractMetricSink
	fork *MetricFork
}

func (sink *AggregatingSink) String() string {
	return "aggregating sink for " + sink.fork.String()
}

func (sink *AggregatingSink) Start(wg *sync.WaitGroup) golib.StopChan {
	return nil
}

func (sink *AggregatingSink) Close() {
	// The actual outgoing sink is closed after waitForSubpipelines() returns
}

func (sink *AggregatingSink) Header(header *bitflow.Header) error {
	if err := sink.fork.CheckSink(); err != nil {
		return err
	}
	return sink.fork.OutgoingSink.Header(header)
}

func (sink *AggregatingSink) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if err := sink.fork.Check(sample, header); err != nil {
		return err
	}
	return sink.fork.OutgoingSink.Sample(sample, header)
}

func (sink *AggregatingSink) GetOriginalSink() bitflow.MetricSink {
	return sink.fork.OutgoingSink
}

type RoundRobinDistributor struct {
	NumSubpipelines int
	current         int
}

func (rr *RoundRobinDistributor) Distribute(_ *bitflow.Sample, _ *bitflow.Header) []interface{} {
	cur := rr.current % rr.NumSubpipelines
	rr.current++
	return []interface{}{cur}
}

func (rr *RoundRobinDistributor) String() string {
	return fmt.Sprintf("round robin (%v)", rr.NumSubpipelines)
}

type MultiplexDistributor struct {
	numSubpipelines int
	keys            []interface{}
}

func NewMultiplexDistributor(numSubpipelines int) *MultiplexDistributor {
	multi := &MultiplexDistributor{
		numSubpipelines: numSubpipelines,
		keys:            make([]interface{}, numSubpipelines),
	}
	for i := 0; i < numSubpipelines; i++ {
		multi.keys[i] = i
	}
	return multi
}

func (d *MultiplexDistributor) Distribute(_ *bitflow.Sample, _ *bitflow.Header) []interface{} {
	return d.keys
}

func (d *MultiplexDistributor) String() string {
	return fmt.Sprintf("multiplex (%v)", d.numSubpipelines)
}

type TagsDistributor struct {
	Tags        []string
	Separator   string
	Replacement string // For missing/empty tags
}

func (d *TagsDistributor) Distribute(sample *bitflow.Sample, _ *bitflow.Header) []interface{} {
	var key bytes.Buffer
	for i, tag := range d.Tags {
		if i > 0 {
			key.WriteString(d.Separator)
		}
		value := sample.Tag(tag)
		if value == "" {
			value = d.Replacement
		}
		key.WriteString(value)
	}
	return []interface{}{key.String()}
}

func (d *TagsDistributor) String() string {
	return fmt.Sprintf("tags %v, separated by %v", d.Tags, d.Separator)
}

type SimplePipelineBuilder struct {
	Build           func() []bitflow.SampleProcessor
	examplePipeline []bitflow.SampleProcessor
}

func (b *SimplePipelineBuilder) String() string {
	if b.examplePipeline == nil {
		if b.Build == nil {
			b.examplePipeline = make([]bitflow.SampleProcessor, 0)
		} else {
			b.examplePipeline = b.Build()
		}
	}
	return fmt.Sprintf("simple %v", b.examplePipeline)
}

func (b *SimplePipelineBuilder) BuildPipeline(key interface{}, output *AggregatingSink) *bitflow.SamplePipeline {
	var res bitflow.SamplePipeline
	res.Sink = output
	if b.Build != nil {
		for _, processor := range b.Build() {
			res.Add(processor)
		}
	}
	return &res
}

type MultiFilePipelineBuilder struct {
	SimplePipelineBuilder
	NewFile     func(originalFile string, key interface{}) string
	Description string
}

func (b *MultiFilePipelineBuilder) String() string {
	_ = b.SimplePipelineBuilder.String()
	return fmt.Sprintf("MultiFiles %v %v", b.Description, b.examplePipeline)
}

func (b *MultiFilePipelineBuilder) BuildPipeline(key interface{}, output *AggregatingSink) *bitflow.SamplePipeline {
	simple := b.SimplePipelineBuilder.BuildPipeline(key, output)
	files := b.getFileSink(output.GetOriginalSink())
	if files != nil {
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

func (b *MultiFilePipelineBuilder) getFileSink(sink bitflow.MetricSink) *bitflow.FileSink {
	if files, ok := sink.(*bitflow.FileSink); ok {
		return files
	}
	if agg, ok := sink.(bitflow.AggregateSink); ok {
		var files *bitflow.FileSink
		warned := false
		for _, sink := range agg {
			converted := b.getFileSink(sink)
			if converted != nil {
				if files == nil {
					files = converted
				} else if !warned {
					log.Warnln("[%v]: Multiple file outputs, using %v", b, files)
					warned = true
				}
			}
		}
		return files
	}
	return nil
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
