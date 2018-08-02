package steps

import (
	"fmt"
	"strconv"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/builder"
)

func RegisterPickPercent(b builder.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("pick",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			pick_percentage, err := strconv.ParseFloat(params["percent"], 64)
			if err != nil {
				return builder.ParameterError("percent", err)
			}
			counter := float64(0)
			p.Add(&SampleFilter{
				Description: pipeline.String(fmt.Sprintf("Pick %.2f%%", pick_percentage*100)),
				IncludeFilter: func(_ *bitflow.Sample, _ *bitflow.Header) (bool, error) {
					counter += pick_percentage
					if counter > 1.0 {
						counter -= 1.0
						return true, nil
					}
					return false, nil
				},
			})
			return nil
		},
		"Forward only a percentage of samples, parameter is in the range 0..1", builder.OptionalParams("percent"))
}

func RegisterPickHead(b builder.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("head",
		func(p *pipeline.SamplePipeline, params map[string]string) (err error) {
			doClose := builder.BoolParam(params, "close", false, true, &err)
			num := builder.IntParam(params, "num", 0, false, &err)
			if err == nil {
				processed := 0
				proc := &pipeline.SimpleProcessor{
					Description: "Pick first " + strconv.Itoa(num) + " samples",
				}
				proc.Process = func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
					if num > processed {
						processed++
						return sample, header, nil
					} else {
						if doClose {
							proc.Error(nil) // Stop processing without an error
						}
						return nil, nil, nil
					}
				}
				p.Add(proc)
			}
			return
		},
		"Forward only a number of the first processed samples. The whole pipeline is closed afterwards, unless close=false is given.", builder.RequiredParams("num"), builder.OptionalParams("close"))
}

func RegisterSkipHead(b builder.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("skip",
		func(p *pipeline.SamplePipeline, params map[string]string) (err error) {
			num := builder.IntParam(params, "num", 0, false, &err)
			if err == nil {
				dropped := 0
				p.Add(&pipeline.SimpleProcessor{
					Description: "Drop first " + strconv.Itoa(num) + " samples",
					Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
						if dropped >= num {
							return sample, header, nil
						} else {
							dropped++
							return nil, nil, nil
						}
					},
				})
			}
			return
		},
		"Drop a number of samples in the beginning", builder.OptionalParams("num"))
}
