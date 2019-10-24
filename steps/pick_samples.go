package steps

import (
	"fmt"
	"log"
	"strconv"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterPickPercent(b reg.ProcessorRegistry) {
	b.RegisterStep("pick",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			counter := float64(0)
			pick_percentage := params["percent"].(float64)
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
		"Forward only a percentage of samples, parameter is in the range 0..1").
		Required("percent", reg.Float())
}

func RegisterPickHead(b reg.ProcessorRegistry) {
	b.RegisterStep("head",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) (err error) {
			doClose := params["close"].(bool)
			num := params["num"].(int)
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
		"Forward only a number of the first processed samples. The whole pipeline is closed afterwards, unless close=false is given.").
		Required("num", reg.Int()).
		Optional("close", reg.Bool(), false)
}

func RegisterSkipHead(b reg.ProcessorRegistry) {
	b.RegisterStep("skip",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) (err error) {
			num := params["num"].(int)
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
		"Drop a number of samples in the beginning").
		Required("num", reg.Int())
}

func RegisterPickTail(b reg.ProcessorRegistry) {
	b.RegisterStep("tail",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) (err error) {
			num := params["num"].(int)
			if err == nil {
				ring := bitflow.NewSampleRing(num)
				proc := &bitflow.SimpleProcessor{
					Description: "Read until end of stream, and forward only the last " + strconv.Itoa(num) + " samples",
					Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
						ring.Push(sample, header)
						return nil, nil, nil
					},
				}
				proc.OnClose = func() {
					flush := ring.Get()
					log.Printf("%v: Reached end of stream, now flushing %v samples", proc, len(flush))
					for _, sample := range flush {
						if err := proc.NoopProcessor.Sample(sample.Sample, sample.Header); err != nil {
							proc.Error(err)
							break
						}
					}
				}
				p.Add(proc)
			}
			return
		},
		"Forward only a number of the first processed samples. The whole pipeline is closed afterwards, unless close=false is given.").
		Required("num", reg.Int())
}

func RegisterDropInvalid(b reg.ProcessorRegistry) {
	b.RegisterStep("drop-invalid",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			p.Add(&SampleFilter{
				Description: bitflow.String("Drop samples with invalid values (NaN/Inf)"),
				IncludeFilter: func(s *bitflow.Sample, _ *bitflow.Header) (bool, error) {
					for _, val := range s.Values {
						if !IsValidNumber(float64(val)) {
							return false, nil
						}
					}
					return true, nil
				},
			})
			return nil
		},
		"Drop samples that contain NaN or Inf values.")
}
