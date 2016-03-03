package metrics

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"sync"
)

// ==================== Data Sink ====================
type MetricSink interface {
	Start(wg *sync.WaitGroup, marshaller Marshaller) error
	Header(header Header) error
	Sample(sample Sample) error
}

type abstractSink struct {
	header     Header
	marshaller Marshaller
}

func (sink *abstractSink) checkSample(sample Sample) error {
	if len(sample.Values) != len(sink.header) {
		return fmt.Errorf("Unexpected number of values in sample: %v, expected %v", len(sample.Values), len(sink.header))
	}
	return nil
}

// ==================== Data Source ====================
type MetricSource interface {
	Start(wg *sync.WaitGroup, um Unmarshaller, sink MetricSink) error
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

func simpleReadSamples(wg *sync.WaitGroup, sourceName string, input io.Reader, um Unmarshaller, sink MetricSink) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		var num_samples int
		log.Println("Reading from", sourceName)
		if num_samples, err = readSamples(input, um, sink); err != nil && err != io.EOF {
			log.Println("Read failed:", err)
		}
		log.Printf("Read %v samples from %v\n", num_samples, sourceName)
	}()
}

// ==================== Aggregating Sink ====================
type AggregateSink []MetricSink

func (agg AggregateSink) Start(wg *sync.WaitGroup, marshaller Marshaller) error {
	for _, sink := range agg {
		if err := sink.Start(wg, marshaller); err != nil {
			return err
		}
	}
	return nil
}

func (agg AggregateSink) Header(header Header) error {
	for _, sink := range agg {
		if err := sink.Header(header); err != nil {
			return err
		}
	}
	return nil
}

func (agg AggregateSink) Sample(sample Sample) error {
	for _, sink := range agg {
		if err := sink.Sample(sample); err != nil {
			return err
		}
	}
	return nil
}
