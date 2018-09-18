package plugin

import (
	"sync"

	"github.com/antongulenko/go-bitflow"
	"github.com/antongulenko/golib"
)

type SampleSourceFactory func(map[string]string) (bitflow.SampleSource, error)

type SampleSourcePluginImplementation struct {
	Create SampleSourceFactory

	wg     sync.WaitGroup
	source bitflow.SampleSource
	sink   DataSink
}

func (s *SampleSourcePluginImplementation) Start(params map[string]string, dataSink DataSink) {
	source, err := s.Create(params)
	if err != nil {
		dataSink.Error(err)
		return
	}
	s.sink = dataSink
	s.source = source
	s.source.SetSink(&pluginSinkAdapter{sink: dataSink, wg: &s.wg})

	s.wg.Add(1) // Will be undone in pluginSinkAdapter.Close()
	go func() {
		// Wait for the SampleSource implementation to finish completely, before forwarding the Close() call
		// This will only return after pluginSinkAdapter.Close() was called.
		s.wg.Wait()
		dataSink.Close()
	}()

	stopper := s.source.Start(&s.wg)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		stopper.Wait()
		if err := stopper.Err(); err != nil {
			dataSink.Error(err)
		}
		// Ignore nil-value of StopChan: imitate behavior of golib.WaitForAny
		if !stopper.IsNil() {
			s.Close()
		}
	}()
}

func (s *SampleSourcePluginImplementation) Close() {
	if s.source != nil {
		s.source.Close()
	}
}

type pluginSinkAdapter struct {
	bitflow.AbstractSampleProcessor
	sink DataSink
	wg   *sync.WaitGroup
}

func (p *pluginSinkAdapter) Start(_ *sync.WaitGroup) (_ golib.StopChan) {
	// This should never be called, since this implementating of SampleProcessor
	// only forwards the Sample() and Close() invokations to the DataSink
	return
}

func (p *pluginSinkAdapter) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.sink.Sample(sample, header)
	return nil
}

func (p *pluginSinkAdapter) Close() {
	p.wg.Done() // This will trigger the goroutine created in Start(), which will forward the Close() invokation to the DataSink
}

func (p *pluginSinkAdapter) String() string {
	return "plugin sink adapter"
}
