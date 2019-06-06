package steps

import (
	"fmt"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/bitflow/fork"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterOutputFiles(b reg.ProcessorRegistry) {
	create := func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
		distributor, err := makeMultiFilePipelineBuilder(params["endpoint-config"].(map[string]string))
		if err == nil {
			distributor.Template = params["file"].(string)
			parallelize := params["parallelize"].(int)
			if parallelize > 0 {
				distributor.ExtendSubpipelines = func(_ string, pipe *bitflow.SamplePipeline) {
					pipe.Add(&DecouplingProcessor{ChannelBuffer: parallelize})
				}
			}
			p.Add(&fork.SampleFork{Distributor: distributor})
		}
		return err
	}

	b.RegisterStep("output_files", create, "Output samples to multiple files, filenames are built from the given template, where placeholders like ${xxx} will be replaced with tag values").
		Required("file", reg.String()).
		Optional("parallelize", reg.Int(), 0).
		Optional("endpoint-config", reg.Map(reg.String()), map[string]string{})

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
