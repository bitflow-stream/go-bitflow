package metrics

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	max_send_retries = 10
)

type MetricSink interface {
	Sink(metric *Metric) error
	CycleStart() error
	CycleFinish()
}

func SinkMetrics(metrics []*Metric, target MetricSink, interval time.Duration) {
	for {
		if err := target.CycleStart(); err != nil {
			log.Printf("Warning: Failed to sink %v metrics: %v\n", len(metrics), err)
		} else {
			for _, metric := range metrics {
				err := target.Sink(metric)
				if err != nil {
					log.Printf("Warning: Failed to sink metric %v: %v\n", metric, err)
				}
			}
			target.CycleFinish()
		}
		time.Sleep(interval)
	}
}

// ==================== Console ====================
type ConsoleSink struct {
}

func (sink *ConsoleSink) Sink(metric *Metric) error {
	fmt.Printf(" %v = %.4f", metric.Name, metric.Val)
	return nil
}

func (sink *ConsoleSink) CycleStart() error {
	log.Printf("")
	return nil
}

func (sink *ConsoleSink) CycleFinish() {
	fmt.Printf("\n")
}

// ==================== TCP ====================
type TcpSink struct {
	Endpoint string
	conn     *net.TCPConn
}

func (sink *TcpSink) CycleStart() error {
	return sink.assertConnection()
}

func (sink *TcpSink) CycleFinish() {
}

func (sink *TcpSink) Sink(metric *Metric) error {
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

func (sink *TcpSink) assertConnection() error {
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
