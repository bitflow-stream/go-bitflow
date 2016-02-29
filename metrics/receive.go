package metrics

import (
	"fmt"
	"log"
	"net"
)

func ReceiveMetrics(listenAddr string, sink MetricSink) error {
	endpoint, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", endpoint)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection:", err)
		} else {
			go handleConnection(conn, sink)
		}
	}
}

func handleConnection(conn *net.TCPConn, sink MetricSink) {
	log.Println("Accepted connection from", conn.RemoteAddr())
	var buf [metricBytes]byte
	for {
		l, err := conn.Read(buf[:])
		if err == nil && l != len(buf) {
			err = fmt.Errorf("Received %v bytes instead of %v", l, len(buf))
		}
		if err != nil {
			log.Println("Error receiving, closing connection:", err)
			if err := conn.Close(); err != nil {
				log.Println("Error closing connection:", err)
			}
			return
		}
		var metric Metric
		err = metric.UnmarshalBinary(buf[:])
		if err != nil {
			log.Println("Error unmarshalling received metric:", err)
			continue
		}
		if err := sink.CycleStart(); err != nil {
			log.Println("Error sinking received metric (%v): %v\n", metric, err)
		} else {
			if err := sink.Sink(&metric); err != nil {
				log.Println("Error sinking received metric (%v): %v\n", metric, err)
			} else {
				sink.CycleFinish()
			}
		}
	}
}
