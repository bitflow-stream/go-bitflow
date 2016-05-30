package sample

import (
	"log"
	"net"
	"sync"

	"github.com/antongulenko/golib"
)

// ==================== TCP listener source ====================
type TCPListenerSource struct {
	AbstractUnmarshallingMetricSource
	TCPConnCounter
	Reader SampleReader
	stream *SampleInputStream
	remote net.Addr
	task   *golib.TCPListenerTask
}

func NewTcpListenerSource(endpoint string, reader SampleReader) *TCPListenerSource {
	source := &TCPListenerSource{
		Reader: reader,
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
	return source.task.ExtendedStart(func(addr net.Addr) {
		log.Println("Listening for incoming", source.Unmarshaller, "samples on", addr)
	}, wg)
}

func (source *TCPListenerSource) handleConnection(wg *sync.WaitGroup, conn *net.TCPConn) {
	if source.stream != nil {
		log.Printf("Rejecting connection from %v, already connected to %v.\n", conn.RemoteAddr(), source.remote)
		_ = conn.Close() // Drop error
		return
	}
	if !source.CountConnectionAccepted(conn) {
		return
	}
	log.Println("Accepted connection from", conn.RemoteAddr())
	stream := source.Reader.Open(conn, source.Unmarshaller, source.OutgoingSink)
	source.stream = stream
	source.remote = conn.RemoteAddr()
	wg.Add(1)
	go func() {
		defer func() {
			source.task.IfRunning(source.closeStream)
			wg.Done()
		}()
		stream.ReadTcpSamples(conn, source.isConnectionClosed)
		if !source.CountConnectionClosed() {
			source.Stop()
		}
	}()
}

func (source *TCPListenerSource) isConnectionClosed() bool {
	return source.stream == nil
}

func (source *TCPListenerSource) closeStream() {
	if source.stream != nil {
		stream := source.stream
		source.stream = nil // Make isConnectionClosed() return true
		_ = stream.Close()  // Drop error
	}
}

func (source *TCPListenerSource) Stop() {
	source.task.ExtendedStop(func() {
		source.closeStream()
	})
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
		log.Println("Listening for", sink.Marshaller, "sample output connections on", addr)
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
