package fork

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/ryanuber/go-glob"
)

//
// TODO RegexDistributor:
// TODO allow controlling pipeline instances: per full pipeline definition, per (wildcard) pattern, per specific key, per sample
//

type PipelineArray struct {
	Subpipelines []*bitflow.SamplePipeline
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
	Weights []int // Optionally define weights for the pipelines (same order as pipelines). Only values >= 1 will be counted. Default weight is 1.

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
	return fmt.Sprintf("round robin (%v pipelines, total weight %v)", len(rr.Subpipelines), rr.TotalWeight())
}

func (rr *RoundRobinDistributor) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, len(rr.Subpipelines))
	total := rr.TotalWeight()
	for i, pipe := range rr.Subpipelines {
		weight := rr.getWeight(i)
		res[i] = &bitflow.TitledSamplePipeline{
			SamplePipeline: pipe,
			Title:          fmt.Sprintf("weight %v (%.2v%%)", weight, float64(weight)/float64(total)),
		}
	}
	return res
}

func (rr *RoundRobinDistributor) TotalWeight() (res int) {
	for i := range rr.Subpipelines {
		res += rr.getWeight(i)
	}
	return
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

func (d *MultiplexDistributor) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, len(d.Subpipelines))
	for i, pipe := range d.Subpipelines {
		res[i] = &bitflow.TitledSamplePipeline{
			SamplePipeline: pipe,
			Title:          fmt.Sprintf("Pipeline %v", i),
		}
	}
	return res
}

type PipelineBuildFunc func(key string) ([]*bitflow.SamplePipeline, error)

type PipelineCache struct {
	pipelines map[string][]*bitflow.SamplePipeline
	keys      map[*bitflow.SamplePipeline][]string
}

func (d *PipelineCache) getPipelines(key string, build PipelineBuildFunc) ([]Subpipeline, error) {
	if d.pipelines == nil {
		d.pipelines = make(map[string][]*bitflow.SamplePipeline)
	}
	if d.keys == nil {
		d.keys = make(map[*bitflow.SamplePipeline][]string)
	}
	pipes, ok := d.pipelines[key]
	if !ok {
		if build != nil {
			var err error
			pipes, err = build(key)
			if err != nil {
				return nil, err
			}
		}
		d.pipelines[key] = pipes
	}
	d.updateSortedPipelineKeys(key, pipes)
	result := make([]Subpipeline, len(pipes))
	for i, pipe := range pipes {
		result[i].Key = key
		result[i].Pipe = pipe
	}
	return result, nil
}

func (d *PipelineCache) updateSortedPipelineKeys(key string, pipes []*bitflow.SamplePipeline) {
	// Maintain a sorted list of keys that lead to each pipeline
	for _, pipe := range pipes {
		if keys, ok := d.keys[pipe]; ok {
			index := sort.SearchStrings(keys, key)
			if index < len(keys) && keys[index] == key {
				// Key already present
			} else {
				// Insert the key and keep the key slice sorted
				keys = append(keys, key)
				sort.Strings(keys)
				d.keys[pipe] = keys
			}
		} else {
			d.keys[pipe] = []string{key}
		}
	}
}

func (d *PipelineCache) ContainedStringers() []fmt.Stringer {
	res := make([]fmt.Stringer, 0, len(d.keys))
	for pipe, keys := range d.keys {
		var keyStr string
		if len(keys) == 1 {
			keyStr = "'" + keys[0] + "'"
		} else {
			keyStr = "['" + strings.Join(keys, "', '") + "']"
		}
		res = append(res, &bitflow.TitledSamplePipeline{
			Title:          "Pipeline " + keyStr,
			SamplePipeline: pipe,
		})
	}
	return res
}

type RegexDistributor struct {
	Pipelines map[string]func() ([]*bitflow.SamplePipeline, error)

	ExactMatch bool // Key patterns must match exactly, no glob (*) processing
	RegexMatch bool // Overrides ExactMatch -> treat key patterns as regexes

	regexCache        map[string]*regexp.Regexp
	cache             PipelineCache
	wildcardPipelines PipelineCache // This extra cache is only for implementing ContainedStringers()
}

func (d *RegexDistributor) Init() error {
	// Initialize the pipeline cache used for ContainedStringers(). Also report early errors.
	for key := range d.Pipelines {
		_, err := d.wildcardPipelines.getPipelines(key, func(key string) ([]*bitflow.SamplePipeline, error) {
			// Strictly build the pipelines for the available keys
			return d.doBuild(key, false, false)
		})
		if err != nil {
			return err
		}
	}
	if d.RegexMatch {
		d.regexCache = make(map[string]*regexp.Regexp)
		for key := range d.Pipelines {
			regex, err := regexp.Compile(key)
			if err != nil {
				return err
			}
			d.regexCache[key] = regex
		}
	}
	return nil
}

func (d *RegexDistributor) getPipelines(key string) ([]Subpipeline, error) {
	return d.cache.getPipelines(key, d.build)
}

func (d *RegexDistributor) build(key string) ([]*bitflow.SamplePipeline, error) {
	return d.doBuild(key, d.RegexMatch, !d.ExactMatch)
}

func (d *RegexDistributor) doBuild(key string, allowRegex bool, allowGlob bool) ([]*bitflow.SamplePipeline, error) {
	var res []*bitflow.SamplePipeline
	for wildcardKey, builderFunc := range d.Pipelines {
		if d.matches(key, wildcardKey, allowRegex, allowGlob) {
			newPipelines, err := builderFunc()
			if err != nil {
				return res, err
			}
			res = append(res, newPipelines...)
		}
	}
	return res, nil
}

func (d *RegexDistributor) matches(key, pattern string, allowRegex bool, allowGlob bool) bool {
	if allowRegex {
		regex := d.regexCache[pattern]
		return regex.MatchString(key)
	} else if allowGlob {
		return glob.Glob(pattern, key)
	} else {
		return key == pattern
	}
}

func (d *RegexDistributor) ContainedStringers() []fmt.Stringer {
	return d.wildcardPipelines.ContainedStringers()
}

type GenericDistributor struct {
	RegexDistributor
	GetKeys     func(sample *bitflow.Sample, header *bitflow.Header) []string
	Description string
}

func (d *GenericDistributor) Distribute(sample *bitflow.Sample, header *bitflow.Header) ([]Subpipeline, error) {
	keys := d.GetKeys(sample, header)
	res := make([]Subpipeline, 0, len(keys)) // Preallocated capacity is just a heuristic
	for _, key := range keys {
		newPipes, err := d.getPipelines(key)
		if err != nil {
			return nil, err
		}
		for _, pipe := range newPipes {
			res = append(res, pipe)
		}
	}
	return res, nil
}

func (d *GenericDistributor) String() string {
	return d.Description
}

type TagDistributor struct {
	RegexDistributor
	bitflow.TagTemplate
}

func (d *TagDistributor) Distribute(sample *bitflow.Sample, _ *bitflow.Header) ([]Subpipeline, error) {
	return d.getPipelines(d.Resolve(sample))
}

func (d *TagDistributor) String() string {
	matchMode := "glob"
	if d.RegexMatch {
		matchMode = "regex"
	} else if d.ExactMatch {
		matchMode = "exact"
	}
	return fmt.Sprintf("tag template (%v matching): %v", matchMode, d.Template)
}

var _ Distributor = new(MultiFileDistributor)

type MultiFileDistributor struct {
	bitflow.TagTemplate
	PipelineCache

	Endpoints          *bitflow.EndpointFactory
	Config             bitflow.FileSink // Configuration parameters in this field will be used for file outputs
	ExtendSubPipelines func(fileName string, pipe *bitflow.SamplePipeline)
}

func MakeMultiFilePipelineBuilder(endpointParams map[string]string, endpoints *bitflow.EndpointFactory) (*MultiFileDistributor, error) {
	endpoints, err := endpoints.CloneWithParams(endpointParams)
	if err != nil {
		return nil, fmt.Errorf("Error parsing parameters: %v", err)
	}
	output, err := endpoints.CreateOutput("file://-") // Create empty file output, will only be used as template with configuration values
	if err != nil {
		return nil, fmt.Errorf("Error creating template file output: %v", err)
	}
	fileOutput, ok := output.(*bitflow.FileSink)
	if !ok {
		return nil, fmt.Errorf("Error creating template file output, received wrong type: %T", output)
	}
	return &MultiFileDistributor{Config: *fileOutput, Endpoints: endpoints}, nil
}

func (b *MultiFileDistributor) Distribute(sample *bitflow.Sample, _ *bitflow.Header) ([]Subpipeline, error) {
	return b.getPipelines(b.Resolve(sample), b.build)
}

func (b *MultiFileDistributor) String() string {
	return "Output to files: " + b.Template
}

func (b *MultiFileDistributor) build(fileName string) ([]*bitflow.SamplePipeline, error) {
	fileOut := b.Config
	fileOut.Filename = fileName
	format := bitflow.EndpointDescription{Target: fileName, Type: bitflow.FileEndpoint}.DefaultOutputFormat()
	var err error
	fileOut.Marshaller, err = b.Endpoints.CreateMarshaller(format)
	if err != nil {
		return nil, err
	}
	pipe := (new(bitflow.SamplePipeline)).Add(&fileOut)
	if extend := b.ExtendSubPipelines; extend != nil {
		extend(fileName, pipe)
	}
	return []*bitflow.SamplePipeline{pipe}, nil
}
