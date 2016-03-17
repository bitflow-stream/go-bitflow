package metrics

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/antongulenko/golib"
)

// ==================== Data Sink ====================
type MetricSink interface {
	golib.Task
	SetMarshaller(marshaller Marshaller)
	Header(header Header) error
	Sample(sample Sample) error
}

type abstractSink struct {
	header     Header
	marshaller Marshaller
}

func (sink *abstractSink) SetMarshaller(marshaller Marshaller) {
	sink.marshaller = marshaller
}

func (sink *abstractSink) checkSample(sample Sample) error {
	if len(sample.Values) != len(sink.header) {
		return fmt.Errorf("Unexpected number of values in sample: %v, expected %v", len(sample.Values), len(sink.header))
	}
	return nil
}

// ==================== Data Source ====================
type MetricSource interface {
	golib.Task
	SetSink(sink MetricSink)
}

type UnmarshallingMetricSource interface {
	MetricSource
	SetUnmarshaller(unmarshaller Unmarshaller) // Must be called before Start()
}

type unmarshallingMetricSource struct {
	Unmarshaller Unmarshaller
	Sink         MetricSink
}

func (s *unmarshallingMetricSource) SetSink(sink MetricSink) {
	s.Sink = sink
}

func (s *unmarshallingMetricSource) SetUnmarshaller(unmarshaller Unmarshaller) {
	s.Unmarshaller = unmarshaller
}

func readSamples(input io.Reader, um Unmarshaller, sink MetricSink) (int, error) {
	reader := bufio.NewReader(input)
	var header Header
	if err := um.ReadHeader(&header, reader); err != nil {
		return 0, err
	}
	if err := sink.Header(header); err != nil {
		return 0, err
	}
	log.Printf("Reading %v metrics\n", len(header))

	num_samples := 0
	for {
		var sample Sample
		if err := um.ReadSample(&sample, reader, len(header)); err != nil {
			return num_samples, err
		}
		if err := sink.Sample(sample); err != nil {
			log.Printf("Error forwarding received sample: %v\n", err)
		}
		num_samples++
	}
}

func simpleReadSamples(wg *sync.WaitGroup, sourceName string, input io.Reader, um Unmarshaller, sink MetricSink) golib.StopChan {
	return golib.WaitErrFunc(wg, func() (err error) {
		var num_samples int
		log.Println("Reading", um, "from", sourceName)
		if num_samples, err = readSamples(input, um, sink); err != nil && err != io.EOF {
			err = fmt.Errorf("Read failed: %v", err)
		}
		log.Printf("Read %v %v samples from %v\n", num_samples, um, sourceName)
		return
	})
}

// ==================== Aggregating Sink ====================
type AggregateSink []MetricSink

// The golib.Task interface cannot really be supported here
func (agg AggregateSink) Start(wg *sync.WaitGroup) golib.StopChan {
	panic("Start should not be called on AggregateSink")
}

func (agg AggregateSink) Stop() {
	panic("Stop should not be called on AggregateSink")
}

func (agg AggregateSink) SetMarshaller(marshaller Marshaller) {
	for _, sink := range agg {
		sink.SetMarshaller(marshaller)
	}
}

func (agg AggregateSink) Header(header Header) error {
	var errors golib.MultiError
	for _, sink := range agg {
		if err := sink.Header(header); err != nil {
			errors.Add(err)
		}
	}
	return errors.NilOrError()
}

func (agg AggregateSink) Sample(sample Sample) error {
	var errors golib.MultiError
	for _, sink := range agg {
		if err := sink.Sample(sample); err != nil {
			errors.Add(err)
		}
	}
	return errors.NilOrError()
}
