package sample

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

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
			log.Warnln("Handled", counter.closed, "TCP connection(s)")
			return false
		}
	}
	return true
}

// Return true, if the connection was accepted. Return false, if it was rejected and closed.
func (counter *TCPConnCounter) CountConnectionAccepted(conn *net.TCPConn) bool {
	if counter.TcpConnLimit > 0 {
		if counter.accepted >= counter.TcpConnLimit {
			log.WithField("remote", conn.RemoteAddr()).Warnln("Rejecting connection, already accepted", counter.accepted, "connections")
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
	stream    *SampleOutputStream
	closeOnce sync.Once
	log       *log.Entry
}

func (sink *TcpMetricSink) OpenWriteConn(conn *net.TCPConn) *TcpWriteConn {
	return &TcpWriteConn{
		stream: sink.Writer.Open(conn, sink.Marshaller),
		log:    log.WithField("remote", conn.RemoteAddr()),
	}
}

func (conn *TcpWriteConn) Header(header Header) {
	conn.log.Println("Serving", len(header.Fields), "metrics")
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
			conn.log.Debugln("Closing connection")
		}
		if closeErr := conn.stream.Close(); closeErr != nil && cause == nil {
			conn.log.Errorln("Error closing connection:", closeErr)
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
			conn.log.Debugln("Connection closed by remote")
			return
		} else {
			if syscallerr, ok := operr.Err.(*os.SyscallError); ok && syscallerr.Err == syscall.EPIPE {
				conn.log.Debugln("Connection closed by remote")
				return
			}
		}
	}
	if err != nil {
		conn.log.Errorln("TCP write failed, closing connection:", err)
	}
}

// ==================== TCP active sink ====================
type TCPSink struct {
	TcpMetricSink
	Endpoint    string
	PrintErrors bool
	conn        *TcpWriteConn
	stopped     *golib.OneshotCondition
}

func (sink *TCPSink) String() string {
	return "TCP sink to " + sink.Endpoint
}

func (sink *TCPSink) Start(wg *sync.WaitGroup) golib.StopChan {
	log.WithField("format", sink.Marshaller).Println("Sending data to", sink.Endpoint)
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
		if !sink.PrintErrors {
			err = nil
		}
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
		if !sink.PrintErrors {
			err = nil
		}
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
	AbstractMetricSource
	TCPConnCounter
	PrintErrors   bool
	RemoteAddrs   []string
	RetryInterval time.Duration
	Reader        SampleReader

	downloaders  []*TCPDownloadTask
	downloadSink MetricSinkBase
}

func (sink *TCPSource) String() string {
	return "TCP download (" + sink.SourceString() + ")"
}

func (sink *TCPSource) SourceString() string {
	if len(sink.RemoteAddrs) == 1 {
		return sink.RemoteAddrs[0]
	} else {
		return strconv.Itoa(len(sink.RemoteAddrs)) + " sources"
	}
}

func (source *TCPSource) Start(wg *sync.WaitGroup) golib.StopChan {
	log.WithField("format", source.Reader.Format()).Println("Downloading from", source.SourceString())
	channels := make([]golib.StopChan, 0, len(source.RemoteAddrs))
	if len(source.RemoteAddrs) > 1 {
		source.downloadSink = &SynchronizingMetricSink{OutgoingSink: source.OutgoingSink}
	} else {
		source.downloadSink = source.OutgoingSink
	}
	for _, remote := range source.RemoteAddrs {
		task := &TCPDownloadTask{
			source: source,
			remote: remote,
		}
		source.downloaders = append(source.downloaders, task)
		channels = append(channels, task.Start(wg))
	}
	return golib.WaitErrFunc(wg, func() error {
		defer source.CloseSink(wg)
		var errors golib.MultiError
		_, err := golib.WaitForAny(channels)
		errors.Add(err)
		source.Stop()
		for _, c := range channels {
			if c != nil {
				errors.Add(<-c)
			}
		}
		return errors.NilOrError()
	})
}

func (source *TCPSource) Stop() {
	for _, downloader := range source.downloaders {
		downloader.Stop()
	}
}

func (source *TCPSource) startStream(conn *net.TCPConn) *SampleInputStream {
	return source.Reader.Open(conn, source.downloadSink)
}

// ==================== Loop connecting to TCP endpoint ====================

type TCPDownloadTask struct {
	source   *TCPSource
	remote   string
	loopTask *golib.LoopTask
	stream   *SampleInputStream
}

func (task *TCPDownloadTask) Start(wg *sync.WaitGroup) golib.StopChan {
	task.loopTask = golib.NewLoopTask("tcp download loop", func(stop golib.StopChan) {
		if conn, err := task.dial(); err != nil {
			if task.source.PrintErrors {
				log.WithField("remote", task.remote).Errorln("Error downloading data:", err)
			}
		} else {
			task.handleConnection(conn)
		}
		select {
		case <-time.After(task.source.RetryInterval):
		case <-stop:
		}
	})
	return task.loopTask.Start(wg)
}

func (task *TCPDownloadTask) handleConnection(conn *net.TCPConn) {
	task.loopTask.IfNotEnabled(func() {
		task.stream = task.source.startStream(conn)
	})
	if !task.loopTask.Enabled() {
		task.stream.ReadTcpSamples(conn, task.isConnectionClosed)
		if !task.source.CountConnectionClosed() {
			task.source.Stop()
		}
	}
}

func (task *TCPDownloadTask) Stop() {
	task.loopTask.Enable(func() {
		_ = task.stream.Close() // Ignore error
	})
}

func (task *TCPDownloadTask) isConnectionClosed() bool {
	return task.loopTask.Enabled()
}

func (task *TCPDownloadTask) dial() (*net.TCPConn, error) {
	endpoint, err := net.ResolveTCPAddr("tcp", task.remote)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, endpoint)
}
