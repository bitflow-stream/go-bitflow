package sample

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"
)

// Unmarshalls samples from an io.Reader, parallelizing the parsing
type SampleReader struct {
	ParallelParsers int
	BufferedSamples int
	IoBuffer        int
}

type SampleInputStream struct {
	ParallelSampleStream
	reader      *bufio.Reader
	num_samples int
	header      Header
	um          Unmarshaller
	sink        MetricSink
}

func (r *SampleReader) ReadSamples(input io.Reader, um Unmarshaller, sink MetricSink) (int, error) {
	state := &SampleInputStream{
		reader: bufio.NewReaderSize(input, r.IoBuffer),
		um:     um,
		sink:   sink,
		ParallelSampleStream: ParallelSampleStream{
			incoming: make(chan *BufferedSample, r.BufferedSamples),
			outgoing: make(chan *BufferedSample, r.BufferedSamples),
		},
	}
	if err := state.readHeader(); err != nil {
		return 0, err
	}

	// Read sample data
	state.wg.Add(1)
	go state.readData()

	// Parse samples
	for i := 0; i < r.ParallelParsers || i < 1; i++ {
		state.wg.Add(1)
		go state.parseSamples()
	}

	// Forward parsed samples
	state.wg.Add(1)
	go state.sinkSamples()

	state.wg.Wait()
	return state.num_samples, state.err
}

func (r *SampleReader) ReadNamedSamples(sourceName string, input io.Reader, um Unmarshaller, sink MetricSink) (err error) {
	var num_samples int
	log.Println("Reading", um, "from", sourceName)
	num_samples, err = r.ReadSamples(input, um, sink)
	if err == io.EOF {
		err = nil
	}
	log.Printf("Read %v %v samples from %v\n", num_samples, um, sourceName)
	return
}

func (r *SampleReader) ReadTcpSamples(conn *net.TCPConn, um Unmarshaller, sink MetricSink, checkClosed func() bool) {
	log.Println("Receiving header from", conn.RemoteAddr())
	var err error
	var num_samples int
	if num_samples, err = r.ReadSamples(conn, um, sink); err == io.EOF {
		log.Println("Connection closed by", conn.RemoteAddr())
	} else if checkClosed() {
		log.Println("Connection to", conn.RemoteAddr(), "closed")
	} else if err != nil {
		log.Printf("Error receiving samples from %v: %v\n", conn.RemoteAddr(), err)
		_ = conn.Close() // Ignore error
	}
	log.Println("Received", num_samples, "samples from", conn.RemoteAddr())
}

func (stream *SampleInputStream) readHeader() (err error) {
	if stream.header, err = stream.um.ReadHeader(stream.reader); err != nil {
		return
	}
	if err = stream.sink.Header(stream.header); err != nil {
		return
	}
	log.Printf("Reading %v metrics\n", len(stream.header.Fields))
	return
}

func (stream *SampleInputStream) readData() {
	defer func() {
		stream.wg.Done()
		close(stream.incoming)
		close(stream.outgoing)
	}()
	for {
		if stream.HasError() {
			return
		}
		if data, err := stream.um.ReadSampleData(stream.header, stream.reader); err != nil {
			stream.err = err
			return
		} else {
			s := &BufferedSample{
				stream:   &stream.ParallelSampleStream,
				data:     data,
				doneCond: sync.NewCond(new(sync.Mutex)),
			}
			stream.incoming <- s
			stream.outgoing <- s
		}
	}
}

func (stream *SampleInputStream) parseSamples() {
	defer stream.wg.Done()
	for sample := range stream.incoming {
		stream.parseOne(sample)
	}
}

func (stream *SampleInputStream) parseOne(sample *BufferedSample) {
	defer sample.NotifyDone()
	if stream.HasError() {
		return
	}
	if parsedSample, err := stream.um.ParseSample(stream.header, sample.data); err != nil {
		stream.err = err
		return
	} else {
		sample.sample = parsedSample
	}
}

func (stream *SampleInputStream) sinkSamples() {
	defer stream.wg.Done()
	for sample := range stream.outgoing {
		if err := sample.WaitDone(); err != nil {
			return
		}
		if err := stream.sink.Sample(sample.sample, stream.header); err != nil {
			stream.err = err
			return
		}
		stream.num_samples++
	}
}
