package metrics

import (
	"bufio"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type MetricSource interface {
	Start(wg *sync.WaitGroup, sink MetricSink) error
}

// ==================== Aggregate source ====================
type AggregateSource []MetricSource

func (agg AggregateSource) Start(wg *sync.WaitGroup, sink MetricSink) error {
	for _, source := range agg {
		if err := source.Start(wg, sink); err != nil {
			return err
		}
	}
	return nil
}

// ==================== TCP active source ====================
type DownloadMetricSource struct {
	RemoteAddr    string
	RetryInterval time.Duration
}

func (source *DownloadMetricSource) Start(wg *sync.WaitGroup, sink MetricSink) error {
	wg.Add(1)
	go source.download(wg, sink)
	return nil
}

func (source *DownloadMetricSource) download(wg *sync.WaitGroup, sink MetricSink) {
	defer wg.Done()
	log.Println("Downloading data from", source.RemoteAddr)
	for {
		if conn, err := source.dial(); err != nil {
			log.Println("Error downloading data:", err)
		} else {
			if err := receiveMetrics(conn, sink); err == io.EOF {
				log.Println("Connection closed by", conn.RemoteAddr())
			} else if err != nil {
				log.Printf("Error downloading metrics from %v: %v\n", conn.RemoteAddr(), err)
			}
		}
		time.Sleep(source.RetryInterval)
	}
}

func (source *DownloadMetricSource) dial() (*net.TCPConn, error) {
	endpoint, err := net.ResolveTCPAddr("tcp", source.RemoteAddr)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, endpoint)
}

func receiveMetrics(conn *net.TCPConn, sink MetricSink) error {
	reader := bufio.NewReader(conn)
	for {
		var metric Metric
		err := metric.ReadFrom(reader)
		if err != nil {
			if err := conn.Close(); err != nil {
				log.Println("Error closing connection:", err)
			}
			return err
		}
		if err := sink.CycleStart(); err != nil {
			log.Printf("Error sinking received metric (%v): %v\n", &metric, err)
		} else {
			if err := sink.Sink(&metric); err != nil {
				log.Printf("Error sinking received metric (%v): %v\n", metric, err)
			} else {
				sink.CycleFinish()
			}
		}
	}
}

// ==================== TCP passive source ====================
type ReceiveMetricSource struct {
	ListenEndpoint string
}

func (source *ReceiveMetricSource) Start(wg *sync.WaitGroup, sink MetricSink) error {
	endpoint, err := net.ResolveTCPAddr("tcp", source.ListenEndpoint)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", endpoint)
	if err != nil {
		return err
	}
	wg.Add(1)
	go source.listen(wg, listener, sink)
	return nil
}

func (source *ReceiveMetricSource) listen(wg *sync.WaitGroup, listener *net.TCPListener, sink MetricSink) {
	log.Println("Listening for incoming metric data on", listener.Addr())
	defer wg.Done()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection:", err)
		} else {
			log.Println("Accepted connection from", conn.RemoteAddr())
			go func() {
				if err := receiveMetrics(conn, sink); err == io.EOF {
					log.Println("Connection closed by", conn.RemoteAddr())
				} else if err != nil {
					log.Println(err)
				}
			}()
		}
	}
}
