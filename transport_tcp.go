package bitflow

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

// TCPConnCounter contains the TcpConnLimit configuration parameter that optionally
// defines a limit for the number of TCP connection that are accepted or initiated by the
// MetricSink and MetricSource implementations using TCP connections.
type TCPConnCounter struct {
	// TcpConnLimit defines a limit for the number of TCP connections that should be accepted
	// or initiated. When this is <= 0, the number of not limited.
	TcpConnLimit uint

	connCounterDescription interface{}
	closed                 uint
	accepted               uint
}

func (counter *TCPConnCounter) msg() string {
	if counter.connCounterDescription != nil {
		return fmt.Sprintf("%v: ", counter.connCounterDescription)
	}
	return ""
}

// Return true, if we can handle further connections. Return false, if the application should stop.
func (counter *TCPConnCounter) countConnectionClosed() bool {
	if counter.TcpConnLimit > 0 {
		counter.closed++
		if counter.closed >= counter.TcpConnLimit {
			log.Warnln(counter.msg()+"Handled", counter.closed, "TCP connection(s)")
			return false
		}
	}
	return true
}

// Return true, if the connection was accepted. Return false, if it was rejected and closed.
func (counter *TCPConnCounter) countConnectionAccepted(conn *net.TCPConn) bool {
	if counter.TcpConnLimit > 0 {
		if counter.accepted >= counter.TcpConnLimit {
			log.WithField("remote", conn.RemoteAddr()).Warnln(counter.msg()+"Rejecting connection, already accepted", counter.accepted, "connections")
			_ = conn.Close() // Drop error
			return false
		}
		counter.accepted++
	}
	return true
}

// AbstractTcpSink is a helper type for TCP-based MetricSink implementations.
// The two fields AbstractMarshallingMetricSink and TCPConnCounter can be used
// to configure different aspects of the marshalling and writing of the data.
// The purpose of AbstractTcpSink is to create instances of TcpWriteConn with the
// configured parameters.
type AbstractTcpSink struct {
	AbstractMarshallingMetricSink
	TCPConnCounter
}

// TcpWriteConn is a helper type for TCP-base MetricSink implementations.
// It can send Headers and Samples over an opened TCP connection.
// It is created from AbstractTcpSink.OpenWriteConn() and can be used until
// Sample() returns an error or Close() is called explicitely.
type TcpWriteConn struct {
	checker   HeaderChecker
	stream    *SampleOutputStream
	closeOnce sync.Once
	log       *log.Entry
}

// OpenWriteConn wraps a net.TCPConn in a new TcpWriteConn using the parameters defined in
// the receiving AbstractTcpSink.
func (sink *AbstractTcpSink) OpenWriteConn(conn *net.TCPConn) *TcpWriteConn {
	return &TcpWriteConn{
		stream: sink.Writer.Open(conn, sink.Marshaller),
		log:    log.WithField("remote", conn.RemoteAddr()),
	}
}

// Sample writes the given sample into the receiving TcpWriteConn and closes
// the underlying TCP connection if there is an error.
func (conn *TcpWriteConn) Sample(sample *Sample, header *Header) {
	if conn.checker.HeaderChanged(header) {
		conn.log.Println("Serving", len(header.Fields), "metrics")
	}
	if err := conn.stream.Sample(sample, header); err != nil {
		conn.doClose(err)
	}
}

// Close explicitely closes the underlying TCP connection of the receiving TcpWriteConn.
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

// IsRunning returns true, if the receiving TcpWriteConn is connected to a remote TCP endpoint.
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

// TCPSink implements MetricSink by sending the received Headers and Samples
// to a given remote TCP endpoint. Everytime it receives a Header or a Sample,
// it checks whether a TCP connection is already established. If so, it sends
// the data on the existing connection. Otherwise, it tries to connect to the
// configured endpoint and sends the data there, if the connection is successful.
type TCPSink struct {
	// AbstractTcpSink contains different configuration options regarding the
	// marshalling and writing of data to the remote TCP connection.
	AbstractTcpSink

	// Endpoint is the target TCP endpoint to connect to for sending marshalled data.
	Endpoint string

	// PrintErrors controls whether TCP related errors are dropped, or treated normally.
	// This should usually be set to true. Setting it to false results in no visible
	// errors, even if sending data fails. It can be useful if failing TCP connections
	// are expected, or if the errors are already printed otherwise.
	PrintErrors bool

	conn    *TcpWriteConn
	stopped *golib.OneshotCondition
}

// String implements the MetricSink interface.
func (sink *TCPSink) String() string {
	return "TCP sink to " + sink.Endpoint
}

// Start implements the MetricSink interface. It creates a log message
// and prepares the TCPSink for sending data.
func (sink *TCPSink) Start(wg *sync.WaitGroup) golib.StopChan {
	sink.connCounterDescription = sink
	log.WithField("format", sink.Marshaller).Println("Sending data to", sink.Endpoint)
	sink.stopped = golib.NewOneshotCondition()
	return sink.stopped.Start(wg)
}

func (sink *TCPSink) closeConnection() {
	sink.conn.Close()
	sink.conn = nil
}

// Close implements the MetricSink interface. It stops the current TCP connection,
// if one is running, and prevents future connections from being created. No more
// data can be sent into the TCPSink after this.
func (sink *TCPSink) Close() {
	sink.stopped.Enable(func() {
		sink.closeConnection()
	})
}

// Sample implements the MetricSink interface. If a connection is already established,
// the Sample is directly sent throgh it. Otherwise, a new connection is established,
// using the last Header that was was sent into the receiving TCPSink. An error is returned,
// if no Header() call has been done before this call to Sample().
func (sink *TCPSink) Sample(sample *Sample, header *Header) error {
	if err := sample.Check(header); err != nil {
		return err
	}
	conn, err := sink.getOutputConnection()
	if err != nil {
		if !sink.PrintErrors {
			err = nil
		}
		return err
	}
	conn.Sample(sample, header)
	return sink.checkConnRunning(conn)
}

func (sink *TCPSink) checkConnRunning(conn *TcpWriteConn) error {
	if !conn.IsRunning() {
		return fmt.Errorf("Connection to %v closed", sink.Endpoint)
	}
	return nil
}

func (sink *TCPSink) getOutputConnection() (conn *TcpWriteConn, err error) {
	closeSink := false
	sink.stopped.IfElseEnabled(func() {
		err = fmt.Errorf("TCP sink to %v already closed", sink.Endpoint)
	}, func() {
		if !sink.conn.IsRunning() {
			// Cleanup errored connection or stop existing connection to negotiate new header
			if sink.conn != nil {
				if !sink.countConnectionClosed() {
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

// TCPSource implements the MetricSource interface by connecting to a list of remote TCP
// endpoints and downloading Header and Sample data from there. A background goroutine continuously
// tries to establish the required TCP connections and reads data from it whenever a connection
// succeeds. The contained AbstractMetricSource and TCPConnCounter fields provide various parmaeters
// for configuring different aspects of the TCP connections and reading of data from them.
type TCPSource struct {
	AbstractMetricSource
	TCPConnCounter

	// PrintErrors controls whether TCP related errors are dropped, or treated normally.
	// This should usually be set to true. Setting it to false results in no visible
	// errors, even if sending data fails. It can be useful if failing TCP connections
	// are expected, or if the errors are already printed otherwise.
	PrintErrors bool

	// RemoteAddrs defines the list of remote TCP endpoints that the TCPSource will try to
	// connect to. If there are more than one connection, all connections will run in parallel.
	// In that case, an additional instance of SynchronizedMetricSink is used to synchronize all
	// received data. For multiple connections, all samples and headers will be pushed into the
	// outgoing MetricSink in an interleaved fashion, so the outgoing MetricSink must be able to handle that.
	RemoteAddrs []string

	// RetryInterval defines the time to wait before trying to reconnect after a closed connection
	// or failed connection attempt.
	RetryInterval time.Duration

	// Reader defines various parameters configuring reading and parsing aspects of TCPSource.
	// See the SampleReader doc for more information.
	Reader SampleReader

	downloaders  []*tcpDownloadTask
	downloadSink MetricSinkBase
}

// String implements the MetricSource interface.
func (sink *TCPSource) String() string {
	return "TCP download (" + sink.SourceString() + ")"
}

// SourceString returns a string representation of the TCP endpoints the TCPSource
// will download data from.
func (sink *TCPSource) SourceString() string {
	if len(sink.RemoteAddrs) == 1 {
		return sink.RemoteAddrs[0]
	} else {
		return strconv.Itoa(len(sink.RemoteAddrs)) + " sources"
	}
}

// Start implements the MetricSource interface. It starts one goroutine for every
// configured TCP endpoint. The goroutines continuously try to connect to the remote
// endpoints and download Headers and Samples as soon as a connection is established.
func (source *TCPSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.connCounterDescription = source
	log.WithField("format", source.Reader.Format()).Println("Downloading from", source.SourceString())
	channels := make([]golib.StopChan, 0, len(source.RemoteAddrs))
	if len(source.RemoteAddrs) > 1 {
		source.downloadSink = &SynchronizingMetricSink{OutgoingSink: source.OutgoingSink}
	} else {
		source.downloadSink = source.OutgoingSink
	}
	for _, remote := range source.RemoteAddrs {
		task := &tcpDownloadTask{
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

// Stop implements the MetricSource interface. It stops all background goroutines and tries
// to gracefully close all established TCP connections.
func (source *TCPSource) Stop() {
	for _, downloader := range source.downloaders {
		downloader.Stop()
	}
}

func (source *TCPSource) startStream(conn *net.TCPConn) *SampleInputStream {
	return source.Reader.Open(conn, source.downloadSink)
}

// ====================== Internal types ======================

type tcpDownloadTask struct {
	source   *TCPSource
	remote   string
	loopTask *golib.LoopTask
	stream   *SampleInputStream
}

func (task *tcpDownloadTask) Start(wg *sync.WaitGroup) golib.StopChan {
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

func (task *tcpDownloadTask) handleConnection(conn *net.TCPConn) {
	task.loopTask.IfNotEnabled(func() {
		task.stream = task.source.startStream(conn)
	})
	if !task.loopTask.Enabled() {
		task.stream.ReadTcpSamples(conn, task.isConnectionClosed)
		if !task.source.countConnectionClosed() {
			task.source.Stop()
		}
	}
}

func (task *tcpDownloadTask) Stop() {
	task.loopTask.Enable(func() {
		_ = task.stream.Close() // Ignore error
	})
}

func (task *tcpDownloadTask) isConnectionClosed() bool {
	return task.loopTask.Enabled()
}

func (task *tcpDownloadTask) dial() (*net.TCPConn, error) {
	endpoint, err := net.ResolveTCPAddr("tcp", task.remote)
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, endpoint)
}
