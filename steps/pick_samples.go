package steps

import (
	"fmt"
	"strconv"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterPickPercent(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("pick",
		func(p *bitflow.SamplePipeline, params map[string]string) error {
			pick_percentage, err := strconv.ParseFloat(params["percent"], 64)
			if err != nil {
				return reg.ParameterError("percent", err)
			}
			counter := float64(0)
			p.Add(&SampleFilter{
				Description: bitflow.String(fmt.Sprintf("Pick %.2f%%", pick_percentage*100)),
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
		"Forward only a percentage of samples, parameter is in the range 0..1", reg.OptionalParams("percent"))
}

func RegisterPickHead(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("head",
		func(p *bitflow.SamplePipeline, params map[string]string) (err error) {
			doClose := reg.BoolParam(params, "close", false, true, &err)
			num := reg.IntParam(params, "num", 0, false, &err)
			if err == nil {
				processed := 0
				proc := &bitflow.SimpleProcessor{
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
		"Forward only a number of the first processed samples. The whole pipeline is closed afterwards, unless close=false is given.", reg.RequiredParams("num"), reg.OptionalParams("close"))
}

func RegisterSkipHead(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("skip",
		func(p *bitflow.SamplePipeline, params map[string]string) (err error) {
			num := reg.IntParam(params, "num", 0, false, &err)
			if err == nil {
				dropped := 0
				p.Add(&bitflow.SimpleProcessor{
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
		"Drop a number of samples in the beginning", reg.OptionalParams("num"))
}
