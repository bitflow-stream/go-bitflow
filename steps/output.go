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
			distributor.ExtendSubpipelines = func(_ string, pipe *bitflow.SamplePipeline) {
				pipe.Add(&DecouplingProcessor{ChannelBuffer: parallelize})
			}
		}
	}

	b.RegisterStep("output_files",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			distributor, err := makeMultiFilePipelineBuilder(params["endpoint-config"].(map[string]string))
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

		distributor, err := makeMultiFilePipelineBuilder(params)
		if err != nil {
			return nil, err
		}
		distributor.Template = url.Path
		addParallelization(parallelize, distributor)
		return &fork.SampleFork{Distributor: distributor}, nil
	}
}

func makeMultiFilePipelineBuilder(params map[string]string) (*fork.MultiFileDistributor, error) {
	if err := bitflow.DefaultEndpointFactory.ParseParameters(params); err != nil {
		return nil, fmt.Errorf("Error parsing parameters: %v", err)
	}
	output, err := bitflow.DefaultEndpointFactory.CreateOutput("file://-") // Create empty file output, will only be used as template with configuration values
	if err != nil {
		return nil, fmt.Errorf("Error creating template file output: %v", err)
	}
	fileOutput, ok := output.(*bitflow.FileSink)
	if !ok {
		return nil, fmt.Errorf("Error creating template file output, received wrong type: %T", output)
	}
	return &fork.MultiFileDistributor{Config: *fileOutput}, nil
}
