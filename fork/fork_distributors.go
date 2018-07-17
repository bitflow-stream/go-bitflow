package fork

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
)

type PipelineArray struct {
	Subpipelines []*pipeline.SamplePipeline
	out          []Subpipeline
}

func (p *PipelineArray) build() []Subpipeline {
	if len(p.out) != len(p.Subpipelines) {
		p.out = make([]Subpipeline, len(p.Subpipelines))
		for i, pipe := range p.Subpipelines {
			p.out[i] = Subpipeline{pipe, strconv.Itoa(i)}
		}
	}
	return p.out
}

type RoundRobinDistributor struct {
	PipelineArray
	Weights []int // Optionally define weights for the pielines (same order as pipelines). Only values >= 1 will be counted. Default weight is 1.

	nextPipe      int
	weightCounter int
}

func (rr *RoundRobinDistributor) Distribute(sample *bitflow.Sample, header *bitflow.Header) ([]Subpipeline, error) {
	index := rr.nextPipe % len(rr.Subpipelines)
	weight := rr.getWeight(index)
	rr.weightCounter++
	if rr.weightCounter >= weight {
		rr.nextPipe++
		rr.weightCounter = 0
	}
	return rr.build()[index : index+1], nil
}

func (rr *RoundRobinDistributor) getWeight(index int) int {
	weight := 1
	if len(rr.Weights) > index && rr.Weights[index] > 0 {
		weight = rr.Weights[index]
	}
	return weight
}

func (rr *RoundRobinDistributor) String() string {
	return fmt.Sprintf("round robin (%v)", len(rr.Subpipelines))
}

func (rr *RoundRobinDistributor) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, len(rr.Subpipelines))
	for i, pipe := range rr.Subpipelines {
		res[i] = &pipeline.TitledSamplePipeline{
			SamplePipeline: pipe,
			Title:          fmt.Sprintf("%v (weight %v)", i, rr.getWeight(i)),
		}
	}
	return res
}

type MultiplexDistributor struct {
	PipelineArray
}

func (d *MultiplexDistributor) Distribute(_ *bitflow.Sample, _ *bitflow.Header) ([]Subpipeline, error) {
	return d.build(), nil
}

func (d *MultiplexDistributor) String() string {
	return fmt.Sprintf("multiplex (%v)", len(d.Subpipelines))
}

func (p *MultiplexDistributor) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, len(p.Subpipelines))
	for i, pipe := range p.Subpipelines {
		res[i] = &pipeline.TitledSamplePipeline{
			SamplePipeline: pipe,
			Title:          fmt.Sprintf("Pipeline %v", i),
		}
	}
	return res
}

type PipelineBuildFunc func(key string) (*pipeline.SamplePipeline, error)

type PipelineCache struct {
	pipelines map[string]*pipeline.SamplePipeline
	keys      map[*pipeline.SamplePipeline][]string
}

func (d *PipelineCache) getPipeline(key string, build PipelineBuildFunc) (Subpipeline, error) {

	//
	// TODO: extend to handle regexes and allow the same key multiple times
	// Also, make possible to share pipelines between keys, controlled
	//

	if d.pipelines == nil {
		d.pipelines = make(map[string]*pipeline.SamplePipeline)
	}
	if d.keys == nil {
		d.keys = make(map[*pipeline.SamplePipeline][]string)
	}
	pipe, ok := d.pipelines[key]
	if !ok {
		if build == nil {
			pipe = new(pipeline.SamplePipeline)
		} else {
			var err error
			pipe, err = build(key)
			if err == nil && pipe == nil {
				err = fmt.Errorf("Build function returned nil pipeline")
			}
			if err != nil {
				return Subpipeline{}, err
			}
		}
		d.pipelines[key] = pipe
	}
	if keys, ok := d.keys[pipe]; ok {
		index := sort.SearchStrings(keys, key)
		if index < len(keys) && keys[index] == key {
			// Key already present
		} else {
			// Insert the key and keep the key slice sorted
			keys = append(keys, "")
			if index <= len(keys) {
				copy(keys[index+1:], keys[index:])
			}
			keys[index] = key
			d.keys[pipe] = keys
		}
	} else {
		d.keys[pipe] = []string{key}
	}
	return Subpipeline{pipe, key}, nil
}

func (d *PipelineCache) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, 0, len(d.keys))
	for pipe, keys := range d.keys {
		var keyStr string
		if len(keys) == 1 {
			keyStr = keys[0]
		} else {
			keyStr = "[" + strings.Join(keys, ", ") + "]"
		}
		res = append(res, &pipeline.TitledSamplePipeline{
			Title:          "Pipeline " + keyStr,
			SamplePipeline: pipe,
		})
	}
	return res
}

type CachingDistributor struct {
	PipelineCache
	Build func(key string) (*pipeline.SamplePipeline, error)
}

func (d *CachingDistributor) getPipeline(key string) (Subpipeline, error) {
	return d.PipelineCache.getPipeline(key, d.Build)
}

func (d *CachingDistributor) EnsurePipelines(keys []string) error {
	for _, key := range keys {
		if _, err := d.getPipeline(key); err != nil {
			return err
		}
	}
	return nil
}

type GenericDistributor struct {
	CachingDistributor
	GetKeys     func(sample *bitflow.Sample, header *bitflow.Header) []string
	Description string
}

func (d *GenericDistributor) Distribute(sample *bitflow.Sample, header *bitflow.Header) ([]Subpipeline, error) {
	keys := d.GetKeys(sample, header)
	res := make([]Subpipeline, len(keys))
	for i, key := range keys {
		var err error
		res[i], err = d.getPipeline(key)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *GenericDistributor) String() string {
	return d.Description
}

type TagTemplate struct {
	Template     string // Placeholders like ${xxx} will be replaced by tag values (left empty if tag missing)
	MissingValue string // Replacement for missing values
}

var templateRegex = regexp.MustCompile("\\$\\{[^\\{]*\\}") // Example: ${hello}

func (t *TagTemplate) BuildKey(sample *bitflow.Sample) string {
	return templateRegex.ReplaceAllStringFunc(t.Template, func(placeholder string) string {
		placeholder = placeholder[2 : len(placeholder)-1] // Strip the ${} prefix/suffix
		if sample.HasTag(placeholder) {
			return sample.Tag(placeholder)
		} else {
			return t.MissingValue
		}
	})
}

type TagDistributor struct {
	CachingDistributor
	TagTemplate
}

func (d *TagDistributor) Distribute(sample *bitflow.Sample, _ *bitflow.Header) ([]Subpipeline, error) {
	key := d.BuildKey(sample)
	res, err := d.getPipeline(key)
	return []Subpipeline{res}, err
}

func (d *TagDistributor) String() string {
	return fmt.Sprintf("tag template: %v", d.Template)
}

var _ ForkDistributor = new(MultiFileDistributor)

type MultiFileDistributor struct {
	TagTemplate
	PipelineCache
	Config bitflow.FileSink // Configuration parameters in this field will be used for file outputs
}

func (b *MultiFileDistributor) Distribute(sample *bitflow.Sample, _ *bitflow.Header) ([]Subpipeline, error) {
	fileName := b.BuildKey(sample)
	res, err := b.getPipeline(fileName, b.build)
	return []Subpipeline{res}, err
}

func (b *MultiFileDistributor) String() string {
	return "Output to files: " + b.Template
}

func (b *MultiFileDistributor) build(fileName string) (*pipeline.SamplePipeline, error) {
	fileOut := b.Config
	fileOut.Filename = fileName
	fileOut.Marshaller = bitflow.EndpointDescription{Target: fileName, Type: bitflow.FileEndpoint}.DefaultOutputFormat().Marshaller()
	return (new(pipeline.SamplePipeline)).Add(&fileOut), nil
}

/*









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

func (b *SimplePipelineBuilder) BuildPipeline(key interface{}, output *ForkMerger) *pipeline.SamplePipeline {
	var res pipeline.SamplePipeline
	if b.Build != nil {
		for _, processor := range b.Build() {
			res.Add(processor)
		}
	}
	res.Add(output)
	return &res
}





type extendedStringPipelineBuilder struct {
	fork.StringPipelineBuilder
	builder                  *PipelineBuilder
	defaultTail              Pipeline
	defaultPipeline          *pipeline.SamplePipeline
	singletonDefaultPipeline *pipeline.SamplePipeline
}

func (b *extendedStringPipelineBuilder) ContainedStringers() []fmt.Stringer {
	res := b.StringPipelineBuilder.ContainedStringers()
	var title string
	var pipe *pipeline.SamplePipeline
	if b.singletonDefaultPipeline != nil {
		title = "Default pipeline"
		pipe = b.defaultPipeline
	} else if b.defaultPipeline != nil {
		title = "Singleton default pipeline"
		pipe = b.singletonDefaultPipeline
	}
	if pipe != nil {
		res = append([]fmt.Stringer{&pipeline.TitledSamplePipeline{Title: title, SamplePipeline: pipe}}, res...)
	}
	return res
}

func (b *extendedStringPipelineBuilder) buildMissing(string) (res *pipeline.SamplePipeline, err error) {
	if b.singletonDefaultPipeline != nil {
		res = b.singletonDefaultPipeline // Use the same pipelien for every fork key
	} else if b.defaultPipeline != nil {
		res, err = b.builder.makePipelineTail(b.defaultTail) // Build new pipeline for every fork key
	} else {
		res = new(pipeline.SamplePipeline)
	}
	return
}

func make() {
	for _, input := range inputs {
		key := input.Content()
		if _, ok := res.Pipelines[key]; ok {
			return nil, fmt.Errorf("Subpipeline key '%v' defined multiple times", key)
		}
		if key == DefaultForkKey {
			res.defaultTail = pipe[1:]
			res.defaultPipeline = builtPipe
		} else if key == SingletonDefaultForkKey {
			res.singletonDefaultPipeline = builtPipe
		} else {
			res.Pipelines[key] = builtPipe
		}
	}
	if res.defaultPipeline != nil && res.singletonDefaultPipeline != nil {
		return nil, fmt.Errorf("Cannot have both singleton and individual default subpipelines (fork keys '%v' and '%v')", DefaultForkKey, SingletonDefaultForkKey)
	}
}

*/
