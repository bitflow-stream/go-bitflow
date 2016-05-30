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

// ==================== TCP connection counter ====================
type TCPConnCounter struct {
	closed       uint
	accepted     uint
	TcpConnLimit uint
}

// Return true, if we can handle further connections. Return false, if the application should stop.
func (counter *TCPConnCounter) CountConnectionClosed() bool {
	if counter.TcpConnLimit > 0 {
		counter.closed++
		if counter.closed >= counter.TcpConnLimit {
			log.Println("Handled", counter.closed, "TCP connection(s)")
			return false
		}
	}
	return true
}

// Return true, if the connection was accepted. Return false, if it was rejected and closed.
func (counter *TCPConnCounter) CountConnectionAccepted(conn *net.TCPConn) bool {
	if counter.TcpConnLimit > 0 {
		if counter.accepted >= counter.TcpConnLimit {
			log.Printf("Rejecting connection from %v, already accepted %v connections\n", conn.RemoteAddr(), counter.accepted)
			_ = conn.Close() // Drop error
			return false
		}
		counter.accepted++
	}
	return true
}

// ==================== TCP write connection ====================
type TcpMetricSink struct {
	AbstractMarshallingMetricSink
	TCPConnCounter
}

type TcpWriteConn struct {
	remote    net.Addr
	stream    *SampleOutputStream
	closeOnce sync.Once
}

func (sink *TcpMetricSink) OpenWriteConn(conn *net.TCPConn) *TcpWriteConn {
	return &TcpWriteConn{
		remote: conn.RemoteAddr(),
		stream: sink.Writer.OpenBuffered(conn, sink.Marshaller),
	}
}

func (conn *TcpWriteConn) Header(header Header) {
	log.Println("Serving", len(header.Fields), "metrics to", conn.remote)
	if err := conn.stream.Header(header); err != nil {
		conn.doClose(err)
	}
}

func (conn *TcpWriteConn) Sample(sample Sample) {
	if err := conn.stream.Sample(sample); err != nil {
		conn.doClose(err)
	}
}

func (conn *TcpWriteConn) Close() {
	if conn != nil {
		conn.doClose(nil)
	}
}

func (conn *TcpWriteConn) doClose(cause error) {
	conn.closeOnce.Do(func() {
		conn.printErr(cause)
		if cause == nil {
			log.Println("Closing connection to", conn.remote)
		}
		if closeErr := conn.stream.Close(); closeErr != nil && cause == nil {
			log.Printf("Error closing connection to %v: %v\n", conn.remote, closeErr)
		}
		conn.stream = nil // Make IsRunning() return false
	})
}

func (conn *TcpWriteConn) IsRunning() bool {
	return conn != nil && conn.stream != nil
}

func (conn *TcpWriteConn) printErr(err error) {
	if operr, ok := err.(*net.OpError); ok {
		if operr.Err == syscall.EPIPE {
			log.Println("Connection closed by", conn.remote)
			return
		} else {
			if syscallerr, ok := operr.Err.(*os.SyscallError); ok && syscallerr.Err == syscall.EPIPE {
				log.Println("Connection closed by", conn.remote)
				return
			}
		}
	}
	if err != nil {
		log.Printf("TCP write to %v failed, closing connection. %v\n", conn.remote, err)
	}
}

// ==================== TCP active sink ====================
type TCPSink struct {
	TcpMetricSink
	Endpoint string
	conn     *TcpWriteConn
	stopped  *golib.OneshotCondition
}

func (sink *TCPSink) String() string {
	return "TCP sink to " + sink.Endpoint
}

func (sink *TCPSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.Println("Sending", sink.Marshaller, "samples to", sink.Endpoint)
	sink.stopped = golib.NewOneshotCondition()
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

func (sink *TCPSink) Header(header Header) error {
	conn, err := sink.getOutputConnection(true)
	if err != nil {
		return err
	}
	conn.Header(header)
	return sink.checkConnRunning(conn)
}

func (sink *TCPSink) Sample(sample Sample, header Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	conn, err := sink.getOutputConnection(false)
	if err != nil {
		return err
	}
	conn.Sample(sample)
	return sink.checkConnRunning(conn)
}

func (sink *TCPSink) checkConnRunning(conn *TcpWriteConn) error {
	if !conn.IsRunning() {
		return fmt.Errorf("Connection to %v closed", sink.Endpoint)
	}
	return nil
}

func (sink *TCPSink) getOutputConnection(renewConnection bool) (conn *TcpWriteConn, err error) {
	closeSink := false
	sink.stopped.IfElseEnabled(func() {
		err = fmt.Errorf("TCP sink to %v already closed", sink.Endpoint)
	}, func() {
		if !sink.conn.IsRunning() || renewConnection {
			// Cleanup errored connection or stop existing connection to negotiate new header
			if sink.conn != nil {
				if !sink.CountConnectionClosed() {
					closeSink = true
				}
			}
			sink.closeConnection()
		}
		if err = sink.assertConnection(); err != nil {
			return
		}
		conn = sink.conn
	})
	if closeSink {
		sink.Close()
	}
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
		sink.conn = sink.OpenWriteConn(conn)
	}
	return nil
}

// ==================== TCP active source ====================
type TCPSource struct {
	AbstractUnmarshallingMetricSource
	TCPConnCounter
	RemoteAddr    string
	RetryInterval time.Duration
	Reader        SampleReader
	loopTask      *golib.LoopTask
	stream        *SampleInputStream
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
			source.handleConnection(conn)
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

func (source *TCPSource) handleConnection(conn *net.TCPConn) {
	source.loopTask.IfNotEnabled(func() {
		source.stream = source.Reader.Open(conn, source.Unmarshaller, source.OutgoingSink)
	})
	if !source.loopTask.Enabled() {
		source.stream.ReadTcpSamples(conn, source.isConnectionClosed)
		if !source.CountConnectionClosed() {
			source.Stop()
		}
	}
}

func (source *TCPSource) Stop() {
	source.loopTask.Enable(func() {
		_ = source.stream.Close() // Ignore error
	})
}

func (source *TCPSource) isConnectionClosed() bool {
	return source.loopTask.Enabled()
}

func (source *TCPSource) dial() (*net.TCPConn, error) {
	endpoint, err := net.ResolveTCPAddr("tcp", source.RemoteAddr)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, endpoint)
}
