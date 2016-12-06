package bitflow

import (
	"log"
	"testing"
	"time"

	"github.com/antongulenko/golib"
	"github.com/stretchr/testify/suite"
)

type TcpListenerTestSuite struct {
	testSuiteWithSamples
}

func TestTcpListener(t *testing.T) {
	suite.Run(t, new(TcpListenerTestSuite))
}

func (suite *TcpListenerTestSuite) testListenerSinkAll(m BidiMarshaller) {
	log.Println("======================= STARTING combined listener test for", m)
	ph := ParallelSampleHandler{
		BufferedSamples: 5,
		ParallelParsers: 5,
	}
	testSink := suite.newFilledTestSink()

	l := NewTcpListenerSink(":7878", 100)
	l.Writer.ParallelSampleHandler = ph
	l.Marshaller = m

	s := &TCPSource{
		PrintErrors:   true,
		RemoteAddrs:   []string{"localhost:7878"},
		RetryInterval: time.Second,
	}
	s.Reader.ParallelSampleHandler = ph
	s.SetSink(testSink)

	// This is okay before Start()
	suite.sendAllSamples(l)

	go func() {
		testSink.waitEmpty()
		s.Stop()
		l.Close()
		l.Stop()
	}()

	group := golib.NewTaskGroup(l, s)
	_, numErrs := group.WaitAndStop(1*time.Second, true)
	suite.Equal(0, numErrs, "number of errors")
}

func (suite *TcpListenerTestSuite) testListenerSinkIndividual(m BidiMarshaller) {
	for i := range suite.headers {
		//	{
		//		i := 3

		log.Println("======================= STARTING listener test:", i, m)

		ph := ParallelSampleHandler{
			BufferedSamples: 5,
			ParallelParsers: 5,
		}
		testSink := suite.newTestSinkFor(i)

		l := NewTcpListenerSink(":7878", 100)
		l.Writer.ParallelSampleHandler = ph
		l.Marshaller = m

		s := &TCPSource{
			PrintErrors:   true,
			RemoteAddrs:   []string{"localhost:7878"},
			RetryInterval: time.Second,
		}
		s.Reader.ParallelSampleHandler = ph
		s.SetSink(testSink)

		// This is okay before Start()
		suite.sendSamples(l, i)

		go func() {
			testSink.waitEmpty()
			s.Stop()
			l.Close()
			l.Stop()
		}()

		group := golib.NewTaskGroup(l, s)
		_, numErrs := group.WaitAndStop(1*time.Second, true)
		suite.Equal(0, numErrs, "number of errors")
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
