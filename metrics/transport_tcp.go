package metrics

import (
	"io"
	"log"
	"net"
	"sync"
	"syscall"
	"time"
)

const (
	tcp_sample_buffer = 50
)

// ==================== TCP write connection ====================
type tcpWriteConn struct {
	sink    *abstractSink
	remote  net.Addr
	conn    *net.TCPConn
	samples chan Sample
}

func (sink *abstractSink) writeConn(conn *net.TCPConn) *tcpWriteConn {
	return &tcpWriteConn{
		sink:    sink,
		conn:    conn,
		remote:  conn.RemoteAddr(),
		samples: make(chan Sample, tcp_sample_buffer),
	}
}

func (conn *tcpWriteConn) stop(err error) {
	connection := conn.conn
	if connection != nil {
		conn.conn = nil
		if operr, ok := err.(*net.OpError); ok && operr.Err == syscall.EPIPE {
			log.Printf("Connection to %v closed\n", conn.remote)
		} else if err != nil {
			log.Printf("TCP write to %v failed, closing connection. %v\n", conn.remote, err)
		} else {
			log.Println("Closing connection to", conn.remote)
		}
		if err := connection.Close(); err != nil {
			log.Printf("Error closing connection to %v: %v\n", conn.remote, err)
		}
	}
}

func (conn *tcpWriteConn) run(wg *sync.WaitGroup) {
	defer func() {
		conn.conn = nil // In case of panic, avoid full channel-buffer
		wg.Done()
	}()
	log.Println("Serving", len(conn.sink.header), "metrics to", conn.remote)
	if err := conn.sink.marshaller.WriteHeader(conn.sink.header, conn.conn); err != nil {
		conn.stop(err)
		return
	}
	for sample := range conn.samples {
		connection := conn.conn
		if connection == nil {
			break
		}
		if err := conn.sink.marshaller.WriteSample(sample, connection); err != nil {
			conn.stop(err)
			break
		}
	}
}

// ==================== TCP active sink ====================
type TCPSink struct {
	abstractSink
	Endpoint string
	wg       *sync.WaitGroup
	conn     *tcpWriteConn
}

func (sink *TCPSink) Start(wg *sync.WaitGroup, marshaller Marshaller) error {
	log.Println("Sending", marshaller, "samples to", sink.Endpoint)
	sink.marshaller = marshaller
	sink.wg = wg
	return nil
}

func (sink *TCPSink) Header(header Header) error {
	if sink.conn != nil {
		// Stop existing connection to negotiate new header
		sink.conn.stop(nil)
	}
	sink.header = header
	return sink.assertConnection()
}

func (sink *TCPSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	connection := sink.conn
	if connection != nil && connection.conn == nil {
		sink.conn = nil
		close(connection.samples)
	}
	if err := sink.assertConnection(); err != nil {
		return err
	}
	sink.conn.samples <- sample
	return nil
}

func (sink *TCPSink) assertConnection() error {
	if sink.conn == nil {
		endpoint, err := net.ResolveTCPAddr("tcp", sink.Endpoint)
		if err != nil {
			return err
		}
		conn, err := net.DialTCP("tcp", nil, endpoint)
		if err != nil {
			return err
		}
		sink.conn = sink.writeConn(conn)
		sink.wg.Add(1)
		go sink.conn.run(sink.wg)
	}
	return nil
}

// ==================== TCP active source ====================
type TCPSource struct {
	unmarshallingMetricSource
	RemoteAddr    string
	RetryInterval time.Duration
}

func (source *TCPSource) Start(wg *sync.WaitGroup, sink MetricSink) error {
	wg.Add(1)
	go source.download(wg, sink)
	return nil
}

func (source *TCPSource) download(wg *sync.WaitGroup, sink MetricSink) {
	defer wg.Done()
	log.Println("Downloading", source.Unmarshaller, "data from", source.RemoteAddr)
	for {
		if conn, err := source.dial(); err != nil {
			log.Println("Error downloading data:", err)
		} else {
			tcpReadSamples(conn, source.Unmarshaller, sink)
		}
		time.Sleep(source.RetryInterval)
	}
}

func (source *TCPSource) dial() (*net.TCPConn, error) {
	endpoint, err := net.ResolveTCPAddr("tcp", source.RemoteAddr)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, endpoint)
}

func tcpReadSamples(conn *net.TCPConn, um Unmarshaller, sink MetricSink) {
	log.Println("Receiving header from", conn.RemoteAddr())
	var err error
	var num_samples int
	if num_samples, err = readSamples(conn, um, sink); err == io.EOF {
		log.Println("Connection closed by", conn.RemoteAddr())
	} else if err != nil {
		log.Printf("Error receiving samples from %v: %v\n", conn.RemoteAddr(), err)
		if err := conn.Close(); err != nil {
			log.Println("Error closing connection:", err)
		}
	}
	log.Println("Received", num_samples, "samples from", conn.RemoteAddr())
}
