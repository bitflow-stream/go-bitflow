package bitflow

import (
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/antongulenko/golib"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

func GetLocalPort() string {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		if closeErr := l.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

type TcpListenerTestSuite struct {
	testSuiteWithSamples
}

func TestTcpListener(t *testing.T) {
	suite.Run(t, new(TcpListenerTestSuite))
}

func (suite *TcpListenerTestSuite) runGroup(sender golib.Task, generator SampleProcessor, receiver SampleSource, sink *testSampleSink) {
	var group golib.TaskGroup
	(&SamplePipeline{
		Processors: []SampleProcessor{generator},
	}).Construct(&group)
	(&SamplePipeline{
		Source:     receiver,
		Processors: []SampleProcessor{sink},
	}).Construct(&group)
	group.Add(&golib.SetupTask{Setup: func() {
		// Before starting the sender, give the pipeline elements a moment to set up. This allows a TCP listener to open its port, before the TCP sender starts sending data.
		// Otherwise, sometimes the port is not yet opened, and a Connection Refused error occurs
		time.Sleep(50 * time.Millisecond)
	}})
	group.Add(sender)
	group.Add(&golib.TimeoutTask{DumpGoroutines: false, Timeout: 1 * time.Second})
	_, numErrs := group.WaitAndStop(1 * time.Second)
	suite.Equal(0, numErrs, "number of errors")
}

func (suite *TcpListenerTestSuite) testListenerSinkAll(m BidiMarshaller) {
	testSink := suite.newFilledTestSink()
	port := GetLocalPort()

	l := &TCPListenerSink{
		Endpoint:        ":" + port,
		BufferedSamples: 100,
	}
	l.Writer.ParallelSampleHandler = parallel_handler
	l.SetMarshaller(m)

	s := &TCPSource{
		PrintErrors:   true,
		RemoteAddrs:   []string{"localhost:" + port},
		RetryInterval: time.Second,
		DialTimeout:   tcpDialTimeout,
	}
	s.Reader.ParallelSampleHandler = parallel_handler

	sender := &oneShotTask{
		do: func() {
			suite.sendAllSamples(l)
		},
	}

	go func() {
		testSink.waitEmpty()
		s.Close()
		l.Close()
	}()

	suite.runGroup(sender, l, s, testSink)
}

func (suite *TcpListenerTestSuite) testListenerSinkIndividual(m Marshaller) {
	for i := range suite.headers {
		testSink := suite.newTestSinkFor(i)
		port := GetLocalPort()

		// TODO test that a smaller buffer leads to dropped samples

		l := &TCPListenerSink{
			Endpoint:        ":" + port,
			BufferedSamples: 100,
		}
		l.Writer.ParallelSampleHandler = parallel_handler
		l.SetMarshaller(m)

		s := &TCPSource{
			PrintErrors:   true,
			RemoteAddrs:   []string{"localhost:" + port},
			RetryInterval: tcpDownloadRetryInterval,
			DialTimeout:   tcpDialTimeout,
		}
		s.Reader.ParallelSampleHandler = parallel_handler

		sender := &oneShotTask{
			do: func() {
				suite.sendSamples(l, i)
			},
		}

		go func() {
			testSink.waitEmpty()
			s.Close()
			l.Close()
		}()

		suite.runGroup(sender, l, s, testSink)
	}
}

func (suite *TcpListenerTestSuite) TestListenerSinkIndividualCsv() {
	suite.testListenerSinkIndividual(new(CsvMarshaller))
}

func (suite *TcpListenerTestSuite) TestListenerSinkIndividualBinary() {
	suite.testListenerSinkIndividual(new(BinaryMarshaller))
}

func (suite *TcpListenerTestSuite) TestListenerSinkAllCsv() {
	suite.testListenerSinkAll(new(CsvMarshaller))
}

func (suite *TcpListenerTestSuite) TestListenerSinkAllBinary() {
	suite.testListenerSinkAll(new(BinaryMarshaller))
}

func (suite *TcpListenerTestSuite) testListenerSourceAll(m Marshaller) {
	testSink := suite.newFilledTestSink()
	port := GetLocalPort()

	l := NewTcpListenerSource(":" + port)
	l.Reader = SampleReader{
		ParallelSampleHandler: parallel_handler,
	}

	s := &TCPSink{
		Endpoint:    "localhost:" + port,
		DialTimeout: tcpDialTimeout,
	}
	s.Writer.ParallelSampleHandler = parallel_handler
	s.SetMarshaller(m)

	sender := &oneShotTask{
		do: func() {
			suite.sendAllSamples(s)
		},
	}

	go func() {
		testSink.waitEmpty()
		l.Close()
		s.Close()
	}()

	suite.runGroup(sender, s, l, testSink)
}

func (suite *TcpListenerTestSuite) testListenerSourceIndividual(m BidiMarshaller) {
	for i := range suite.headers {
		testSink := suite.newTestSinkFor(i)
		port := GetLocalPort()

		l := NewTcpListenerSource(":" + port)
		l.Reader = SampleReader{
			ParallelSampleHandler: parallel_handler,
		}

		s := &TCPSink{
			Endpoint:    "localhost:" + port,
			DialTimeout: tcpDialTimeout,
		}
		s.Writer.ParallelSampleHandler = parallel_handler
		s.SetMarshaller(m)

		sender := &oneShotTask{
			do: func() {
				suite.sendSamples(s, i)
			},
		}

		go func() {
			testSink.waitEmpty()
			l.Close()
			s.Close()
		}()

		suite.runGroup(sender, s, l, testSink)
	}
}

type oneShotTask struct {
	do func()
}

func (t *oneShotTask) Start(wg *sync.WaitGroup) (_ golib.StopChan) {
	t.do()
	return
}

func (t *oneShotTask) Stop() {
}

func (t *oneShotTask) String() string {
	return "oneshot"
}

func (suite *TcpListenerTestSuite) TestListenerSourceIndividualCsv() {
	suite.testListenerSourceIndividual(new(CsvMarshaller))
}

func (suite *TcpListenerTestSuite) TestListenerSourceIndividualBinary() {
	suite.testListenerSourceIndividual(new(BinaryMarshaller))
}

func (suite *TcpListenerTestSuite) TestListenerSourceAllCsv() {
	suite.testListenerSourceAll(new(CsvMarshaller))
}

func (suite *TcpListenerTestSuite) TestListenerSourceAllBinary() {
	suite.testListenerSourceAll(new(BinaryMarshaller))
}

func (suite *TcpListenerTestSuite) TestTcpListenerSourceError() {
	// Suppress error output
	level := log.GetLevel()
	defer log.SetLevel(level)
	log.SetLevel(log.PanicLevel)

	l := NewTcpListenerSource("8.8.8.8:7777") // The IP should not be valid for the current host -> give error
	l.Reader = SampleReader{
		ParallelSampleHandler: parallel_handler,
	}

	var group golib.TaskGroup
	(&SamplePipeline{
		Source: l,
	}).Construct(&group)
	task, numErrs := group.WaitAndStop(1 * time.Second)
	suite.Equal(1, numErrs, "number of errors")
	suite.IsType(new(SourceTaskWrapper), task)
	suite.Equal(l, task.(*SourceTaskWrapper).SampleSource)
}
