package steps

import (
	"fmt"
	"strconv"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/bitflow/fork"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterOutputFiles(b reg.ProcessorRegistry) {
	addParallelization := func(parallelize int, distributor *fork.MultiFileDistributor) {
		if parallelize > 0 {
			distributor.ExtendSubPipelines = func(_ string, pipe *bitflow.SamplePipeline) {
				pipe.Add(&DecouplingProcessor{ChannelBuffer: parallelize})
			}
		}
	}

	b.RegisterStep("output_files",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			distributor, err := fork.MakeMultiFilePipelineBuilder(params["endpoint-config"].(map[string]string), b.Endpoints)
			if err == nil {
				distributor.Template = params["file"].(string)
				addParallelization(params["parallelize"].(int), distributor)
				p.Add(&fork.SampleFork{Distributor: distributor})
			}
			return err
		}, "Output samples to multiple files, filenames are built from the given template, where placeholders like ${xxx} will be replaced with tag values").
		Required("file", reg.String()).
		Optional("parallelize", reg.Int(), 0).
		Optional("endpoint-config", reg.Map(reg.String()), map[string]string{})

	b.Endpoints.CustomDataSinks["files"] = func(urlTarget string) (bitflow.SampleProcessor, error) {
		url, err := reg.ParseEndpointFilepath(urlTarget)
		if err != nil {
			return nil, err
		}
		params, err := reg.ParseQueryParameters(url)
		if err != nil {
			return nil, err
		}

		parallelize := 0
		if parallelizeStr, ok := params["parallelize"]; ok {
			parallelize, err = strconv.Atoi(parallelizeStr)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse 'parallellize' parameter to int: %v", err)
			}
		}
		delete(params, "parallelize")

		distributor, err := fork.MakeMultiFilePipelineBuilder(params, b.Endpoints)
		if err != nil {
			return nil, err
		}
		distributor.Template = url.Path
		addParallelization(parallelize, distributor)
		return &fork.SampleFork{Distributor: distributor}, nil
	}
}
