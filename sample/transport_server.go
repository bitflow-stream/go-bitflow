package sample

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/golib"
)

// ==================== TCP listener source ====================
type TCPListenerSource struct {
	AbstractMetricSource
	TCPConnCounter
	Reader                  SampleReader
	SimultaneousConnections uint

	task             *golib.TCPListenerTask
	synchronizedSink MetricSinkBase
	connections      map[*TCPListenerConnection]bool
}

func NewTcpListenerSource(endpoint string, reader SampleReader) *TCPListenerSource {
	source := &TCPListenerSource{
		Reader:      reader,
		connections: make(map[*TCPListenerConnection]bool),
	}
	source.task = &golib.TCPListenerTask{
		ListenEndpoint: endpoint,
		Handler:        source.handleConnection,
	}
	return source
}

func (source *TCPListenerSource) String() string {
	return "TCP source on " + source.task.ListenEndpoint
}

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
	if !source.CountConnectionAccepted(conn) {
		return
	}
	log.WithField("remote", conn.RemoteAddr()).Println("Accepted connection")
	listenerConn := &TCPListenerConnection{
		source: source,
		stream: source.Reader.Open(conn, source.synchronizedSink),
	}
	source.connections[listenerConn] = true
	wg.Add(1)
	go listenerConn.readSamples(wg, conn)
}

func (source *TCPListenerSource) Stop() {
	source.task.ExtendedStop(func() {
		for conn := range source.connections {
			conn.closeStream()
		}
	})
}

type TCPListenerConnection struct {
	source *TCPListenerSource
	stream *SampleInputStream
}

func (conn *TCPListenerConnection) isConnectionClosed() bool {
	return conn.stream == nil
}

func (conn *TCPListenerConnection) readSamples(wg *sync.WaitGroup, connection *net.TCPConn) {
	defer wg.Done()
	conn.stream.ReadTcpSamples(connection, conn.isConnectionClosed)
	if !conn.source.CountConnectionClosed() {
		conn.source.Stop()
	}
	conn.source.task.IfRunning(conn.closeStream)
}

func (conn *TCPListenerConnection) closeStream() {
	if stream := conn.stream; stream != nil {
		conn.stream = nil  // Make isConnectionClosed() return true
		_ = stream.Close() // Drop error
		delete(conn.source.connections, conn)
	}
}

// ==================== TCP listener sink ====================
type TCPListenerSink struct {
	TcpMetricSink
	connections map[*TcpWriteConn]bool
	task        *golib.TCPListenerTask
	lastHeader  Header
	headerLock  sync.Mutex
}

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

func (sink *TCPListenerSink) String() string {
	return "TCP sink on " + sink.task.ListenEndpoint
}

func (sink *TCPListenerSink) Start(wg *sync.WaitGroup) golib.StopChan {
	return sink.task.ExtendedStart(func(addr net.Addr) {
		log.WithField("format", sink.Marshaller).Println("Listening for output connections on", addr)
	}, wg)
}

func (sink *TCPListenerSink) Close() {
	sink.task.ExtendedStop(func() {
		for conn := range sink.connections {
			conn.Close()
		}
	})
}

func (sink *TCPListenerSink) handleConnection(_ *sync.WaitGroup, conn *net.TCPConn) {
	if !sink.CountConnectionAccepted(conn) {
		return
	}
	writeConn, header := sink.openTcpOutputStream(conn)
	if writeConn != nil {
		writeConn.Header(header)
	}
}

func (sink *TCPListenerSink) openTcpOutputStream(conn *net.TCPConn) (*TcpWriteConn, Header) {
	sink.headerLock.Lock()
	defer sink.headerLock.Unlock()
	writeConn := sink.OpenWriteConn(conn)
	sink.connections[writeConn] = true
	return writeConn, sink.lastHeader
}

func (sink *TCPListenerSink) Header(header Header) error {
	sink.headerLock.Lock()
	defer sink.headerLock.Unlock()
	sink.lastHeader = header
	// Close all running connections, since we have to negotiate a new header.
	for conn := range sink.connections {
		sink.cleanupConnection(conn)
	}
	return nil
}

func (sink *TCPListenerSink) Sample(sample Sample, header Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	for conn := range sink.connections {
		if conn.stream == nil {
			// Clean up closed connections
			sink.cleanupConnection(conn)
			continue
		}
		conn.Sample(sample)
	}
	return nil
}

func (sink *TCPListenerSink) cleanupConnection(conn *TcpWriteConn) {
	delete(sink.connections, conn)
	conn.Close()
	if !sink.CountConnectionClosed() {
		sink.Close()
	}
}
