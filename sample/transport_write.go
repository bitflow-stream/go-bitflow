package sample

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sync"
)

type SampleWriter struct {
	BufferedSamples     int
	IoBuffer            int
	MarshallingRoutines int
}

type SampleOutputStream struct {
	ParallelSampleStream
	headerLock     sync.Mutex
	header         *Header
	writer         io.Writer
	marshaller     Marshaller
	marshallBuffer int
}

func (w *SampleWriter) OpenBuffered(writer io.Writer, marshaller Marshaller) *SampleOutputStream {
	buf := bufio.NewWriterSize(writer, w.IoBuffer)
	return w.Open(buf, marshaller)
}

func (w *SampleWriter) Open(writer io.Writer, marshaller Marshaller) *SampleOutputStream {
	stream := &SampleOutputStream{
		writer:     writer,
		marshaller: marshaller,
		ParallelSampleStream: ParallelSampleStream{
			incoming: make(chan *BufferedSample, w.BufferedSamples),
			outgoing: make(chan *BufferedSample, w.BufferedSamples),
		},
	}

	for i := 0; i < w.MarshallingRoutines || i < 1; i++ {
		stream.wg.Add(1)
		go stream.marshall()
	}
	stream.wg.Add(1)
	go stream.flush()

	return stream
}

func (stream *SampleOutputStream) Header(header Header) error {
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
	if buf, ok := stream.writer.(*bufio.Writer); ok {
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
	close(stream.incoming)
	close(stream.outgoing)
	stream.wg.Wait()
	err := stream.flushBuffered()
	if stream.HasError() {
		err = stream.err
	}
	return err
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
		stream.err = errors.New("Cannot write Sample before a Header")
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
		if err := sample.WaitDone(); err != nil {
			return
		}
		if _, err := stream.writer.Write(sample.data); err != nil {
			stream.err = err
			return
		}
	}
}
