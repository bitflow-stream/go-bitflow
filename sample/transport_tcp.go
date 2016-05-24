package sample

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/antongulenko/golib"
)

const (
	tcp_sample_buffer = 50
)

// ==================== TCP write connection ====================
type TcpMetricSink struct {
	AbstractMarshallingMetricSink
	LastHeader Header
}

type TcpWriteConn struct {
	sink      *TcpMetricSink
	remote    net.Addr
	conn      *net.TCPConn
	stream    *SampleOutputStream
	samples   chan Sample
	writerWg  sync.WaitGroup
	closeOnce sync.Once
}

func (sink *TcpMetricSink) writeConn(wg *sync.WaitGroup, conn *net.TCPConn) *TcpWriteConn {
	writeConn := &TcpWriteConn{
		sink:    sink,
		conn:    conn,
		remote:  conn.RemoteAddr(),
		samples: make(chan Sample, tcp_sample_buffer),
		stream:  sink.Writer.OpenBuffered(conn, sink.Marshaller),
	}
	wg.Add(1)
	writeConn.writerWg.Add(1)
	go writeConn.run(wg)
	return writeConn
}

func (conn *TcpWriteConn) Close() {
	if conn != nil {
		conn.closeOnce.Do(func() {
			if conn.Running() {
				log.Println("Closing connection to", conn.remote)
			}
			close(conn.samples)
		})
		conn.writerWg.Wait()
	}
}

func (conn *TcpWriteConn) Running() bool {
	return conn != nil && conn.conn != nil
}

func (conn *TcpWriteConn) printErr(err error) {
	if operr, ok := err.(*net.OpError); ok {
		if operr.Err == syscall.EPIPE {
			log.Printf("Connection closed by %v\n", conn.remote)
			return
		} else {
			if syscallerr, ok := operr.Err.(*os.SyscallError); ok && syscallerr.Err == syscall.EPIPE {
				log.Printf("Connection closed by %v\n", conn.remote)
				return
			}
		}
	}
	if err != nil {
		log.Printf("TCP write to %v failed, closing connection. %v\n", conn.remote, err)
	}
}

func (c *TcpWriteConn) flush(conn *net.TCPConn, reportError bool) {
	if err := c.stream.Close(); err != nil {
		if reportError {
			log.Printf("Error flushing connection to %v: %v\n", c.remote, err)
		}
	}
	if err := conn.Close(); err != nil {
		if reportError {
			log.Printf("Error closing connection to %v: %v\n", c.remote, err)
		}
	}
}

func (conn *TcpWriteConn) run(wg *sync.WaitGroup) {
	connection := conn.conn
	var err error
	defer func() {
		conn.flush(connection, err == nil)
		conn.conn = nil // Make Running() return false
		wg.Done()
		conn.writerWg.Done()
	}()

	log.Println("Serving", len(conn.sink.LastHeader.Fields), "metrics to", conn.remote)
	if err = conn.stream.Header(conn.sink.LastHeader); err != nil {
		conn.printErr(err)
		return
	}
	for sample := range conn.samples {
		if err = conn.stream.Sample(sample); err != nil {
			conn.printErr(err)
			return
		}
	}
}

// ==================== TCP active sink ====================
type TCPSink struct {
	TcpMetricSink
	Endpoint string
	wg       *sync.WaitGroup
	conn     *TcpWriteConn
	stopped  *golib.OneshotCondition
}

func (sink *TCPSink) String() string {
	return "TCP sink to " + sink.Endpoint
}

func (sink *TCPSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Sending", sink.Marshaller, "samples to", sink.Endpoint)
	sink.stopped = golib.NewOneshotCondition()
	sink.wg = wg
	return sink.stopped.Start(wg)
}

func (sink *TCPSink) closeConnection() {
	sink.conn.Close()
	sink.conn = nil
}

func (sink *TCPSink) Close() {
	sink.stopped.Enable(func() {
		sink.closeConnection()
	})
}

func (sink *TCPSink) Header(header Header) (err error) {
	sink.stopped.IfElseEnabled(func() {
		err = fmt.Errorf("TCP sink to %v already stopped", sink.Endpoint)
	}, func() {
		sink.closeConnection() // Stop existing connection to negotiate new header
		sink.LastHeader = header
		err = sink.assertConnection()
	})
	return
}

func (sink *TCPSink) Sample(sample Sample, header Header) (err error) {
	sink.stopped.IfElseEnabled(func() {
		err = fmt.Errorf("TCP sink to %v already stopped", sink.Endpoint)
	}, func() {
		if err = sample.Check(header); err != nil {
			return
		}
		if !sink.conn.Running() {
			sink.closeConnection() // Cleanup errored connection
		}
		if err = sink.assertConnection(); err != nil {
			return
		}
		sink.conn.samples <- sample
	})
	return
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
		sink.conn = sink.writeConn(sink.wg, conn)
	}
	return nil
}

// ==================== TCP active source ====================
type TCPSource struct {
	AbstractUnmarshallingMetricSource
	RemoteAddr    string
	RetryInterval time.Duration
	Reader        SampleReader
	loopTask      *golib.LoopTask
	conn          *net.TCPConn
}

func (sink *TCPSource) String() string {
	return "TCP source from " + sink.RemoteAddr
}

func (source *TCPSource) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Downloading", source.Unmarshaller, "data from", source.RemoteAddr)
	source.loopTask = golib.NewLoopTask("tcp download source", func(stop golib.StopChan) {
		if conn, err := source.dial(); err != nil {
			log.Println("Error downloading data:", err)
		} else {
			source.loopTask.IfElseEnabled(func() {
				return
			}, func() {
				source.conn = conn
			})
			source.Reader.ReadTcpSamples(conn, source.Unmarshaller, source.OutgoingSink, source.connectionClosed)
		}
		select {
		case <-time.After(source.RetryInterval):
		case <-stop:
		}
	})
	source.loopTask.StopHook = func() {
		source.CloseSink(wg)
	}
	return source.loopTask.Start(wg)
}

func (source *TCPSource) Stop() {
	source.loopTask.Enable(func() {
		if conn := source.conn; conn != nil {
			_ = conn.Close() // Ignore error
		}
	})
}

func (source *TCPSource) connectionClosed() bool {
	return source.loopTask.Enabled()
}

func (source *TCPSource) dial() (*net.TCPConn, error) {
	endpoint, err := net.ResolveTCPAddr("tcp", source.RemoteAddr)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, endpoint)
}
