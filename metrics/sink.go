package metrics

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
)

type MetricSink interface {
	Start(wg *sync.WaitGroup) error
	Header(header Header) error
	Sample(sample Sample) error
}

type abstractSink struct {
	header Header
}

func (sink *abstractSink) Header(header Header) error {
	sink.header = header
	return nil
}

func (sink *abstractSink) checkSample(sample Sample) error {
	if len(sample.Values) != len(sink.header) {
		return fmt.Errorf("Unexpected number of values in sample: %v, expected %v", len(sample.Values), len(sink.header))
	}
	return nil
}

// ==================== Aggregating sink ====================
type AggregateSink []MetricSink

func (agg AggregateSink) Start(wg *sync.WaitGroup) error {
	for _, sink := range agg {
		if err := sink.Start(wg); err != nil {
			return err
		}
	}
	return nil
}

func (agg AggregateSink) Header(header Header) error {
	for _, sink := range agg {
		if err := sink.Header(header); err != nil {
			return err
		}
	}
	return nil
}

func (agg AggregateSink) Sample(sample Sample) error {
	for _, sink := range agg {
		if err := sink.Sample(sample); err != nil {
			return err
		}
	}
	return nil
}

// ==================== Console ====================
type ConsoleSink struct {
	abstractSink
	tags []string
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup) error {
	return nil
}

func (sink *ConsoleSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	timeStr := sample.Time.Format("2006-01-02 15:04:05.999")
	fmt.Printf("%s: ", timeStr)
	for i, value := range sample.Values {
		fmt.Printf("%s = %.4f", sink.tags[i], value)
		if i < len(sample.Values)-1 {
			fmt.Printf(", ")
		}
	}
	fmt.Println()
	return nil
}

// ==================== TCP active ====================
type ActiveTcpSink struct {
	abstractSink
	Endpoint string
	conn     *net.TCPConn
}

func (sink *ActiveTcpSink) Start(wg *sync.WaitGroup) error {
	return nil
}

func (sink *ActiveTcpSink) Header(header Header) error {
	if err := sink.assertConnection(); err != nil {
		return err
	}
	if err := sink.abstractSink.Header(header); err != nil {
		return err
	}
	return sink.checkResult(header.WriteBinary(sink.conn))
}

func (sink *ActiveTcpSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	if err := sink.assertConnection(); err != nil {
		return err
	}
	return sink.checkResult(sample.WriteBinary(sink.conn))
}

func (sink *ActiveTcpSink) checkResult(err error) error {
	if err != nil {
		log.Println("TCP write failed, closing connection.", err)
		if err := sink.conn.Close(); err != nil {
			log.Println("Error closing connection:", err)
		}
		sink.conn = nil
		return err
	}
	return nil
}

func (sink *ActiveTcpSink) assertConnection() error {
	if sink.conn == nil {
		endpoint, err := net.ResolveTCPAddr("tcp", sink.Endpoint)
		if err != nil {
			return err
		}
		sink.conn, err = net.DialTCP("tcp", nil, endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

// ==================== TCP passive ====================
const (
	tcp_metric_buffer = 50
)

type PassiveTcpSink struct {
	abstractSink
	Endpoint    string
	connections map[*passiveTcpConn]bool
}

type passiveTcpConn struct {
	sink    *PassiveTcpSink
	conn    *net.TCPConn
	samples chan Sample
}

func (sink *PassiveTcpSink) Start(wg *sync.WaitGroup) error {
	if sink.connections == nil {
		sink.connections = make(map[*passiveTcpConn]bool)
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

func (sink *PassiveTcpSink) listen(wg *sync.WaitGroup, listener *net.TCPListener) {
	defer wg.Done()
	log.Println("Listening for data sink connections on", listener.Addr())
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection:", err)
		} else {
			passiveConn := &passiveTcpConn{
				conn:    conn,
				samples: make(chan Sample, tcp_metric_buffer),
			}
			wg.Add(1)
			go passiveConn.run(wg)
			sink.connections[passiveConn] = true
		}
	}
}

func (sink *PassiveTcpSink) Header(header Header) error {
	if err := sink.abstractSink.Header(header); err != nil {
		return err
	}
	// Close all running connections, since we have to negotiate a new header.
	for conn := range sink.connections {
		conn.stop(nil)
	}
	return nil
}

func (sink *PassiveTcpSink) Sample(sample Sample) error {
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

func (conn *passiveTcpConn) stop(err error) {
	connection := conn.conn
	if connection != nil {
		conn.conn = nil
		if err == syscall.EPIPE {
			log.Printf("Connection to %v closed\n", connection.RemoteAddr())
		} else if err != nil {
			log.Printf("TCP write to %v failed, closing connection. %v\n", connection.RemoteAddr(), err)
		} else {
			log.Println("Closing connection to", connection.RemoteAddr())
		}
		if err := connection.Close(); err != nil {
			log.Printf("Error closing connection to %v: %v\n", connection.RemoteAddr(), err)
		}
	}
}

func (conn *passiveTcpConn) run(wg *sync.WaitGroup) {
	defer func() {
		conn.conn = nil // In case of panic, avoid full channel-buffers
		wg.Done()
	}()
	log.Printf("Accepted connection from %v, serving %v metrics\n", conn.conn.RemoteAddr(), len(conn.sink.header))
	if err := conn.sink.header.WriteBinary(conn.conn); err != nil {
		conn.stop(err)
		return
	}
	for sample := range conn.samples {
		connection := conn.conn
		if connection == nil {
			break
		}
		if err := sample.WriteBinary(connection); err != nil {
			conn.stop(err)
			break
		}
	}
}

// ==================== CSV file sink ====================
type CSVFileSink struct {
	abstractSink
	Filename string
	file     *os.File
}

func (sink *CSVFileSink) Start(wg *sync.WaitGroup) error {
	f, err := os.OpenFile(sink.Filename, os.O_WRONLY, os.FileMode(0664))
	if err != nil {
		return err
	}
	sink.file = f
	return nil
}

func (sink *CSVFileSink) Header(header Header) error {
	if err := sink.abstractSink.Header(header); err != nil {
		return err
	}
	return header.WriteCsv(sink.file)
}

func (sink *CSVFileSink) Sample(sample Sample) error {
	if err := sink.checkSample(sample); err != nil {
		return err
	}
	return sample.WriteCsv(sink.file)
}

func (sink *CSVFileSink) Close() error {
	return sink.file.Close()
}
