package bitflow

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const _samples_per_test = 5

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
	suite.headers = []*Header{
		&Header{
			Fields:  []string{"a", "b", "c"},
			HasTags: true,
		},
		&Header{
			Fields:  []string{"x"},
			HasTags: false,
		},
		&Header{
			Fields:  nil,
			HasTags: true,
		},
		&Header{
			Fields:  nil,
			HasTags: false,
		},
	}
	suite.samples = make([][]*Sample, len(suite.headers))
	for i, header := range suite.headers {
		suite.samples[i] = suite.makeSamples(header)
	}
}

func (suite *testSuiteWithSamples) compareSamples(expected *Sample, sample *Sample) {
	suite.Equal(expected.tags, sample.tags)
	suite.Equal(expected.orderedTags, sample.orderedTags)
	suite.Equal(expected.Values, sample.Values)
	suite.True(expected.Time.Equal(sample.Time), fmt.Sprintf("Times differ: expected %v, but was %v", expected.Time, sample.Time))
}

func (suite *testSuiteWithSamples) makeSamples(header *Header) (res []*Sample) {
	for i := 0; i < _samples_per_test; i++ {
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
