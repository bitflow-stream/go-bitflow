package sample

import (
	"bufio"
	"io"
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
)

const input_io_buffer = 16 // Needed for auto-detecting stream format

// Unmarshalls samples from an io.Reader, parallelizing the parsing
type SampleReader struct {
	ParallelSampleHandler
	Handler      ReadSampleHandler // Optional, for modifying incoming headers/samples based on their source
	Unmarshaller Unmarshaller
}

type ReadSampleHandler interface {
	HandleHeader(header *Header, source string) // Allows modifying fields on incoming headers
	HandleSample(sample *Sample, source string) // Allows modifying tags/values on read samples
}

type SampleInputStream struct {
	ParallelSampleStream
	um               Unmarshaller
	sampleReader     *SampleReader
	reader           *bufio.Reader
	underlyingReader io.ReadCloser
	num_samples      int
	header           Header
	outHeader        Header
	sink             MetricSinkBase
}

func (r *SampleReader) Open(input io.ReadCloser, sink MetricSinkBase) *SampleInputStream {
	return r.OpenBuffered(input, sink, input_io_buffer)
}

func (r *SampleReader) OpenBuffered(input io.ReadCloser, sink MetricSinkBase, bufSize int) *SampleInputStream {
	return &SampleInputStream{
		um:               r.Unmarshaller,
		reader:           bufio.NewReaderSize(input, bufSize),
		sampleReader:     r,
		underlyingReader: input,
		sink:             sink,
		ParallelSampleStream: ParallelSampleStream{
			incoming: make(chan *BufferedSample, r.BufferedSamples),
			outgoing: make(chan *BufferedSample, r.BufferedSamples),
			closed:   golib.NewOneshotCondition(),
		},
	}
}

func (stream *SampleInputStream) ReadSamples(source string) (int, error) {
	if err := stream.readHeader(source); err != nil {
		return 0, err
	}

	// Parse samples
	for i := 0; i < stream.sampleReader.ParallelParsers || i < 1; i++ {
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
	l := log.WithFields(log.Fields{"source": sourceName, "format": stream.Format()})
	l.Println("Reading samples")
	num_samples, err = stream.ReadSamples(sourceName)
	l.Println("Read", num_samples, "samples")
	return
}

func (stream *SampleInputStream) ReadTcpSamples(conn *net.TCPConn, checkClosed func() bool) {
	remote := conn.RemoteAddr()
	l := log.WithFields(log.Fields{"remote": remote, "format": stream.Format()})
	l.Println("Receiving data")
	var err error
	var num_samples int
	if num_samples, err = stream.ReadSamples(remote.String()); err == nil {
		l.Println("Connection closed by remote")
	} else {
		if checkClosed() {
			l.Println("Connection closed")
		} else {
			l.Errorln("Error receiving samples:", err)
		}
		_ = conn.Close() // Ignore error
	}
	l.Println("Received", num_samples, "samples")
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

func (reader *SampleReader) Format() string {
	if reader.Unmarshaller == nil {
		return "auto-detected"
	} else {
		return reader.Unmarshaller.String()
	}
}

func (stream *SampleInputStream) Format() string {
	if stream.um == nil {
		return "auto-detected"
	} else {
		return stream.um.String()
	}
}

func (stream *SampleInputStream) readHeader(source string) (err error) {
	if stream.um == nil {
		if stream.um, err = detectFormat(stream.reader); err != nil {
			return
		}
	}
	if stream.header, err = stream.um.ReadHeader(stream.reader); err != nil {
		return
	}
	log.WithFields(log.Fields{"format": stream.um, "source": source}).Println("Reading", len(stream.header.Fields), "metrics")
	stream.outHeader = Header{
		Fields:  make([]string, len(stream.header.Fields)),
		HasTags: stream.header.HasTags,
	}
	copy(stream.outHeader.Fields, stream.header.Fields)
	if handler := stream.sampleReader.Handler; handler != nil {
		handler.HandleHeader(&stream.outHeader, source)
	}
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
	// TODO See BufferedSample.WaitDone()
	/*
		if stream.HasError() {
			return
		}
	*/
	if parsedSample, err := stream.um.ParseSample(stream.header, sample.data); err != nil {
		if !stream.HasError() {
			stream.err = err
		}
		return
	} else {
		if handler := stream.sampleReader.Handler; handler != nil {
			handler.HandleSample(&parsedSample, source)
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
