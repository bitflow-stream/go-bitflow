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
}

type sampleReaderState struct {
	reader         *bufio.Reader
	err            error
	num_samples    int
	read_samples   chan *parsedSample
	parsed_samples chan *parsedSample
	wg             sync.WaitGroup
	header         Header
	um             Unmarshaller
	sink           MetricSink
}

type parsedSample struct {
	data       []byte
	sample     Sample
	parsed     bool
	parsedCond *sync.Cond
}

func (r *SampleReader) ReadSamples(input io.Reader, um Unmarshaller, sink MetricSink) (int, error) {
	state := &sampleReaderState{
		reader:         bufio.NewReader(input),
		read_samples:   make(chan *parsedSample, r.BufferedSamples),
		parsed_samples: make(chan *parsedSample, r.BufferedSamples),
		um:             um,
		sink:           sink,
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

func (state *sampleReaderState) hasError() bool {
	return state.err != nil && state.err != io.EOF
}

func (state *sampleReaderState) readHeader() (err error) {
	if state.header, err = state.um.ReadHeader(state.reader); err != nil {
		return
	}
	if err = state.sink.Header(state.header); err != nil {
		return
	}
	log.Printf("Reading %v metrics\n", len(state.header.Fields))
	return
}

func (state *sampleReaderState) readData() {
	defer func() {
		state.wg.Done()
		close(state.read_samples)
		close(state.parsed_samples)
	}()
	for {
		if state.hasError() {
			return
		}
		if data, err := state.um.ReadSampleData(state.header, state.reader); err != nil {
			state.err = err
			return
		} else {
			s := &parsedSample{
				data:       data,
				parsedCond: sync.NewCond(new(sync.Mutex)),
			}
			state.read_samples <- s
			state.parsed_samples <- s
		}
	}
}

func (state *sampleReaderState) parseSamples() {
	defer state.wg.Done()
	for sample := range state.read_samples {
		if state.hasError() {
			break
		}
		if parsedSample, err := state.um.ParseSample(state.header, sample.data); err != nil {
			state.err = err
			return
		} else {
			sample.sample = parsedSample
		}
		func() {
			sample.parsedCond.L.Lock()
			defer sample.parsedCond.L.Unlock()
			sample.parsed = true
			sample.parsedCond.Broadcast()
		}()
	}
}

func (state *sampleReaderState) sinkSamples() {
	defer state.wg.Done()
	for sample := range state.parsed_samples {
		if state.hasError() {
			return
		}
		func() {
			sample.parsedCond.L.Lock()
			defer sample.parsedCond.L.Unlock()
			for !sample.parsed && !state.hasError() {
				sample.parsedCond.Wait()
			}
		}()
		if state.hasError() {
			return
		}
		if err := state.sink.Sample(sample.sample, state.header); err != nil {
			err = err
			return
		}
		state.num_samples++
	}
}
