package sample

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"

	"github.com/antongulenko/golib"
)

// Unmarshalls samples from an io.Reader, parallelizing the parsing
type SampleReader struct {
	ParallelSampleHandler
	ReadHook SampleReadHook // Allows modifying tags on read samples
}

type SampleReadHook func(sample *Sample, source string)

type SampleInputStream struct {
	ParallelSampleStream
	readHook         SampleReadHook
	reader           *bufio.Reader
	underlyingReader io.ReadCloser
	num_samples      int
	numParsers       int
	header           Header
	outHeader        Header
	um               Unmarshaller
	sink             MetricSink
}

func (r *SampleReader) Open(input io.ReadCloser, um Unmarshaller, sink MetricSink) *SampleInputStream {
	return &SampleInputStream{
		reader:           bufio.NewReaderSize(input, r.IoBuffer),
		readHook:         r.ReadHook,
		underlyingReader: input,
		um:               um,
		sink:             sink,
		numParsers:       r.ParallelParsers,
		ParallelSampleStream: ParallelSampleStream{
			incoming: make(chan *BufferedSample, r.BufferedSamples),
			outgoing: make(chan *BufferedSample, r.BufferedSamples),
			closed:   golib.NewOneshotCondition(),
		},
	}
}

func (stream *SampleInputStream) ReadSamples(source string) (int, error) {
	if err := stream.readHeader(); err != nil {
		return 0, err
	}

	// Parse samples
	for i := 0; i < stream.numParsers || i < 1; i++ {
		stream.wg.Add(1)
		go stream.parseSamples(source)
	}

	// Forward parsed samples
	stream.wg.Add(1)
	go stream.sinkSamples()

	stream.readData()
	stream.wg.Wait()
	if stream.err == io.EOF {
		stream.err = nil // io.EOF is expected
	}
	return stream.num_samples, stream.err
}

func (stream *SampleInputStream) ReadNamedSamples(sourceName string) (err error) {
	var num_samples int
	log.Println("Reading", stream.um, "from", sourceName)
	num_samples, err = stream.ReadSamples(sourceName)
	log.Printf("Read %v %v samples from %v\n", num_samples, stream.um, sourceName)
	return
}

func (stream *SampleInputStream) ReadTcpSamples(conn *net.TCPConn, checkClosed func() bool) {
	remote := conn.RemoteAddr()
	log.Println("Receiving", stream.um, "from", remote)
	var err error
	var num_samples int
	if num_samples, err = stream.ReadSamples(remote.String()); err == nil {
		log.Println("Connection closed by", remote)
	} else {
		if checkClosed() {
			log.Println("Connection to", remote, "closed")
		} else {
			log.Printf("Error receiving samples from %v: %v\n", remote, err)
		}
		_ = conn.Close() // Ignore error
	}
	log.Println("Received", num_samples, "samples from", remote)
}

func (stream *SampleInputStream) Close() error {
	if stream != nil {
		stream.closeUnderlyingReader()
		return stream.err
	}
	return nil
}

func (stream *SampleInputStream) closeUnderlyingReader() {
	stream.closed.Enable(func() {
		err := stream.underlyingReader.Close()
		if !stream.HasError() {
			stream.err = err
		}
	})
}

func (stream *SampleInputStream) readHeader() (err error) {
	if stream.header, err = stream.um.ReadHeader(stream.reader); err != nil {
		return
	}
	log.Printf("Reading %v metrics\n", len(stream.header.Fields))
	hasTags := stream.header.HasTags || stream.readHook != nil
	stream.outHeader = Header{Fields: stream.header.Fields, HasTags: hasTags}
	if err = stream.sink.Header(stream.outHeader); err != nil {
		return
	}
	return
}

func (stream *SampleInputStream) readData() {
	defer func() {
		stream.closeUnderlyingReader()
		close(stream.incoming)
		close(stream.outgoing)
	}()
	closedChan := stream.closed.Start(nil)
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
			select {
			case stream.outgoing <- s:
			case <-closedChan:
				return
			}
			stream.incoming <- s
		}
	}
}

func (stream *SampleInputStream) parseSamples(source string) {
	defer stream.wg.Done()
	for sample := range stream.incoming {
		stream.parseOne(source, sample)
	}
}

func (stream *SampleInputStream) parseOne(source string, sample *BufferedSample) {
	defer sample.NotifyDone()
	if stream.HasError() {
		return
	}
	if parsedSample, err := stream.um.ParseSample(stream.header, sample.data); err != nil {
		stream.err = err
		return
	} else {
		if hook := stream.readHook; hook != nil {
			hook(&parsedSample, source)
		}
		sample.sample = parsedSample
	}
}

func (stream *SampleInputStream) sinkSamples() {
	defer stream.wg.Done()
	for sample := range stream.outgoing {
		if err := sample.WaitDone(); err != nil {
			return
		}
		if err := stream.sink.Sample(sample.sample, stream.outHeader); err != nil {
			stream.err = err
			return
		}
		stream.num_samples++
	}
}
