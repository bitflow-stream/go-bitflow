package bitflow

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
)

// TCPConnCounter contains the TcpConnLimit configuration parameter that optionally
// defines a limit for the number of TCP connection that are accepted or initiated by the
// SampleSink and SampleSource implementations using TCP connections.
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
func (counter *TCPConnCounter) countConnectionAccepted(remoteAddr string) bool {
	if counter.TcpConnLimit > 0 {
		if counter.accepted >= counter.TcpConnLimit {
			log.WithField("remote", remoteAddr).Warnln(counter.msg()+"Rejecting connection, already accepted", counter.accepted, "connections")
			return false
		}
		counter.accepted++
	}
	return true
}

// AbstractTcpSink is a helper type for TCP-based SampleSink implementations.
// The two fields AbstractSampleOutput and TCPConnCounter can be used
// to configure different aspects of the marshalling and writing of the data.
// The purpose of AbstractTcpSink is to create instances of TcpWriteConn with the
// configured parameters.
type AbstractTcpSink struct {
	AbstractMarshallingSampleOutput
	TCPConnCounter

	// LogReceivedTraffic enables logging received TCP traffic, which is usually not expected.
	// Only the values log.ErrorLevel, log.WarnLevel, log.InfoLevel, log.DebugLevel enable logging.
	LogReceivedTraffic log.Level

	// Protocol is used for more detailed logging
	Protocol string
}

// TcpWriteConn is a helper type for TCP-base SampleSink implementations.
// It can send Headers and Samples over an opened TCP connection.
// It is created from AbstractTcpSink.OpenWriteConn() and can be used until
// Sample() returns an error or Close() is called explicitly.
type TcpWriteConn struct {
	checker   HeaderChecker
	stream    *SampleOutputStream
	closeOnce sync.Once
	log       *log.Entry
	proto     string
}

// OpenWriteConn wraps a net.TCPConn in a new TcpWriteConn using the parameters defined in
// the receiving AbstractTcpSink.
func (sink *AbstractTcpSink) OpenWriteConn(wg *sync.WaitGroup, remoteAddr string, conn io.WriteCloser) *TcpWriteConn {
	res := &TcpWriteConn{
		stream: sink.Writer.Open(conn, sink.Marshaller),
		log:    log.WithField("remote", remoteAddr).WithField("protocol", sink.Protocol).WithField("format", sink.Marshaller),
		proto:  sink.Protocol,
	}
	switch sink.LogReceivedTraffic {
	case log.ErrorLevel, log.WarnLevel, log.InfoLevel, log.DebugLevel:
		if readWriteCloser, ok := conn.(io.ReadWriteCloser); ok {
			wg.Add(1)
			go res.logReceivedTraffic(wg, readWriteCloser, sink.LogReceivedTraffic)
		}
	}
	return res
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

// Close explicitly closes the underlying TCP connection of the receiving TcpWriteConn.
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

func (conn *TcpWriteConn) logReceivedTraffic(wg *sync.WaitGroup, tcpConn io.Reader, level log.Level) {
	defer wg.Done()

	// TODO make the receive buffer configurable
	buf := bufio.NewReader(tcpConn)
	for {
		lineBytes, _, err := buf.ReadLine()
		line := string(lineBytes)
		if err != nil {
			// This most likely means the connection was closed
			conn.log.Debugln("Error receiving data (probably because of closed connection):", err)
			if !conn.IsRunning() {
				return
			}
		} else {
			msg := "Received data on output connection:"
			switch level {
			case log.ErrorLevel:
				conn.log.Errorln(msg, line)
			case log.WarnLevel:
				conn.log.Warnln(msg, line)
			case log.InfoLevel:
				conn.log.Infoln(msg, line)
			case log.DebugLevel:
				conn.log.Debugln(msg, line)
			}
		}
	}
}

// IsRunning returns true, if the receiving TcpWriteConn is connected to a remote TCP endpoint.
func (conn *TcpWriteConn) IsRunning() bool {
	return conn != nil && conn.stream != nil
}

func (conn *TcpWriteConn) printErr(err error) {
	if opErr, ok := err.(*net.OpError); ok && IsBrokenPipeError(opErr.Err) {
		conn.log.Debugln("Connection closed by remote")
		return
	}
	if err != nil {
		conn.log.Errorln("Write failed, closing connection:", err)
	}
}

// TCPSink implements SampleSink by sending the received Headers and Samples
// to a given remote TCP endpoint. Every time it receives a Header or a Sample,
// it checks whether a TCP connection is already established. If so, it sends
// the data on the existing connection. Otherwise, it tries to connect to the
// configured endpoint and sends the data there, if the connection is successful.
type TCPSink struct {
	// AbstractTcpSink contains different configuration options regarding the
	// marshalling and writing of data to the remote TCP connection.
	AbstractTcpSink

	// Endpoint is the target TCP endpoint to connect to for sending marshalled data.
	Endpoint string

	// DialTimeout can be set to time out automatically when connecting to a remote TCP endpoint
	DialTimeout time.Duration

	conn    *TcpWriteConn
	stopped golib.StopChan
	wg      *sync.WaitGroup
}

// String implements the SampleSink interface.
func (sink *TCPSink) String() string {
	return "TCP sink to " + sink.Endpoint
}

// Start implements the SampleSink interface. It creates a log message
// and prepares the TCPSink for sending data.
func (sink *TCPSink) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	sink.connCounterDescription = sink
	sink.Protocol = "TCP"
	log.WithField("format", sink.Marshaller).Println("Sending data to", sink.Endpoint)
	sink.stopped = golib.NewStopChan()
	sink.wg = wg
	return
}

func (sink *TCPSink) closeConnection() {
	sink.conn.Close()
	sink.conn = nil
}

// Close implements the SampleSink interface. It stops the current TCP connection,
// if one is running, and prevents future connections from being created. No more
// data can be sent into the TCPSink after this.
func (sink *TCPSink) Close() {
	sink.stopped.StopFunc(func() {
		sink.closeConnection()
		sink.CloseSink()
	})
}

// Sample implements the SampleSink interface. If a connection is already established,
// the Sample is directly sent through it. Otherwise, a new connection is established,
// and the sample is sent there.
func (sink *TCPSink) Sample(sample *Sample, header *Header) error {
	conn, err := sink.getOutputConnection()
	if err == nil {
		conn.Sample(sample, header)
		err = sink.checkConnRunning(conn)
	}
	return sink.AbstractSampleOutput.Sample(err, sample, header)
}

func (sink *TCPSink) checkConnRunning(conn *TcpWriteConn) error {
	if !conn.IsRunning() {
		return fmt.Errorf("Connection to %v closed", sink.Endpoint)
	}
	return nil
}

func (sink *TCPSink) getOutputConnection() (conn *TcpWriteConn, err error) {
	closeSink := false
	sink.stopped.IfElseStopped(func() {
		err = fmt.Errorf("TCP sink to %v already closed", sink.Endpoint)
	}, func() {
		if !sink.conn.IsRunning() {
			// Cleanup failed connection or stop existing connection to negotiate new header
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
		conn, _, err := dialTcp(sink.Endpoint, sink.DialTimeout)
		if err != nil {
			return err
		}
		sink.conn = sink.OpenWriteConn(sink.wg, conn.RemoteAddr().String(), conn)
	}
	return nil
}

// TCPSource implements the SampleSource interface by connecting to a list of remote TCP
// endpoints and downloading Header and Sample data from there. A background goroutine continuously
// tries to establish the required TCP connections and reads data from it whenever a connection
// succeeds. The contained AbstractUnmarshallingSampleSource and TCPConnCounter fields provide various parameters
// for configuring different aspects of the TCP connections and reading of data from them.
type TCPSource struct {
	AbstractUnmarshallingSampleSource
	TCPConnCounter

	// RemoteAddrs defines the list of remote TCP endpoints that the TCPSource will try to
	// connect to. If there are more than one connection, all connections will run in parallel.
	// In that case, an additional instance of SynchronizedSampleSink is used to synchronize all
	// received data. For multiple connections, all samples and headers will be pushed into the
	// outgoing SampleSink in an interleaved fashion, so the outgoing SampleSink must be able to handle that.
	RemoteAddrs []string

	// PrintErrors controls whether errors from establishing download TCP connections are logged or not.
	PrintErrors bool

	// RetryInterval defines the time to wait before trying to reconnect after a closed connection
	// or failed connection attempt.
	RetryInterval time.Duration

	// DialTimeout can be set to time out automatically when connecting to a remote TCP endpoint
	DialTimeout time.Duration

	// UseHTTP instructs this data source to use the HTTP protocol instead of TCP. In this case, the RemoteAddrs
	// strings are treated as HTTP URLs, but without the http:// prefix. This prefix is appended before attempting to
	// send an HTTP request.
	UseHTTP bool

	downloadTasks []*tcpDownloadTask
	downloadSink  SampleSink
}

// String implements the SampleSource interface.
func (source *TCPSource) String() string {
	proto := "TCP"
	if source.UseHTTP {
		proto = "HTTP"
	}
	return proto + " download (" + source.SourceString() + ")"
}

// SourceString returns a string representation of the TCP endpoints the TCPSource
// will download data from.
func (source *TCPSource) SourceString() string {
	if len(source.RemoteAddrs) == 1 {
		return source.RemoteAddrs[0]
	} else {
		return strconv.Itoa(len(source.RemoteAddrs)) + " sources"
	}
}

// Start implements the SampleSource interface. It starts one goroutine for every
// configured TCP endpoint. The goroutines continuously try to connect to the remote
// endpoints and download Headers and Samples as soon as a connection is established.
func (source *TCPSource) Start(wg *sync.WaitGroup) golib.StopChan {
	source.connCounterDescription = source
	log.WithField("format", source.Reader.Format()).Println("Downloading from", source.SourceString())
	if len(source.RemoteAddrs) > 1 {
		source.downloadSink = &SynchronizingSampleSink{Out: source.GetSink()}
	} else {
		source.downloadSink = source.GetSink()
	}
	tasks := make(golib.TaskGroup, 0, len(source.RemoteAddrs))
	for _, remote := range source.RemoteAddrs {
		task := &tcpDownloadTask{
			source: source,
			remote: remote,
		}
		source.downloadTasks = append(source.downloadTasks, task)
		tasks.Add(task)
	}
	channels := tasks.StartTasks(wg)
	return golib.WaitErrFunc(wg, func() error {
		defer source.CloseSinkParallel(wg)
		golib.WaitForAny(channels)
		source.Close()
		return tasks.CollectMultiError(channels).NilOrError()
	})
}

// Close implements the SampleSource interface. It stops all background goroutines and tries
// to gracefully close all established TCP connections.
func (source *TCPSource) Close() {
	for _, downloader := range source.downloadTasks {
		downloader.Stop()
	}
}

func (source *TCPSource) startStream(conn io.ReadCloser) *SampleInputStream {
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
	task.loopTask = &golib.LoopTask{
		Description: "tcp download loop",
		Loop: func(stop golib.StopChan) error {
			if conn, remote, err := task.dial(); err != nil {
				if task.source.PrintErrors {
					log.WithField("remote", task.remote).Errorln("Error downloading data:", err)
				}
			} else {
				task.handleConnection(conn, remote)
			}
			stop.WaitTimeout(task.source.RetryInterval)
			return nil
		},
	}
	return task.loopTask.Start(wg)
}

func (task *tcpDownloadTask) handleConnection(conn io.ReadCloser, remote string) {
	task.loopTask.IfNotStopped(func() {
		task.stream = task.source.startStream(conn)
	})
	if !task.loopTask.Stopped() {
		task.stream.ReadTcpSamples(conn, remote, task.isConnectionClosed)
		if !task.source.countConnectionClosed() {
			task.source.Close()
		}
	}
}

func (task *tcpDownloadTask) Stop() {
	task.loopTask.StopFunc(func() {
		_ = task.stream.Close() // Ignore error
	})
}

func (task *tcpDownloadTask) String() string {
	return fmt.Sprintf("TCP downloader (%v)", task.remote)
}

func (task *tcpDownloadTask) isConnectionClosed() bool {
	return task.loopTask.Stopped()
}

func (task *tcpDownloadTask) dial() (io.ReadCloser, string, error) {
	if task.source.UseHTTP {
		return dialHTTP(task.remote, task.source.DialTimeout)
	} else {
		return dialTcp(task.remote, task.source.DialTimeout)
	}
}

func dialTcp(endpoint string, timeout time.Duration) (*net.TCPConn, string, error) {
	conn, err := net.DialTimeout("tcp", endpoint, timeout)
	if err != nil {
		return nil, "", err
	}
	if netConn, ok := conn.(*net.TCPConn); !ok {
		return nil, "", fmt.Errorf("net.DialTimeout() returned a %T (%v) instead of *net.TCPConn", conn, conn)
	} else {
		return netConn, netConn.RemoteAddr().String(), nil
	}
}
