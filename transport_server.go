package bitflow

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
)

// TCPListenerSource implements the MetricSource interface as a TCP server.
// It listens for incoming TCP connections on a port and reads Headers and Samples
// from every accepted connection. See the doc for the different fields for options
// affecting the TCP connections and aspects of reading and parsing.
type TCPListenerSource struct {
	AbstractUnmarshallingMetricSource

	// TCPConnCounter has a configuration for limiting the total number of
	// accepted connections. After that number of connections were accepted, no
	// further connections are accepted. After they all are closed, the TCPListenerSource
	// automatically stops.
	TCPConnCounter

	// SimultaneousConnections can limit the number of TCP connections accepted
	// at the same time. Set to >0 to activate the limit. Connections going over
	// the limit will be immediately closed, and a warning will be printed on the logger.
	SimultaneousConnections uint

	task             *golib.TCPListenerTask
	synchronizedSink MetricSinkBase
	connections      map[*tcpListenerConnection]bool
}

// NewTcpListenerSource creates a new instance of TCPListenerSource listening on the given
// TCP endpoint. It must be a IP/hostname combined with a port that can be bound on the local
// machine.
func NewTcpListenerSource(endpoint string) *TCPListenerSource {
	source := &TCPListenerSource{
		connections: make(map[*tcpListenerConnection]bool),
	}
	source.task = &golib.TCPListenerTask{
		ListenEndpoint: endpoint,
	}
	return source
}

// String implements the MetricSource interface.
func (source *TCPListenerSource) String() string {
	return "TCP source on " + source.task.ListenEndpoint
}

// Start implements the MetricSource interface. It creates a socket listening
// for incoming connections on the configured endpoint. New connections are
// handled in separate goroutines.
func (source *TCPListenerSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.connCounterDescription = source
	source.task.Handler = source.handleConnection
	source.task.StopHook = func() {
		source.CloseSink(wg)
	}
	if source.SimultaneousConnections == 1 {
		source.synchronizedSink = source.OutgoingSink
	} else {
		source.synchronizedSink = &SynchronizingMetricSink{OutgoingSink: source.OutgoingSink}
	}
	return source.task.ExtendedStart(func(addr net.Addr) {
		log.WithField("format", source.Reader.Format()).Println("Listening for incoming data on", addr)
	}, wg)
}

func (source *TCPListenerSource) handleConnection(wg *sync.WaitGroup, conn *net.TCPConn) {
	if source.SimultaneousConnections > 0 && len(source.connections) >= int(source.SimultaneousConnections) {
		log.WithField("remote", conn.RemoteAddr()).Warnln("Rejecting connection, already have", len(source.connections), "connections")
		_ = conn.Close() // Drop error
		return
	}
	if !source.countConnectionAccepted(conn) {
		return
	}
	log.WithField("remote", conn.RemoteAddr()).Debugln("Accepted connection")
	listenerConn := &tcpListenerConnection{
		source: source,
		stream: source.Reader.Open(conn, source.synchronizedSink),
	}
	source.connections[listenerConn] = true
	wg.Add(1)
	go listenerConn.readSamples(wg, conn)
}

// Stop implements the MetricSource interface. It close sall active TCP connections
// and closes the listening socket.
func (source *TCPListenerSource) Stop() {
	source.task.ExtendedStop(func() {
		for conn := range source.connections {
			conn.closeStream()
		}
	})
}

type tcpListenerConnection struct {
	source *TCPListenerSource
	stream *SampleInputStream
}

func (conn *tcpListenerConnection) isConnectionClosed() bool {
	return conn.stream == nil
}

func (conn *tcpListenerConnection) readSamples(wg *sync.WaitGroup, connection *net.TCPConn) {
	defer wg.Done()
	conn.stream.ReadTcpSamples(connection, conn.isConnectionClosed)
	if !conn.source.countConnectionClosed() {
		conn.source.Stop()
	}
	conn.source.task.IfRunning(conn.closeStream)
}

func (conn *tcpListenerConnection) closeStream() {
	if stream := conn.stream; stream != nil {
		conn.stream = nil  // Make isConnectionClosed() return true
		_ = stream.Close() // Drop error
		delete(conn.source.connections, conn)
	}
}

// TCPListenerSink implements the MetricSink interface through a TCP server.
// It creates a socket listening on a local TCP endpoint and listens for incoming
// TCP connections. Once one or more connections are established, it forwards
// all incoming Headers and Samples to those connections. If a new header should
// be sent into a TCP connection, the old connection is instead closed and
// the TCPListenerSink waits for a new connection to be created.
type TCPListenerSink struct {
	// AbstractTcpSink defines parameters for controlling TCP and marshalling
	// aspects of the TCPListenerSink. See AbstractTcpSink for details.
	AbstractTcpSink

	// Endpoint defines the TCP host and port to listen on for incoming TCP connections.
	// The host can be empty (e.g. ":1234"). If not, it must contain a hostname or IP of the
	// local host.
	Endpoint string

	// If BufferedSamples is >0, the given number of samples will be kept in a ring buffer.
	// New incoming connections will first receive all samples currently in the buffer, and will
	// afterwards continue receiving live incoming samples.
	BufferedSamples uint

	buf  outputSampleBuffer
	task *golib.TCPListenerTask
}

// String implements the MetrincSink interface.
func (sink *TCPListenerSink) String() string {
	return "TCP sink on " + sink.Endpoint
}

// Start implements the MetricSink interface. It creates the TCP socket and
// starts listening on it in a separate goroutine. Any incoming connection is
// then handled in their own goroutine.
func (sink *TCPListenerSink) Start(wg *sync.WaitGroup) golib.StopChan {
	sink.connCounterDescription = sink
	capacity := sink.BufferedSamples
	if capacity == 0 {
		capacity = 1
	}
	sink.buf = outputSampleBuffer{
		Capacity: capacity,
		cond:     sync.NewCond(new(sync.Mutex)),
	}
	sink.task = &golib.TCPListenerTask{
		ListenEndpoint: sink.Endpoint,
		Handler:        sink.handleConnection,
	}
	return sink.task.ExtendedStart(func(addr net.Addr) {
		log.WithField("format", sink.Marshaller).Println("Listening for output connections on", addr)
	}, wg)
}

// Close implements the MetricSink interface. It closes any existing connection
// and closes the TCP socket.
func (sink *TCPListenerSink) Close() {
	sink.task.ExtendedStop(func() {
		sink.buf.close()
	})
}

func (sink *TCPListenerSink) handleConnection(wg *sync.WaitGroup, conn *net.TCPConn) {
	if !sink.countConnectionAccepted(conn) {
		return
	}
	writeConn := sink.OpenWriteConn(conn)
	wg.Add(1)
	go sink.sendSamples(wg, writeConn)
}

// Sample implements the MetricSink interface. It stores the sample in a ring buffer
// and sends it to all established connections. New connections will first receive
// all samples stored in the buffer, before getting the live samples directly.
// If the buffer is disable or full, and there are no established connections,
// samples are dropped.
func (sink *TCPListenerSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	sink.buf.add(sample, header)
	return nil
}

func (sink *TCPListenerSink) closeConn(conn *TcpWriteConn) {
	conn.Close()
	if !sink.countConnectionClosed() {
		sink.Close()
	}
}

func (sink *TCPListenerSink) sendSamples(wg *sync.WaitGroup, conn *TcpWriteConn) {
	defer sink.closeConn(conn)
	defer wg.Done()
	first, num := sink.buf.getFirst()
	if num > 1 {
		conn.log.Debugln("Sending", num, "buffered samples")
	}
	for first != nil {
		if !conn.IsRunning() {
			return
		}
		conn.Sample(first.sample, first.header)
		if !conn.IsRunning() {
			return
		}
		first = sink.buf.next(first)
	}
}

// ======================================= output sample buffer =======================================

type outputSampleBuffer struct {
	Capacity uint

	size   uint
	first  *sampleListLink
	last   *sampleListLink
	cond   *sync.Cond
	closed bool
}

type sampleListLink struct {
	sample *Sample
	header *Header
	next   *sampleListLink
}

func (b *outputSampleBuffer) add(sample *Sample, header *Header) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	link := &sampleListLink{
		sample: sample,
		header: header,
	}
	if b.first == nil {
		b.first = link
	} else {
		b.last.next = link
	}
	b.last = link
	if b.size >= b.Capacity {
		b.first = b.first.next
	} else {
		b.size++
	}

	b.cond.Broadcast()
}

func (b *outputSampleBuffer) getFirst() (*sampleListLink, uint) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	for b.first == nil && !b.closed {
		b.cond.Wait()
	}
	return b.first, b.size
}

func (b *outputSampleBuffer) next(l *sampleListLink) *sampleListLink {
	if b.closed {
		return nil
	}
	if l.next != nil {
		return l.next
	}
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	for l.next == nil && !b.closed {
		b.cond.Wait()
	}
	return l.next
}

func (b *outputSampleBuffer) close() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	b.closed = true
	b.cond.Broadcast()
}
