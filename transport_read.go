package bitflow

import (
	"bufio"
	"io"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/antongulenko/golib"
)

const MinimumInputIoBuffer = 16 // Needed for auto-detecting stream format

// SampleReader is used to read Headers and Samples from an io.Reader,
// parallelizing the reading and parsing procedures. The parallelization
// must be configured through the ParallelSampleHandler parameters before starting
// this SampleReader.
type SampleReader struct {
	ParallelSampleHandler

	// Handler is an optional hook for modifying Headers and Samples that were
	// read by this SampleReader. The hook method receives a string-representation of
	// the data source and can use it to modify tags in the Samples.
	Handler ReadSampleHandler

	// Unmarshaller will be used when reading and parsing Headers and Samples.
	// If this field is nil when creating an input stream, the SampleInputStream will try
	// to automatically determine the format of the incoming data and create
	// a fitting Unmarshaller instance accordingly.
	Unmarshaller Unmarshaller
}

// ReadSampleHandler defines a hook for modifying unmarshalled Samples.
type ReadSampleHandler interface {
	// HandleSample allows modifying received Samples. It can be used to modify
	// the tags of the Sample based on the source string. The source string depends on the
	// MetricSource that is using the SampleReader that contains this ReadSampleHandler.
	// In general it represents the data source of the sample. For FileSource this will be
	// the file name, for TCPSource it will be the remote TCP endpoint sending the data.
	// It might also be useful to change the values or the timestamp of the Sample here,
	// but that should rather be done in a later processing step.
	HandleSample(sample *Sample, source string)
}

// SampleInputStream represents one input stream of Headers and Samples that reads
// and parses data from one io.ReadCloser instance. A SampleInputStream can be created
// using SampleReader.Open or .OpenBuffered. The stream then has to be started using
// one of the Read* methods. The Read* method will block until the stream is finished.
// Reading and parsing the Samples will be done in parallel goroutines. The Read* methods behave
// differently in terms of printing errors. The stream can be closed forcefully using the
// Close method.
type SampleInputStream struct {
	parallelSampleStream
	incoming         chan *bufferedIncomingSample
	outgoing         chan *bufferedIncomingSample
	um               Unmarshaller
	sampleReader     *SampleReader
	reader           *bufio.Reader
	underlyingReader io.ReadCloser
	num_samples      int
	header           *UnmarshalledHeader // Header received from the input stream
	outHeader        *Header             // Header after modified by the ReadSampleHandler
	sink             MetricSinkBase
}

// Open creates an input stream reading from the given io.ReadCloser and writing
// the received Headers and Samples to the given MetricSinkBase. Although no buffer size
// is given, the stream will actually have a small input buffer to enable automatically
// detecting the format of incoming data, if no Unmarshaller was configured in the receiving
// SampleReader.
func (r *SampleReader) Open(input io.ReadCloser, sink MetricSinkBase) *SampleInputStream {
	return r.OpenBuffered(input, sink, MinimumInputIoBuffer)
}

// OpenBuffered creates an input stream with a given buffer size. The buffer size should be at least
// MinimumInputIoBuffer bytes to support automatically discovering the input stream format. See Open() for
// more details.
func (r *SampleReader) OpenBuffered(input io.ReadCloser, sink MetricSinkBase, bufSize int) *SampleInputStream {
	return &SampleInputStream{
		um:               r.Unmarshaller,
		reader:           bufio.NewReaderSize(input, bufSize),
		sampleReader:     r,
		underlyingReader: input,
		sink:             sink,
		incoming:         make(chan *bufferedIncomingSample, r.BufferedSamples),
		outgoing:         make(chan *bufferedIncomingSample, r.BufferedSamples),
		parallelSampleStream: parallelSampleStream{
			closed: golib.NewStopChan(),
		},
	}
}

// ReadSamples starts the receiving input stream and blocks until the stream is finished
// or closed by Close(). It returns the number of successfully received samples and a
// non-nil error, if any occurred while reading or parsing. The source string parameter
// will be forwarded to the ReadSampleHandler, if one is set in the SampleReader that
// created this SampleInputStream. The source string will be used for the HandleSample() method.
func (stream *SampleInputStream) ReadSamples(source string) (int, error) {
	if stream.um == nil {
		if um, err := detectFormat(stream.reader); err != nil {
			return 0, err
		} else {
			stream.um = um
		}
	}

	// Parse samples
	for i := 0; i < stream.sampleReader.ParallelParsers || i < 1; i++ {
		stream.wg.Add(1)
		go stream.parseSamples(source)
	}

	// Forward parsed samples
	stream.wg.Add(1)
	go stream.sinkSamples()

	stream.readData(source)
	stream.wg.Wait()
	return stream.num_samples, stream.getErrorNoEOF()
}

// ReadNamedSamples calls ReadSamples with the given source string, and prints
// some additional logging information. It is a convenience function for different
// implementations of MetricSource.
func (stream *SampleInputStream) ReadNamedSamples(sourceName string) (err error) {
	var num_samples int
	l := log.WithFields(log.Fields{"source": sourceName, "format": stream.Format()})
	l.Debugln("Reading samples")
	num_samples, err = stream.ReadSamples(sourceName)
	l.Debugln("Read", num_samples, "samples")
	return
}

// ReadTcpSamples reads Samples from the given net.TCPConn and blocks until the connection
// is closed by the remote host, or Close() is called on the input stream. Any error is logged
// instead of being returned. The checkClosed() function parameter is used when a read error occurs:
// if it returns true, ReadTcpSamples assumes that the connection was closed by the local host,
// because of a call to Close() or some other external reason. If checkClosed() returns false,
// it is assumed that a network error or timeout caused the connection to be closed.
func (stream *SampleInputStream) ReadTcpSamples(conn *net.TCPConn, checkClosed func() bool) {
	remote := conn.RemoteAddr()
	l := log.WithFields(log.Fields{"remote": remote, "format": stream.Format()})
	l.Debugln("Receiving data")
	var err error
	var num_samples int
	if num_samples, err = stream.ReadSamples(remote.String()); err == nil {
		l.Debugln("Connection closed by remote")
	} else {
		if checkClosed() {
			l.Debugln("Connection closed")
		} else {
			l.Errorln("Error receiving samples:", err)
		}
		_ = conn.Close() // Ignore error
	}
	l.Debugln("Received", num_samples, "samples")
}

// Close closes the receiving SampleInputStream. Close should be called even if the
// Read* method, that started the stream, returns an error. Close() might return the
// same error as the Read* method.
func (stream *SampleInputStream) Close() error {
	if stream != nil {
		stream.closeUnderlyingReader()
		return stream.getErrorNoEOF()
	}
	return nil
}

func (stream *SampleInputStream) closeUnderlyingReader() {
	stream.closed.StopFunc(func() {
		err := stream.underlyingReader.Close()
		stream.addError(err)
	})
}

// Format returns a string description of the unmarshalling format used by the receiving
// SampleReader. It returns "auto-detected", if no Unmarshaller is configured.
func (reader *SampleReader) Format() string {
	if reader.Unmarshaller == nil {
		return "auto-detected"
	} else {
		return reader.Unmarshaller.String()
	}
}

// Format returns a string description of the unmarshalling format used by the receiving
// SampleInputStream. It returns "auto-detected", if no Unmarshaller is configured, and if
// the unmarshalling format was not yet detected automatically. After the unmarshalling format
// is detected, Format will return the correct format description.
func (stream *SampleInputStream) Format() string {
	if stream.um == nil {
		return "auto-detected"
	} else {
		return stream.um.String()
	}
}

func (stream *SampleInputStream) readData(source string) {
	defer func() {
		stream.closeUnderlyingReader()
		close(stream.incoming)
		close(stream.outgoing)
	}()
	closedChan := stream.closed.WaitChan()
	for {
		if stream.hasError() {
			return
		}

		header, data, err := stream.um.Read(stream.reader, stream.header)
		if err != nil {
			stream.addError(err)
			if err != io.EOF || (len(data) == 0 && header == nil) {
				return
			}
		}
		if header != nil {
			stream.updateHeader(header, source)
		} else {
			s := &bufferedIncomingSample{
				inHeader:  stream.header,
				outHeader: stream.outHeader,
				bufferedSample: bufferedSample{
					stream:   &stream.parallelSampleStream,
					data:     data,
					doneCond: sync.NewCond(new(sync.Mutex)),
				},
			}
			select {
			case stream.outgoing <- s:
			case <-closedChan:
				return
			}
			stream.incoming <- s
			if err != nil {
				return
			}
		}
	}
}

func (stream *SampleInputStream) updateHeader(header *UnmarshalledHeader, source string) error {
	logger := log.WithFields(log.Fields{"format": stream.um, "source": source})
	if stream.header == nil {
		logger.Println("Reading", len(header.Fields), "metrics")
	} else {
		logger.Println("Updated header to", len(header.Fields), "metrics")
	}
	stream.header = header
	stream.outHeader = new(Header)
	if numFields := len(header.Fields); numFields > 0 {
		stream.outHeader.Fields = make([]string, numFields)
		copy(stream.outHeader.Fields, header.Fields)
	}
	return nil
}

func (stream *SampleInputStream) parseSamples(source string) {
	defer stream.wg.Done()
	for sample := range stream.incoming {
		stream.parseOne(source, sample)
	}
}

func (stream *SampleInputStream) parseOne(source string, sample *bufferedIncomingSample) {
	defer sample.notifyDone()
	numValues := RequiredValues(len(sample.inHeader.Fields), stream.sink)
	if parsedSample, err := stream.um.ParseSample(sample.inHeader, numValues, sample.data); err != nil {
		stream.addError(err)
		sample.ParserError = true
		return
	} else {
		if handler := stream.sampleReader.Handler; handler != nil {
			handler.HandleSample(parsedSample, source)
		}
		sample.sample = parsedSample
	}
}

func (stream *SampleInputStream) sinkSamples() {
	defer stream.wg.Done()
	for sample := range stream.outgoing {
		sample.waitDone()
		if sample.ParserError {
			// The first parser error makes the input stream stop.
			return
		}
		if err := stream.sink.Sample(sample.sample, sample.outHeader); err != nil {
			stream.addError(err)
			return
		}
		stream.num_samples++
	}
}

// ============= Helper types =============

type bufferedIncomingSample struct {
	bufferedSample
	ParserError bool
	inHeader    *UnmarshalledHeader
	outHeader   *Header
}
