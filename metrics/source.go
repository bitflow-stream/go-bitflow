package metrics

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
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

func receiveMetrics(conn *net.TCPConn, sink MetricSink) (err error) {
	defer func() {
		if err != nil {
			if err := conn.Close(); err != nil {
				log.Println("Error closing connection:", err)
			}
		}
	}()
	log.Println("Receiving header from", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	var header Header
	if err = header.ReadBinary(reader); err != nil {
		return
	}
	if err = sink.Header(header); err != nil {
		return
	}
	log.Printf("Receiving %v metrics\n", len(header))

	for {
		var sample Sample
		if err = sample.ReadBinary(reader, len(header)); err != nil {
			return
		}
		if err := sink.Sample(sample); err != nil {
			log.Printf("Error forwarding received sample: %v\n", err)
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
	defer wg.Done()
	log.Println("Listening for incoming metric data on", listener.Addr())
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Println("Error accepting connection:", err)
		} else {
			log.Println("Accepted connection from", conn.RemoteAddr())
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := receiveMetrics(conn, sink); err == io.EOF {
					log.Println("Connection closed by", conn.RemoteAddr())
				} else if err != nil {
					log.Println(err)
				}
			}()
		}
	}
}

// ==================== CSV file source ====================
type CSVFileMetricSource struct {
	Filename string
	file     *os.File
}

func (source *CSVFileMetricSource) Start(wg *sync.WaitGroup, sink MetricSink) error {
	f, err := os.Open(source.Filename)
	if err != nil {
		return err
	}
	source.file = f
	wg.Add(1)
	go source.readFile(wg, sink)
	return nil
}

func (source *CSVFileMetricSource) readFile(wg *sync.WaitGroup, sink MetricSink) {
	defer wg.Done()
	reader := bufio.NewReader(source.file)
	var header Header
	if err := header.ReadCsv(reader); err != nil && err != CSV_EOF {
		log.Println("Failed to read CSV header", err)
		return
	}
	if err := sink.Header(header); err != nil {
		log.Println("Failed to forward CSV header:", err)
		return
	}
	log.Printf("Reading %v metrics from %v\n", len(header), source.file.Name())

	num_samples := 0
	for {
		var sample Sample
		if err := sample.ReadCsv(reader); err == CSV_EOF {
			break
		} else if err != nil {
			log.Println("CSV read failed:", err)
			return
		}
		num_samples++
		if err := sink.Sample(sample); err != nil {
			log.Printf("Error forwarding read sample: %v\n", err)
		}
	}
	log.Printf("Read %v samples from file, %v metrics each\n", num_samples, len(header))
}

func (source *CSVFileMetricSource) Close() error {
	return source.file.Close()
}
