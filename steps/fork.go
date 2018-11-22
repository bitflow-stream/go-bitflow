package steps

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/fork"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
)

// This function is placed in this package to avoid circular dependency between the fork and the query package.
func RegisterForks(b reg.ProcessorRegistry) {
	b.RegisterFork("rr", fork_round_robin, "The round-robin fork distributes the samples to the subpipelines based on weights. The pipeline selector keys must be positive integers denoting the weight of the respective pipeline.")
	b.RegisterFork("fork_tag", fork_tag, "Fork based on the values of the given tag", reg.RequiredParams("tag"), reg.OptionalParams("regex", "exact"))
	b.RegisterFork("fork_tag_template", fork_tag_template, "Fork based on a template string, placeholders like ${xxx} are replaced by tag values.", reg.RequiredParams("template"), reg.OptionalParams("regex", "exact"))
}

func fork_round_robin(subpipelines []reg.Subpipeline, _ map[string]string) (fork.Distributor, error) {
	res := new(fork.RoundRobinDistributor)
	res.Weights = make([]int, len(subpipelines))
	res.Subpipelines = make([]*pipeline.SamplePipeline, len(subpipelines))
	for i, subpipeAST := range subpipelines {
		weightSum := 0
		for _, keyStr := range subpipeAST.Keys() {

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

func fork_tag(subpipelines []reg.Subpipeline, params map[string]string) (fork.Distributor, error) {
	tag := params["tag"]
	delete(params, "tag")
	params["template"] = "${" + tag + "}"
	return fork_tag_template(subpipelines, params)
}

func fork_tag_template(subpipelines []reg.Subpipeline, params map[string]string) (fork.Distributor, error) {
	wildcardPipelines := make(map[string]func() ([]*pipeline.SamplePipeline, error))
	var keysArray []string
	for _, pipe := range subpipelines {
		for _, key := range pipe.Keys() {
			if _, ok := wildcardPipelines[key]; ok {
				return nil, fmt.Errorf("Subpipeline key occurs multiple times: %v", key)
			}
			wildcardPipelines[key] = (&wildcardSubpipeline{p: pipe}).build
			keysArray = append(keysArray, key)
		}
	}
	sort.Strings(keysArray)

	var err error
	dist := &fork.TagDistributor{
		TagTemplate: pipeline.TagTemplate{
			Template: params["template"],
		},
		RegexDistributor: fork.RegexDistributor{
			Pipelines:  wildcardPipelines,
			ExactMatch: reg.BoolParam(params, "exact", false, true, &err),
			RegexMatch: reg.BoolParam(params, "regex", false, true, &err),
		},
	}
	if err == nil {
		err = dist.Init()
	}
	return dist, err
}

type wildcardSubpipeline struct {
	p reg.Subpipeline
}

func (m wildcardSubpipeline) build() ([]*pipeline.SamplePipeline, error) {
	pipe, err := m.p.Build()
	return []*pipeline.SamplePipeline{pipe}, err
}
