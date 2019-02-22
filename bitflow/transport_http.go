package bitflow

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/antongulenko/golib"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// HttpServerSink implements the SampleSink interface as an HTTP server.
// It listens for incoming HTTP connections on a port and provides incoming data on certain HTTP request paths.
type HttpServerSink struct {
	AbstractTcpSink

	// Endpoint defines the TCP host and port to listen on for incoming TCP connections.
	// The host can be empty (e.g. ":1234"). If not, it must contain a hostname or IP of the
	// local host.
	Endpoint string

	// If BufferedSamples is >0, the given number of latest samples will be kept in a ring buffer.
	// New requests will first receive all samples currently in the buffer, and will
	// afterwards continue receiving live incoming samples.
	BufferedSamples uint

	// SubPathTag can be set to allow requesting samples on HTTP path /<val>, which will only output that
	// contain the tag <SubPathTag>=<val>. The root path '/' still serves all samples.
	SubPathTag string

	// RootPathPrefix is the base path for requests. A '/' will be appended.
	RootPathPrefix string

	buf outputSampleBuffer
	gin *golib.GinTask
	wg  *sync.WaitGroup
}

// String implements the SampleSink interface.
func (sink *HttpServerSink) String() string {
	msg := "HTTP sink on " + sink.Endpoint
	if sink.SubPathTag != "" {
		msg += " (paths defined by tag '" + sink.SubPathTag + "')"
	}
	return msg
}

// Start implements the SampleSink interface. It creates the TCP socket and
// starts listening on it in a separate goroutine. Any incoming connection is
// then handled in their own goroutine.
func (sink *HttpServerSink) Start(wg *sync.WaitGroup) golib.StopChan {
	sink.connCounterDescription = sink
	sink.Protocol = "HTTP"
	sink.wg = wg
	capacity := sink.BufferedSamples
	if capacity == 0 {
		capacity = 1
	}
	sink.buf = outputSampleBuffer{
		Capacity: capacity,
		cond:     sync.NewCond(new(sync.Mutex)),
	}
	sink.gin = golib.NewGinTask(sink.Endpoint)
	sink.gin.ShutdownHook = func() {
		sink.buf.closeBuffer()
		sink.CloseSink()
	}
	log.WithField("format", sink.Marshaller).Println("Listening for output HTTP requests on", sink.Endpoint)
	sink.gin.GET(sink.RootPathPrefix+"/", sink.handleRequest)
	if sink.SubPathTag != "" {
		sink.gin.GET(sink.RootPathPrefix+"/:tagVal", sink.handleRequest)
	}
	return sink.gin.Start(wg)
}

// Close implements the SampleSink interface. It closes any existing connection
// and shuts down the HTTP server.
func (sink *HttpServerSink) Close() {
	sink.gin.Stop()
}

// Sample implements the SampleSink interface. It stores the sample in a ring buffer
// and sends it to all established connections. New connections will first receive
// all samples stored in the buffer, before getting the live samples directly.
// If the buffer is disable or full, and there are no established connections,
// samples are dropped.
func (sink *HttpServerSink) Sample(sample *Sample, header *Header) error {
	sink.buf.add(sample, header)
	return sink.AbstractSampleOutput.Sample(nil, sample, header)
}

func (sink *HttpServerSink) handleRequest(ctx *gin.Context) {
	if !sink.countConnectionAccepted(ctx.Request.RemoteAddr) {
		ctx.Status(http.StatusGone)
		msg := fmt.Sprintf("%sRejecting connection, already accepted %v connections", sink.msg(), sink.accepted)
		_, _ = ctx.Writer.WriteString(msg) // Error dropped
		return
	}

	filterTagValue := ""
	if sink.SubPathTag != "" {
		filterTagValue = ctx.Param("tagVal")
	}

	ctx.Header("Connection", "Keep-Alive")
	ctx.Header("Transfer-Encoding", "chunked")
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Writer.Flush()

	writeConn := sink.OpenWriteConn(sink.wg, ctx.Request.RemoteAddr, httpResponseWriteCloser{ctx.Writer})
	sink.wg.Add(1)

	if sink.Marshaller.ShouldCloseAfterFirstSample() {
		sink.buf.sendOneSample(writeConn, ctx.Writer.Flush)
		writeConn.Close()
	} else {
		sink.sendSamples(sink.wg, writeConn, ctx.Writer, filterTagValue)
	}

}

func (sink *HttpServerSink) sendSamples(wg *sync.WaitGroup, conn *TcpWriteConn, flusher http.Flusher, filterTagValue string) {
	defer wg.Done()
	defer func() {
		conn.Close()
		if !sink.countConnectionClosed() {
			sink.Close()
		}
	}()
	if filterTagValue != "" {
		log.Printf("Serving samples over HTTP, containing tag %v=%v", sink.SubPathTag, filterTagValue)
	}
	sink.buf.sendFilteredSamples(conn, flusher.Flush,
		func(sample *Sample, header *Header) bool {
			if sink.SubPathTag != "" && filterTagValue != "" {
				tagVal := sample.Tag(sink.SubPathTag)
				return filterTagValue == tagVal
			}
			return true
		})
}

type httpResponseWriteCloser struct {
	io.Writer
}

func (httpResponseWriteCloser) Close() error {
	// Do nothing. The HTTP response will be closed automatically.
	return nil
}

func dialHTTP(endpoint string, timeout time.Duration) (io.ReadCloser, string, error) {
	if !strings.HasPrefix(endpoint, "http://") {
		endpoint = "http://" + endpoint
	}
	// There is no way to set a timeout for establishing the connection and exchanging HTTP headers,
	// so set the same timeout for all involved steps (except for receiving the body, since it is streamed/chunked)
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
		},
	}

	resp, err := client.Get(endpoint)
	if err == nil && resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		var bodyStr string
		if err != nil {
			bodyStr = fmt.Sprintf("Failed to get response body: %v", err)
		} else {
			bodyStr = "Body: " + string(body)
		}
		err = fmt.Errorf("Reponse return status code %v. %v", resp.StatusCode, bodyStr)
	}
	if err != nil {
		return nil, "", err
	}
	return resp.Body, endpoint, nil
}
