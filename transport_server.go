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
	AbstractMetricSource

	// TCPConnCounter has a configuration for limiting the total number of
	// accepted connections. After that number of connections were accepted, no
	// further connections are accepted. After they all are closed, the TCPListenerSource
	// automatically stops.
	TCPConnCounter

	// Reader configured aspects of parallel reading and parsing. See SampleReader for more info.
	// This field should not be accessed directly, as it will be passed along
	// when calling NewTcpListenerSource().
	Reader SampleReader

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
// machine. The reader parameter will be set as Reader in the new TCPListenerSource.
func NewTcpListenerSource(endpoint string, reader SampleReader) *TCPListenerSource {
	source := &TCPListenerSource{
		Reader:      reader,
		connections: make(map[*tcpListenerConnection]bool),
	}
	source.task = &golib.TCPListenerTask{
		ListenEndpoint: endpoint,
		Handler:        source.handleConnection,
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

	connections map[*TcpWriteConn]bool
	task        *golib.TCPListenerTask
}

// NewTcpListenerSink creates a new TCPListener sink listening on the given local
// TCP endpoint. It must be a combination of IP or hostname with a port number, that
// can be bound to locally.
func NewTcpListenerSink(endpoint string) *TCPListenerSink {
	sink := &TCPListenerSink{
		connections: make(map[*TcpWriteConn]bool),
	}
	sink.task = &golib.TCPListenerTask{
		ListenEndpoint: endpoint,
		Handler:        sink.handleConnection,
	}
	return sink
}

// String implements the MetrincSink interface.
func (sink *TCPListenerSink) String() string {
	return "TCP sink on " + sink.task.ListenEndpoint
}

// Start implements the MetricSink interface. It creates the TCP socket and
// starts listening on it in a separate goroutine. Any incoming connection is
// then handled in their own goroutine.
func (sink *TCPListenerSink) Start(wg *sync.WaitGroup) golib.StopChan {
	return sink.task.ExtendedStart(func(addr net.Addr) {
		log.WithField("format", sink.Marshaller).Println("Listening for output connections on", addr)
	}, wg)
}

// Close implements the MetricSink interface. It closes any existing connection
// and closes the TCP socket.
func (sink *TCPListenerSink) Close() {
	sink.task.ExtendedStop(func() {
		for conn := range sink.connections {
			conn.Close()
		}
	})
}

func (sink *TCPListenerSink) handleConnection(_ *sync.WaitGroup, conn *net.TCPConn) {
	if !sink.countConnectionAccepted(conn) {
		return
	}
	writeConn := sink.OpenWriteConn(conn)
	sink.connections[writeConn] = true
}

// Samples implements the MetricSink interface by sending the Sample on all currently
// established TCP connections. If no connection is currently established, the sample
// is simply dropped.
func (sink *TCPListenerSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	for conn := range sink.connections {
		if conn.stream == nil {
			// Clean up closed connections
			sink.cleanupConnection(conn)
			continue
		}
		conn.Sample(sample, header)
	}
	return nil
}

func (sink *TCPListenerSink) cleanupConnection(conn *TcpWriteConn) {
	delete(sink.connections, conn)
	conn.Close()
	if !sink.countConnectionClosed() {
		sink.Close()
	}
}

// ======================================= output sample buffer =======================================

// TODO finish implementation: add optional buffer that is delivered to every
// new connection

type outputSampleBuffer struct {
	capacity int
	size     int
	first    *sampleListLink
	last     *sampleListLink
	cond     sync.Cond
}

type sampleListLink struct {
	sample *Sample
	next   *sampleListLink
}

func (l *outputSampleBuffer) add(sample *Sample) {
	link := &sampleListLink{
		sample: sample,
	}
	if l.first == nil {
		l.first = link
	} else {
		l.last.next = link
	}
	l.last = link
	if l.size >= l.capacity {
		l.first = l.first.next
	} else {
		l.size++
	}
}

func (b *outputSampleBuffer) next(l *sampleListLink) *sampleListLink {
	if l.next != nil {
		return l.next
	}
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	for l.next == nil {
		b.cond.Wait()
	}
	return l.next
}
