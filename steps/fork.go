package steps

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/fork"
	"github.com/antongulenko/go-bitflow-pipeline/query"
)

// This function is placed in this package to avoid circular dependency between the fork and the query package.
func RegisterForks(b *query.PipelineBuilder) {
	b.RegisterFork("rr", fork_round_robin, "The round-robin fork distributes the samples to the subpipelines based on weights. The pipeline selector keys must be positive integers denoting the weight of the respective pipeline.", []string{})
	b.RegisterFork("fork_tag", fork_tag, "Fork based on the values of the given tag", []string{"tag"})
	b.RegisterFork("fork_tag_template", fork_tag_template, "Fork based on a template string, placeholders like ${xxx} are replaced by tag values.", []string{"template"})
}

func fork_round_robin(subpipelines []query.Subpipeline, _ map[string]string) (fork.Distributor, error) {
	res := new(fork.RoundRobinDistributor)
	res.Weights = make([]int, len(subpipelines))
	res.Subpipelines = make([]*pipeline.SamplePipeline, len(subpipelines))
	for i, subpipeAST := range subpipelines {
		weightSum := 0
		for _, keyStr := range subpipeAST.Keys {
			weight, err := strconv.Atoi(keyStr)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse Round Robin subpipeline key '%v' to integer: %v", keyStr, err)
			}
			if weight <= 0 {
				return nil, fmt.Errorf("Round robin subpipeline keys must be positive (wrong key: %v)", weight)
			}
			weightSum += weight
		}
		res.Weights[i] = weightSum
		subpipe, err := subpipeAST.Build()
		if err != nil {
			return nil, err
		}
		res.Subpipelines[i] = subpipe
	}
	return res, nil
}

func fork_tag(subpipelines []query.Subpipeline, params map[string]string) (fork.Distributor, error) {
	tag := params["tag"]
	delete(params, "tag")
	params["template"] = "${" + tag + "}"
	return fork_tag_template(subpipelines, params)
}

func fork_tag_template(subpipelines []query.Subpipeline, params map[string]string) (fork.Distributor, error) {
	wildcardPipelines := make(map[string]func() ([]*pipeline.SamplePipeline, error))
	var keysArray []string
	for _, pipe := range subpipelines {
		for _, key := range pipe.Keys {
			if _, ok := wildcardPipelines[key]; ok {
				return nil, fmt.Errorf("Subpipeline key occurs multiple times: %v", key)
			}
			wildcardPipelines[key] = (&wildcardSubpipeline{p: pipe}).build
			keysArray = append(keysArray, key)
		}
	}
	sort.Strings(keysArray)

	dist := new(fork.TagDistributor)
	dist.Template = params["template"]
	dist.Pipelines = wildcardPipelines
	return dist, dist.Init()
}

type wildcardSubpipeline struct {
	p query.Subpipeline
}

func (m wildcardSubpipeline) build() ([]*pipeline.SamplePipeline, error) {
	pipe, err := m.p.Build()
	return []*pipeline.SamplePipeline{pipe}, err
}
