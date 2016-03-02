package metrics

import (
	"fmt"
	"log"
	"net"
	"sync"
	"syscall"
)

const (
	max_send_retries = 10
)

type MetricSink interface {
	Start(wg *sync.WaitGroup) error

	Sink(metric *Metric) error
	CycleStart() error
	CycleFinish()
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

func (agg AggregateSink) Sink(metric *Metric) error {
	for _, sink := range agg {
		if err := sink.Sink(metric); err != nil {
			return err
		}
	}
	return nil
}

func (agg AggregateSink) CycleStart() error {
	for _, sink := range agg {
		if err := sink.CycleStart(); err != nil {
			return err
		}
	}
	return nil
}

func (agg AggregateSink) CycleFinish() {
	for _, sink := range agg {
		sink.CycleFinish()
	}
}

// ==================== Console ====================
type ConsoleSink struct {
}

func (sink *ConsoleSink) Start(wg *sync.WaitGroup) error {
	return nil
}

func (sink *ConsoleSink) Sink(metric *Metric) error {
	fmt.Printf(" %v", metric)
	return nil
}

func (sink *ConsoleSink) CycleStart() error {
	log.Printf("")
	return nil
}

func (sink *ConsoleSink) CycleFinish() {
	fmt.Printf("\n")
}

// ==================== TCP active ====================
type ActiveTcpSink struct {
	Endpoint string
	conn     *net.TCPConn
}

func (sink *ActiveTcpSink) Start(wg *sync.WaitGroup) error {
	return nil
}

func (sink *ActiveTcpSink) CycleStart() error {
	return sink.assertConnection()
}

func (sink *ActiveTcpSink) CycleFinish() {
}

func (sink *ActiveTcpSink) Sink(metric *Metric) error {
	if err := sink.assertConnection(); err != nil {
		return err
	}

	if err := metric.WriteTo(sink.conn); err != nil {
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
	Endpoint    string
	connections map[*passiveTcpConn]bool
}

type passiveTcpConn struct {
	conn    *net.TCPConn
	metrics chan Metric
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
				metrics: make(chan Metric, tcp_metric_buffer),
			}
			wg.Add(1)
			go passiveConn.run(wg)
			sink.connections[passiveConn] = true
		}
	}
}

func (sink *PassiveTcpSink) CycleStart() error {
	return nil
}

func (sink *PassiveTcpSink) CycleFinish() {
}

func (sink *PassiveTcpSink) Sink(metric *Metric) error {
	for conn, _ := range sink.connections {
		if conn.conn == nil {
			close(conn.metrics)
			delete(sink.connections, conn)
			continue
		}
		conn.metrics <- *metric // Push a copy of the metric
	}
	return nil
}

func (conn *passiveTcpConn) run(wg *sync.WaitGroup) {
	defer func() {
		conn.conn = nil // In case of panic, avoid full channel-buffers
		wg.Done()
	}()
	log.Println("Accepted connection from", conn.conn.RemoteAddr())
	for metric := range conn.metrics {
		connection := conn.conn
		if connection == nil {
			break
		}
		if err := metric.WriteTo(connection); err != nil {
			conn.conn = nil
			if err == syscall.EPIPE {
				log.Printf("Connection to %v closed\n", connection.RemoteAddr())
			} else {
				log.Printf("TCP write to %v failed, closing connection. %v\n", connection.RemoteAddr(), err)
			}
			if err := connection.Close(); err != nil {
				log.Println("Error closing connection with %v:", connection.RemoteAddr(), err)
			}
		}
	}
}
