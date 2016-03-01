package metrics

import (
	"bufio"
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
	reader := bufio.NewReader(conn)
	for {
		var metric Metric
		err := metric.ReadFrom(reader)
		if err != nil {
			log.Println("Error receiving metric, closing connection:", err)
			if err := conn.Close(); err != nil {
				log.Println("Error closing connection:", err)
			}
			return
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
