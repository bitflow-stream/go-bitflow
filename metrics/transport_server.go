package metrics

import (
	"log"
	"net"
	"sync"
)

// ==================== TCP listener source ====================
type TCPListenerSource struct {
	ListenEndpoint string
}

func (source *TCPListenerSource) Start(wg *sync.WaitGroup, um Unmarshaller, sink MetricSink) error {
	endpoint, err := net.ResolveTCPAddr("tcp", source.ListenEndpoint)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", endpoint)
	if err != nil {
		return err
	}
	wg.Add(1)
	go source.listen(wg, listener, um, sink)
	return nil
}

func (source *TCPListenerSource) listen(wg *sync.WaitGroup, listener *net.TCPListener, um Unmarshaller, sink MetricSink) {
	defer wg.Done()
	log.Println("Listening for incoming metric data on", listener.Addr())
	var connected net.Addr
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection:", err)
		} else {
			if connected != nil {
				log.Println("Rejecting connection from %v, already connected to %v\n", conn.RemoteAddr(), connected)
				if err := conn.Close(); err != nil {
					log.Println("Error closing connection:", err)
				}
				continue
			}
			connected = conn.RemoteAddr()
			log.Println("Accepted connection from", connected)
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					connected = nil
				}()
				tcpRead(conn, um, sink)
			}()
		}
	}
}

// ==================== TCP listener sink ====================
type TCPListenerSink struct {
	abstractSink
	Endpoint    string
	connections map[*tcpWriteConn]bool
}

func (sink *TCPListenerSink) Start(wg *sync.WaitGroup, marshaller Marshaller) error {
	sink.marshaller = marshaller
	if sink.connections == nil {
		sink.connections = make(map[*tcpWriteConn]bool)
	}

	addr, err := net.ResolveTCPAddr("tcp", sink.Endpoint)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	wg.Add(1)
	go sink.listen(wg, listener)
	return nil
}

func (sink *TCPListenerSink) listen(wg *sync.WaitGroup, listener *net.TCPListener) {
	defer wg.Done()
	log.Println("Listening for data sink connections on", listener.Addr())
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection:", err)
		} else {
			writeConn := sink.writeConn(conn)
			wg.Add(1)
			go writeConn.run(wg)
			sink.connections[writeConn] = true
		}
	}
}

func (sink *TCPListenerSink) Header(header Header) error {
	sink.header = header
	// Close all running connections, since we have to negotiate a new header.
	for conn := range sink.connections {
		conn.stop(nil)
	}
	return nil
}

func (sink *TCPListenerSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	for conn, _ := range sink.connections {
		if conn.conn == nil {
			// Clean up closed connections
			delete(sink.connections, conn)
			close(conn.samples)
			continue
		}
		conn.samples <- sample
	}
	return nil
}
