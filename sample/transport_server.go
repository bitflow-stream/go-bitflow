package sample

import (
	"log"
	"net"
	"sync"

	"github.com/antongulenko/golib"
)

// ==================== TCP listener source ====================
type TCPListenerSource struct {
	unmarshallingMetricSource
	conn *net.TCPConn
	task *golib.TCPListenerTask
}

func NewTcpListenerSource(endpoint string) MetricSource {
	source := &TCPListenerSource{}
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
	return source.task.ExtendedStart(func(addr net.Addr) {
		log.Println("Listening for incoming", source.Unmarshaller, "samples on", addr)
	}, wg)
}

func (source *TCPListenerSource) handleConnection(wg *sync.WaitGroup, conn *net.TCPConn) {
	if existing := source.conn; existing != nil {
		log.Printf("Rejecting connection from %v, already connected to %v\n", conn.RemoteAddr(), existing.RemoteAddr())
		_ = conn.Close() // Drop error
		return
	}
	log.Println("Accepted connection from", conn.RemoteAddr())
	source.conn = conn
	wg.Add(1)
	go func() {
		defer func() {
			source.conn = nil
			wg.Done()
		}()
		tcpReadSamples(conn, source.Unmarshaller, source.Sink, source.connectionClosed)
	}()
}

func (source *TCPListenerSource) connectionClosed() bool {
	return source.conn == nil
}

func (source *TCPListenerSource) Stop() {
	source.task.ExtendedStop(func() {
		if conn := source.conn; conn != nil {
			source.conn = nil
			_ = conn.Close() // Drop error
		}
	})
}

// ==================== TCP listener sink ====================
type TCPListenerSink struct {
	abstractSink
	connections map[*tcpWriteConn]bool
	task        *golib.TCPListenerTask
}

func NewTcpListenerSink(endpoint string) MarshallingMetricSink {
	sink := &TCPListenerSink{
		connections: make(map[*tcpWriteConn]bool),
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
		log.Println("Listening for", sink.marshaller, "sample output connections on", addr)
	}, wg)
}

func (sink *TCPListenerSink) Stop() {
	sink.task.ExtendedStop(func() {
		for conn := range sink.connections {
			conn.Stop()
		}
	})
}

func (sink *TCPListenerSink) handleConnection(wg *sync.WaitGroup, conn *net.TCPConn) {
	writeConn := sink.writeConn(conn)
	wg.Add(1)
	go writeConn.Run(wg)
	sink.connections[writeConn] = true
}

func (sink *TCPListenerSink) Header(header Header) error {
	sink.header = header
	// Close all running connections, since we have to negotiate a new header.
	for conn := range sink.connections {
		conn.Stop()
	}
	return nil
}

func (sink *TCPListenerSink) Sample(sample Sample, header Header) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	for conn, _ := range sink.connections {
		if conn.conn == nil {
			// Clean up closed connections
			delete(sink.connections, conn)
			conn.Stop()
			continue
		}
		conn.samples <- sample
	}
	return nil
}
