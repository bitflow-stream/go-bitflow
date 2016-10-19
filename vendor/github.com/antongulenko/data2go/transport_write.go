package data2go

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sync"

	"github.com/antongulenko/golib"
)

// SampleWriter implements parallel writing of Headers and Samples to an instance of
// io.WriteCloser. The WriteCloser can be anything like a file or a network connection.
// The parallel writing must be configured before using the SampleWriter. See
// ParallelSampleHandler for the configuration variables.
//
// SampleWriter instances are mainly used by implementations of MetricSink that write
// to output streams, like FileSink or TCPSink.
type SampleWriter struct {
	ParallelSampleHandler
}

// SampleOutputStream represents one open output stream that marshalls and writes
// Headers and Samples in parallel. It is created by using SampleWriter.Open or
// SampleWriter.OpenBuffered. The Header()s and Sample() methods can be used
// to output data on this stream, and the Close() method must be called when no
// more Samples are expected. No more data can be written after calling Close().
type SampleOutputStream struct {
	parallelSampleStream
	incoming       chan *bufferedSample
	outgoing       chan *bufferedSample
	closed         *golib.OneshotCondition
	headerLock     sync.Mutex
	header         *Header
	writer         io.WriteCloser
	marshaller     Marshaller
	marshallBuffer int
}

// BufferedWriteCloser is a helper type that wraps a bufio.Writer around a
// io.WriteCloser, while still implementing the io.WriteCloser interface and
// forwarding all method calls to the correct receiver. The Writer field should
// not be accessed directly.
type BufferedWriteCloser struct {
	*bufio.Writer
	closer io.Closer
}

// NewBufferedWriteCloser creates a BufferedWriteCloser instance wrapping the
// writer parameter. It creates a bufio.Writer with a buffer size of the io_buffer
// parameter.
func NewBufferedWriteCloser(writer io.WriteCloser, io_buffer int) *BufferedWriteCloser {
	return &BufferedWriteCloser{
		Writer: bufio.NewWriterSize(writer, io_buffer),
		closer: writer,
	}
}

// Close implements the Close method in io.WriteCloser by flusing its bufio.Writer and
// forwarding the Close call to the io.WriteCloser used to create it.
func (writer *BufferedWriteCloser) Close() (err error) {
	err = writer.Writer.Flush()
	if closeErr := writer.closer.Close(); err == nil {
		err = closeErr
	}
	return
}

// OpenBufferd returns a buffered output stream with a buffer of the size io_buffer.
// Samples coming into that stream are marshalled using marshaller and finally written
// the given writer.
func (w *SampleWriter) OpenBuffered(writer io.WriteCloser, marshaller Marshaller, io_buffer int) *SampleOutputStream {
	if io_buffer <= 0 {
		return w.Open(writer, marshaller)
	}
	buf := NewBufferedWriteCloser(writer, io_buffer)
	return w.Open(buf, marshaller)
}

// Open returns an output stream that sends the marshalled samples directly to the given writer.
// Marshalling and writing is done in separate routines, as configured in the SampleWriter
// configuration parameters.
func (w *SampleWriter) Open(writer io.WriteCloser, marshaller Marshaller) *SampleOutputStream {
	stream := &SampleOutputStream{
		writer:     writer,
		marshaller: marshaller,
		closed:     golib.NewOneshotCondition(),
		incoming:   make(chan *bufferedSample, w.BufferedSamples),
		outgoing:   make(chan *bufferedSample, w.BufferedSamples),
	}

	for i := 0; i < w.ParallelParsers || i < 1; i++ {
		stream.wg.Add(1)
		go stream.marshall()
	}
	stream.wg.Add(1)
	go stream.flush()

	return stream
}

// Header marshalls the given header and writes the resulting byte buffer into
// the writer behind the stream receiver. If a non-nil error is returned here,
// the stream should not be used any further, but still must be closed externally.
func (stream *SampleOutputStream) Header(header *Header) error {
	if stream == nil {
		return nil
	}
	stream.headerLock.Lock()
	defer stream.headerLock.Unlock()
	if stream.hasError() {
		return stream.err
	}
	if stream.header == nil {
		if err := stream.marshaller.WriteHeader(header, stream.writer); err != nil {
			stream.err = err
		}
		if err := stream.flushBuffered(); err != nil {
			stream.err = err
		}
		stream.header = header
	} else {
		if !stream.hasError() && stream.header != nil {
			return errors.New("A header has already been written to this stream")
		}
	}
	return stream.err
}

func (stream *SampleOutputStream) flushBuffered() error {
	if buf, ok := stream.writer.(*BufferedWriteCloser); ok {
		return buf.Flush()
	}
	return nil
}

// Sample marshalles the given Sample and writes the resulting byte buffer into the
// writer behind the stream receiver. If a non-nil error is returned here,
// the stream should not be used any further, but still must be closed externally.
func (stream *SampleOutputStream) Sample(sample *Sample) error {
	if stream.hasError() {
		return stream.err
	}
	bufferedSample := &bufferedSample{
		stream:   &stream.parallelSampleStream,
		sample:   sample,
		doneCond: sync.NewCond(new(sync.Mutex)),
	}
	stream.incoming <- bufferedSample
	stream.outgoing <- bufferedSample
	return nil
}

// Close closes the receiving SampleOutputStream. After calling this, neither
// Sample nor Header can be called anymore! The returned error is the first error
// that ever ocurred in any of the Sample/Header/Close calls on this stream.
func (stream *SampleOutputStream) Close() error {
	if stream == nil {
		return nil
	}
	stream.closed.Enable(func() {
		close(stream.incoming)
		close(stream.outgoing)
		stream.wg.Wait()
		if err := stream.flushBuffered(); !stream.hasError() {
			stream.err = err
		}
		if err := stream.writer.Close(); !stream.hasError() {
			stream.err = err
		}
	})
	return stream.err
}

func (stream *SampleOutputStream) marshall() {
	defer stream.wg.Done()
	for sample := range stream.incoming {
		stream.marshallOne(sample)
	}
}

func (stream *SampleOutputStream) marshallOne(sample *bufferedSample) {
	defer sample.notifyDone()
	if stream.hasError() {
		return
	}
	if stream.header == nil {
		if !stream.hasError() {
			stream.err = errors.New("Cannot write Sample before a Header")
		}
		return
	}
	buf := bytes.NewBuffer(make([]byte, 0, stream.marshallBuffer))
	stream.marshaller.WriteSample(sample.sample, stream.header, buf)
	if l := buf.Len(); l > stream.marshallBuffer {
		// Avoid buffer copies for future samples
		// TODO could reuse allocated buffers
		stream.marshallBuffer = l
	}
	sample.data = buf.Bytes()
}

func (stream *SampleOutputStream) flush() {
	defer stream.wg.Done()
	for sample := range stream.outgoing {
		sample.waitDone()
		if stream.hasError() {
			return
		}
		if _, err := stream.writer.Write(sample.data); err != nil {
			stream.err = err
			return
		}
	}
}
