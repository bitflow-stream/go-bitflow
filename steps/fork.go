package steps

import (
	"fmt"
	"strconv"

	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/fork"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	log "github.com/sirupsen/logrus"
)

const DefaultForkKey = "*"

// This function is placed in this package to avoid circular dependency between the fork and the query package.
func RegisterForks(b *query.PipelineBuilder) {
	b.RegisterFork("rr", fork_round_robin, "The round-robin fork distributes the samples to the subpipelines based on weights. The pipeline selector keys must be positive integers denoting the weight of the respective pipeline.", []string{})
	b.RegisterFork("fork_tags", fork_tags, "Placeholders like ${xxx} are replaced by tag values. The resulting string forms the fork key", []string{"template"})
}

func fork_round_robin(subpiplines []query.Subpipeline, params map[string]string) (fork.ForkDistributor, error) {
	res := new(fork.RoundRobinDistributor)
	res.Weights = make([]int, len(subpiplines))
	res.Subpipelines = make([]*pipeline.SamplePipeline, len(subpiplines))
	for i, pipe := range subpiplines {
		weightSum := 0
		for _, keyStr := range pipe.Keys {
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
		pipeline, err := pipe.Build()
		if err != nil {
			return nil, err
		}
		res.Subpipelines[i] = pipeline
	}
	return res, nil
}

func fork_tags(subpiplines []query.Subpipeline, params map[string]string) (fork.ForkDistributor, error) {
	// TODO use a more generic version of fork.CachedDistributor (should be renamed to RegexDistributor)

	var initialKeys []string
	keys := make(map[string]*query.Subpipeline)
	for _, pipe := range subpiplines {
		for _, key := range pipe.Keys {
			if _, ok := keys[key]; ok {
				return nil, fmt.Errorf("Subpipeline key occurs multiple times: %v", key)
			}
			keys[key] = &pipe
			initialKeys = append(initialKeys, key)
		}
	}
	defaultPipe, haveDefault := keys[DefaultForkKey]

	dist := new(fork.TagDistributor)
	dist.Template = params["template"]
	dist.Build = func(key string) (*pipeline.SamplePipeline, error) {
		pipe, found := keys[key]
		if found {
			return pipe.Build()
		} else {
			if haveDefault {
				log.Debugf("No subpipeline defined for key '%v'. Building default pipeline (Have pipelines: %v)", key, keys)
				return defaultPipe.Build()
			} else {
				log.Warnf("No subpipeline defined for key '%v'. Using empty pipeline (Have pipelines: %v)", key, keys)
				return new(pipeline.SamplePipeline), nil
			}
		}
	}
	return dist, dist.EnsurePipelines(initialKeys)
}
