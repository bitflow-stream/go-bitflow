package bitflow

import (
	"sync"
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
	testSink := suite.newFilledTestSink()

	l := NewTcpListenerSink(":7878", 100)
	l.Writer.ParallelSampleHandler = parallel_handler
	l.SetMarshaller(m)

	s := &TCPSource{
		PrintErrors:   true,
		RemoteAddrs:   []string{"localhost:7878"},
		RetryInterval: time.Second,
	}
	s.Reader.ParallelSampleHandler = parallel_handler
	s.SetSink(testSink)

	sender := &oneshotTask{
		do: func() {
			suite.sendAllSamples(l)
		},
	}

	go func() {
		testSink.waitEmpty()
		s.Stop()
		l.Close()
		l.Stop()
	}()

	group := golib.NewTaskGroup(l, s, sender)
	_, numErrs := group.WaitAndStop(1*time.Second, true)
	suite.Equal(0, numErrs, "number of errors")
}

func (suite *TcpListenerTestSuite) testListenerSinkIndividual(m Marshaller) {
	for i := range suite.headers {
		testSink := suite.newTestSinkFor(i)

		// TODO test that a smaller buffer leads to dropped samples

		l := NewTcpListenerSink(":7878", 100)
		l.Writer.ParallelSampleHandler = parallel_handler
		l.SetMarshaller(m)

		s := &TCPSource{
			PrintErrors:   true,
			RemoteAddrs:   []string{"localhost:7878"},
			RetryInterval: time.Second,
		}
		s.Reader.ParallelSampleHandler = parallel_handler
		s.SetSink(testSink)

		sender := &oneshotTask{
			do: func() {
				suite.sendSamples(l, i)
			},
		}

		go func() {
			testSink.waitEmpty()
			s.Stop()
			l.Close()
			l.Stop()
		}()

		group := golib.NewTaskGroup(l, s, sender)
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

func (suite *TcpListenerTestSuite) testListenerSourceAll(m Marshaller) {
	testSink := suite.newFilledTestSink()

	l := NewTcpListenerSource(":7878")
	l.Reader = SampleReader{
		ParallelSampleHandler: parallel_handler,
	}
	l.SetSink(testSink)

	s := &TCPSink{
		PrintErrors: true,
		Endpoint:    "localhost:7878",
	}
	s.Writer.ParallelSampleHandler = parallel_handler
	s.SetMarshaller(m)

	sender := &oneshotTask{
		do: func() {
			suite.sendAllSamples(s)
		},
	}

	go func() {
		testSink.waitEmpty()
		l.Stop()
		s.Close()
		s.Stop()
	}()

	group := golib.NewTaskGroup(l, s, sender)
	_, numErrs := group.WaitAndStop(1*time.Second, true)
	suite.Equal(0, numErrs, "number of errors")
}

func (suite *TcpListenerTestSuite) testListenerSourceIndividual(m BidiMarshaller) {
	for i := range suite.headers {
		testSink := suite.newTestSinkFor(i)

		l := NewTcpListenerSource(":7878")
		l.Reader = SampleReader{
			ParallelSampleHandler: parallel_handler,
		}
		l.SetSink(testSink)

		s := &TCPSink{
			PrintErrors: true,
			Endpoint:    "localhost:7878",
		}
		s.Writer.ParallelSampleHandler = parallel_handler
		s.SetMarshaller(m)

		sender := &oneshotTask{
			do: func() {
				suite.sendSamples(s, i)
			},
		}

		go func() {
			testSink.waitEmpty()
			l.Stop()
			s.Close()
			s.Stop()
		}()

		group := golib.NewTaskGroup(l, s, sender)
		_, numErrs := group.WaitAndStop(1*time.Second, true)
		suite.Equal(0, numErrs, "number of errors")
	}
}

type oneshotTask struct {
	do func()
}

func (t *oneshotTask) Start(wg *sync.WaitGroup) golib.StopChan {
	t.do()
	return nil
}

func (t *oneshotTask) Stop() {
}

func (t *oneshotTask) String() string {
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
