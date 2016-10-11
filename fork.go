package analysis

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
)

type ForkDistributor interface {
	Distribute(sample *sample.Sample, header *sample.Header) []interface{}
	String() string
}

type PipelineBuilder interface {
	BuildPipeline(key interface{}, output *AggregatingSink) *sample.SamplePipeline
	String() string
}

type MetricFork struct {
	AbstractProcessor

	Distributor ForkDistributor
	Builder     PipelineBuilder

	pipelines        map[interface{}]sample.MetricSink
	lastHeaders      map[interface{}]*sample.Header
	runningPipelines int
	stopped          bool
	stoppedCond      sync.Cond
	subpipelineWg    sync.WaitGroup
	aggregatingSink  AggregatingSink
}

func NewMetricFork(distributor ForkDistributor, builder PipelineBuilder) *MetricFork {
	fork := &MetricFork{
		Builder:     builder,
		Distributor: distributor,
		pipelines:   make(map[interface{}]sample.MetricSink),
		lastHeaders: make(map[interface{}]*sample.Header),
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

func (f *MetricFork) Header(header *sample.Header) error {
	// Drop header here, only send to respective subpipeline if changed.
	return nil
}

func (f *MetricFork) Sample(sample *sample.Sample, header *sample.Header) error {
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
	for _, pipeline := range f.pipelines {
		if pipeline != nil {
			pipeline.Close()
		}
	}
	f.stoppedCond.Broadcast()
}

func (f *MetricFork) String() string {
	return fmt.Sprintf("Fork. Distributor: %v, Builder: %v", f.Distributor, f.Builder)
}

func (f *MetricFork) newPipeline(key interface{}) sample.MetricSink {
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

		f.stoppedCond.L.Lock()
		defer f.stoppedCond.L.Unlock()
		if idx == -1 {
			// Inactive subpipeline can occur when all processors and the sink return nil from Start().
			// This means that they simply react on Header()/Sample() and wait for the final Close() call.
			log.Debugf("[%v]: Subpipeline inactive: %v", f, key)
		} else {
			log.Debugf("[%v]: Finished forked subpipeline %v", f, key)
		}
		f.runningPipelines--
		f.stoppedCond.Broadcast()
	}()

	var first sample.MetricSink
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
	sample.AbstractMetricSink
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

func (sink *AggregatingSink) Header(header *sample.Header) error {
	if err := sink.fork.CheckSink(); err != nil {
		return err
	}
	return sink.fork.OutgoingSink.Header(header)
}

func (sink *AggregatingSink) Sample(sample *sample.Sample, header *sample.Header) error {
	if err := sink.fork.Check(sample, header); err != nil {
		return err
	}
	return sink.fork.OutgoingSink.Sample(sample, header)
}

func (sink *AggregatingSink) GetOriginalSink() sample.MetricSink {
	return sink.fork.OutgoingSink
}

type RoundRobinDistributor struct {
	NumSubpipelines int
	current         int
}

func (rr *RoundRobinDistributor) Distribute(_ *sample.Sample, _ *sample.Header) []interface{} {
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

func (d *MultiplexDistributor) Distribute(_ *sample.Sample, _ *sample.Header) []interface{} {
	return d.keys
}

func (d *MultiplexDistributor) String() string {
	return fmt.Sprintf("multiplex (%v)", d.numSubpipelines)
}

type TagDistributor struct {
	Tag string
}

func (d *TagDistributor) Distribute(sample *sample.Sample, _ *sample.Header) []interface{} {
	return []interface{}{sample.Tag(d.Tag)}
}

func (d *TagDistributor) String() string {
	return fmt.Sprintf("tags (%v)", d.Tag)
}

type SimplePipelineBuilder struct {
	Build           func() []sample.SampleProcessor
	examplePipeline []sample.SampleProcessor
}

func (b *SimplePipelineBuilder) String() string {
	if b.examplePipeline == nil {
		if b.Build == nil {
			b.examplePipeline = make([]sample.SampleProcessor, 0)
		} else {
			b.examplePipeline = b.Build()
		}
	}
	return fmt.Sprintf("simple %v", b.examplePipeline)
}

func (b *SimplePipelineBuilder) BuildPipeline(key interface{}, output *AggregatingSink) *sample.SamplePipeline {
	var res sample.SamplePipeline
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

func (b *MultiFilePipelineBuilder) BuildPipeline(key interface{}, output *AggregatingSink) *sample.SamplePipeline {
	simple := b.SimplePipelineBuilder.BuildPipeline(key, output)
	files := b.getFileSink(output.GetOriginalSink())
	if files != nil {
		newFilename := b.NewFile(files.Filename, key)
		newFiles := &sample.FileSink{
			AbstractMarshallingMetricSink: files.AbstractMarshallingMetricSink,
			Filename:                      newFilename,
			CleanFiles:                    files.CleanFiles,
			IoBuffer:                      files.IoBuffer,
		}
		simple.Sink = newFiles
	} else {
		log.Warnf("[%v]: Cannot assign new files, did not find *sample.FileSink as my direct output", b)
	}
	return simple
}

func (b *MultiFilePipelineBuilder) getFileSink(sink sample.MetricSink) *sample.FileSink {
	if files, ok := sink.(*sample.FileSink); ok {
		return files
	}
	if agg, ok := sink.(sample.AggregateSink); ok {
		var files *sample.FileSink
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

func SimpleMultiFileBuilder(buildPipeline func() []sample.SampleProcessor) *MultiFilePipelineBuilder {
	builder := &MultiFilePipelineBuilder{
		Description: "files suffixed with subpipeline key",
		NewFile: func(oldFile string, key interface{}) string {
			suffix := fmt.Sprintf("%v", key)
			group := sample.NewFileGroup(oldFile)
			return group.BuildFilenameStr(suffix)
		},
	}
	builder.Build = buildPipeline
	return builder
}
