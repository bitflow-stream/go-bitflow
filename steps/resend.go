package steps

import (
	"fmt"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

func RegisterResendStep(b reg.ProcessorRegistry) {
	b.RegisterStep("resend",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			p.Add(&ResendProcessor{
				Interval: params["interval"].(time.Duration),
			})
			return nil
		},
		"If no new sample is received within the given period of time, resend a copy of it.").
		Required("interval", reg.Duration())
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

func RegisterFillUpStep(b reg.ProcessorRegistry) {
	b.RegisterStep("fill-up",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) (err error) {
			interval := params["interval"].(time.Duration)
			stepInterval := params["step-interval"].(time.Duration)
			if stepInterval == 0 {
				stepInterval = interval
			}
			p.Add(&FillUpProcessor{
				MinMissingInterval: interval,
				StepInterval:       stepInterval,
			})
			return
		},
		"If the timestamp different between two consecutive samples is larger than the given interval, send copies of the first sample to fill the gap").
		Required("interval", reg.Duration()).
		Optional("step-interval", reg.Duration(), time.Duration(0))
}

type FillUpProcessor struct {
	bitflow.NoopProcessor
	MinMissingInterval time.Duration
	StepInterval       time.Duration
	previous           *bitflow.Sample
}

func (p *FillUpProcessor) String() string {
	return fmt.Sprintf("Fill up samples missing for %v (in intervals of %v)", p.MinMissingInterval, p.StepInterval)
}

func (p *FillUpProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	if p.previous != nil && !p.previous.Time.Add(p.MinMissingInterval).After(sample.Time) {
		for t := p.previous.Time.Add(p.StepInterval); t.Before(sample.Time); t = t.Add(p.StepInterval) {
			clone := p.previous.DeepClone()
			clone.Time = t
			if err := p.NoopProcessor.Sample(clone, header); err != nil {
				return err
			}
		}
	}
	p.previous = sample.DeepClone() // Clone necessary because follow-up steps might modify the sample in-place
	return p.NoopProcessor.Sample(sample, header)
}
