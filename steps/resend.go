package steps

import (
	"fmt"
	"sync"
	"time"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

func RegisterResendStep(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("resend",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			interval, err := time.ParseDuration(params["interval"])
			if err != nil {
				return query.ParameterError("interval", err)
			}
			p.Add(&ResendProcessor{
				Interval: interval,
			})
			return nil
		},
		"If no new sample is received within the given period of time, resend a copy of it.", []string{"interval"})
}

type ResendProcessor struct {
	bitflow.NoopProcessor
	Interval time.Duration

	wg          *sync.WaitGroup
	currentLoop golib.StopChan
}

func (p *ResendProcessor) String() string {
	return fmt.Sprintf("Resend every %v", p.Interval)
}

func (p *ResendProcessor) Start(wg *sync.WaitGroup) golib.StopChan {
	if p.Interval < 0 {
		p.Interval = 0
	}
	p.wg = wg
	return p.NoopProcessor.Start(wg)
}

func (p *ResendProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.currentLoop.Stop()
	err := p.NoopProcessor.Sample(sample, header)
	if err == nil {
		p.currentLoop = SendPeriodically(sample, header, p.GetSink(), p.Interval, p.wg)
	}
	return err
}

func (p *ResendProcessor) Close() {
	p.currentLoop.Stop()
	p.NoopProcessor.Close()
}

func SendPeriodically(sample *bitflow.Sample, header *bitflow.Header, receiver bitflow.SampleSink, interval time.Duration, wg *sync.WaitGroup) golib.StopChan {
	stopper := golib.NewStopChan()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for stopper.WaitTimeout(interval) {
			stopper.IfNotStopped(func() {
				err := receiver.Sample(sample.DeepClone(), header.Clone(header.Fields))
				if err != nil {
					log.Errorf("Error periodically resending sample (interval %v) (%v metric(s), tags: %v): %v",
						interval, len(header.Fields), sample.TagString(), err)
				}
			})
		}
	}()
	return stopper
}
