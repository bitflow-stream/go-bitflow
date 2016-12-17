package bitflow

import (
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/antongulenko/golib"
	"github.com/stretchr/testify/require"
)

var parallel_handler = ParallelSampleHandler{
	BufferedSamples: 5,
	ParallelParsers: 6,
}

var debug_tests = false

func init() {
	//debug_tests = true
	if debug_tests {
		golib.LogVerbose = true
	} else {
		WarnObsoleteBinaryFormat = false
		golib.LogQuiet = true
	}
	golib.ConfigureLogging()
}

type testSuiteWithSamples struct {
	t *testing.T
	*require.Assertions

	rand      *rand.Rand
	timestamp time.Time

	headers []*Header
	samples [][]*Sample
}

func (suite *testSuiteWithSamples) T() *testing.T {
	return suite.t
}

func (suite *testSuiteWithSamples) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (suite *testSuiteWithSamples) SetupTest() {
	suite.timestamp = time.Now()
	suite.rand = rand.New(rand.NewSource(123)) // deterministic
	headers := []*Header{
		&Header{
			Fields: []string{"a", "b", "c"},
		},
		&Header{
			Fields: []string{" ", "_", "="},
		},
		&Header{
			Fields: []string{"x"},
		},
		&Header{
			Fields: nil,
		},
	}

	// Add a copy of every header, now with the HasTags flag enabled
	n := len(headers)
	for i := 0; i < n; i++ {
		headers = append(headers, &Header{
			Fields:  headers[i].Fields,
			HasTags: true,
		})
	}

	suite.headers = make([]*Header, len(headers)*3)
	suite.samples = make([][]*Sample, len(suite.headers))
	step := len(headers)
	for i, header := range headers {
		suite.headers[i+0*step] = header.Clone(header.Fields)
		suite.samples[i+0*step] = suite.makeSamples(header, 5)
		suite.headers[i+1*step] = header.Clone(header.Fields)
		suite.samples[i+1*step] = suite.makeSamples(header, 3)
		suite.headers[i+2*step] = header.Clone(header.Fields)
		suite.samples[i+2*step] = suite.makeSamples(header, 1)
	}
}

func (suite *testSuiteWithSamples) compareSamples(expected *Sample, sample *Sample) {
	suite.Equal(expected.tags, sample.tags, "Sample.tags")
	suite.Equal(expected.orderedTags, sample.orderedTags, "Sample.orderedTags")
	suite.Equal(expected.Values, sample.Values, "Sample.Values")
	suite.True(expected.Time.Equal(sample.Time), fmt.Sprintf("Times differ: expected %v, but was %v", expected.Time, sample.Time))
}

func (suite *testSuiteWithSamples) compareHeaders(expected *Header, header *Header) {
	suite.Equal(expected.HasTags, header.HasTags, "Header.HasTags")
	suite.Equal(expected.Fields, header.Fields, "Header.Fields")
}

func (suite *testSuiteWithSamples) makeSamples(header *Header, num int) (res []*Sample) {
	for i := 0; i < num; i++ {
		sample := &Sample{
			Values: suite.makeValues(header),
			Time:   suite.nextTimestamp(),
		}
		res = append(res, sample)
		suite.setTags(sample, header)
	}
	return
}

func (suite *testSuiteWithSamples) makeValues(header *Header) []Value {
	var res []Value
	for _ = range header.Fields {
		res = append(res, Value(suite.rand.Float64()))
	}
	return res
}

func (suite *testSuiteWithSamples) setTags(sample *Sample, header *Header) {
	if !header.HasTags {
		return
	}
	for key, val := range map[string]string{
		suite.randomString(): suite.randomString(),
		suite.randomString(): suite.randomString(),
	} {
		sample.SetTag(key, val)
	}
}

func (suite *testSuiteWithSamples) randomString() string {
	return fmt.Sprintf("string_%v", suite.rand.Int())
}

func (suite *testSuiteWithSamples) nextTimestamp() time.Time {
	res := suite.timestamp
	suite.timestamp = res.Add(10 * time.Second)
	return res
}

func (suite *testSuiteWithSamples) newTestSinkFor(headerIndex int) *testSampleSink {
	s := &testSampleSink{
		suite:     suite,
		emptyCond: sync.NewCond(new(sync.Mutex)),
	}
	s.add(suite.samples[headerIndex], suite.headers[headerIndex])
	return s
}

func (suite *testSuiteWithSamples) newFilledTestSink() *testSampleSink {
	s := &testSampleSink{
		suite:     suite,
		emptyCond: sync.NewCond(new(sync.Mutex)),
	}
	for i, header := range suite.headers {
		s.add(suite.samples[i], header)
	}
	return s
}

func (suite *testSuiteWithSamples) sendSamples(w MetricSinkBase, headerIndex int) (res int) {
	for _, sample := range suite.samples[headerIndex] {
		suite.NoError(w.Sample(sample, suite.headers[headerIndex]))
		res++
	}
	return
}

func (suite *testSuiteWithSamples) sendAllSamples(w MetricSinkBase) (res int) {
	for i := range suite.headers {
		res += suite.sendSamples(w, i)
	}
	return
}

type countingBuf struct {
	data   []byte
	closed bool
}

func (c *countingBuf) Read(b []byte) (num int, err error) {
	num = copy(b, c.data)
	c.data = c.data[num:]
	if num < len(b) || len(c.data) == 0 {
		err = io.EOF
	}

	return
}

func (c *countingBuf) Close() error {
	c.closed = true
	return nil
}

func (c *countingBuf) checkClosed(r *require.Assertions) {
	r.True(c.closed, "Counting buf not closed")
}

type testSampleSink struct {
	suite     *testSuiteWithSamples
	emptyCond *sync.Cond
	samples   []*Sample
	headers   []*Header
	received  int
	closed    bool
}

func (s *testSampleSink) add(samples []*Sample, header *Header) {
	s.samples = append(s.samples, samples...)
	for _ = range samples {
		s.headers = append(s.headers, header)
	}
}

func (s *testSampleSink) Sample(sample *Sample, header *Header) error {
	s.emptyCond.L.Lock()
	defer s.emptyCond.L.Unlock()

	if s.received > len(s.samples)-1 {
		s.suite.Fail("Did not expect any more samples, received " + strconv.Itoa(s.received))
	}
	expectedSample := s.samples[s.received]
	expectedHeader := s.headers[s.received]
	s.received++
	s.suite.compareHeaders(expectedHeader, header)
	s.suite.compareSamples(expectedSample, sample)
	s.emptyCond.Broadcast()
	return nil
}

func (s *testSampleSink) checkEmpty() {
	s.suite.Equal(len(s.samples), s.received, "Number of samples received")
}

func (s *testSampleSink) waitEmpty() {
	s.emptyCond.L.Lock()
	defer s.emptyCond.L.Unlock()
	for s.received < len(s.samples) {
		s.emptyCond.Wait()
	}
}

func (s *testSampleSink) Close() {
	s.closed = true
}

func (s *testSampleSink) checkClosed() {
	s.suite.True(s.closed, "test sink closed")
}

func (s *testSampleSink) Start(_ *sync.WaitGroup) golib.StopChan {
	s.suite.Fail("testSampleSink.Start() called")
	return nil
}

func (s *testSampleSink) Stop() {
	s.suite.Fail("testSampleSink.Stop() called")
}

func (s *testSampleSink) String() string {
	s.suite.Fail("testSampleSink.String() called")
	return ""
}

type testSampleHandler struct {
	suite  *testSuiteWithSamples
	source string
}

func (suite *testSuiteWithSamples) newHandler(source string) *testSampleHandler {
	return &testSampleHandler{
		suite:  suite,
		source: source,
	}
}

func (h *testSampleHandler) HandleHeader(header *Header, source string) {
	h.suite.Equal(h.source, source)
}

func (h *testSampleHandler) HandleSample(sample *Sample, source string) {
	h.suite.Equal(h.source, source)
}
