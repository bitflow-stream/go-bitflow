package sample

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sync"

	"github.com/antongulenko/golib"
)

type SampleWriter struct {
	ParallelSampleHandler
}

type SampleOutputStream struct {
	ParallelSampleStream
	incoming       chan *BufferedSample
	outgoing       chan *BufferedSample
	closed         *golib.OneshotCondition
	headerLock     sync.Mutex
	header         *Header
	writer         io.WriteCloser
	marshaller     Marshaller
	marshallBuffer int
}

type BufferedWriteCloser struct {
	*bufio.Writer
	Closer io.Closer
}

func (writer *BufferedWriteCloser) Close() (err error) {
	err = writer.Writer.Flush()
	if closeErr := writer.Closer.Close(); err == nil {
		err = closeErr
	}
	return
}

func (w *SampleWriter) OpenBuffered(writer io.WriteCloser, marshaller Marshaller, io_buffer int) *SampleOutputStream {
	if io_buffer <= 0 {
		return w.Open(writer, marshaller)
	}
	buf := &BufferedWriteCloser{
		Writer: bufio.NewWriterSize(writer, io_buffer),
		Closer: writer,
	}
	return w.Open(buf, marshaller)
}

func (w *SampleWriter) Open(writer io.WriteCloser, marshaller Marshaller) *SampleOutputStream {
	stream := &SampleOutputStream{
		writer:     writer,
		marshaller: marshaller,
		closed:     golib.NewOneshotCondition(),
		incoming:   make(chan *BufferedSample, w.BufferedSamples),
		outgoing:   make(chan *BufferedSample, w.BufferedSamples),
	}

	for i := 0; i < w.ParallelParsers || i < 1; i++ {
		stream.wg.Add(1)
		go stream.marshall()
	}
	stream.wg.Add(1)
	go stream.flush()

	return stream
}

func (stream *SampleOutputStream) Header(header Header) error {
	if stream == nil {
		return nil
	}
	stream.headerLock.Lock()
	defer stream.headerLock.Unlock()
	if stream.HasError() {
		return stream.err
	}
	if stream.header == nil {
		if err := stream.marshaller.WriteHeader(header, stream.writer); err != nil {
			stream.err = err
		}
		if err := stream.flushBuffered(); err != nil {
			stream.err = err
		}
		stream.header = &header
	} else {
		if !stream.HasError() && stream.header != nil {
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

func (stream *SampleOutputStream) Sample(sample Sample) error {
	if stream.HasError() {
		return stream.err
	}
	bufferedSample := &BufferedSample{
		stream:   &stream.ParallelSampleStream,
		sample:   sample,
		doneCond: sync.NewCond(new(sync.Mutex)),
	}
	stream.incoming <- bufferedSample
	stream.outgoing <- bufferedSample
	return nil
}

func (stream *SampleOutputStream) Close() error {
	if stream == nil {
		return nil
	}
	stream.closed.Enable(func() {
		close(stream.incoming)
		close(stream.outgoing)
		stream.wg.Wait()
		if err := stream.flushBuffered(); !stream.HasError() {
			stream.err = err
		}
		if err := stream.writer.Close(); !stream.HasError() {
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

func (stream *SampleOutputStream) marshallOne(sample *BufferedSample) {
	defer sample.NotifyDone()
	if stream.HasError() {
		return
	}
	if stream.header == nil {
		if !stream.HasError() {
			stream.err = errors.New("Cannot write Sample before a Header")
		}
		return
	}
	buf := bytes.NewBuffer(make([]byte, 0, stream.marshallBuffer))
	stream.marshaller.WriteSample(sample.sample, *stream.header, buf)
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
		sample.WaitDone()
		if stream.HasError() {
			return
		}
		if _, err := stream.writer.Write(sample.data); err != nil {
			stream.err = err
			return
		}
	}
}
