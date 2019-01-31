package steps

import (
	"testing"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	testAssert "github.com/stretchr/testify/assert"
)

func TestFailedRegexMetricSplitter(t *testing.T) {
	assert := testAssert.New(t)
	splitter, err := NewMetricSplitter([]string{"cc***"})
	assert.Error(err)
	assert.Nil(splitter)
}

func _makeSplitter(t *testing.T) *MetricSplitter {
	assert := testAssert.New(t)
	splitter, err := NewMetricSplitter([]string{"^a(?P<key1>x*)b(?P<key2>x*)c$", "A(?P<key3>x*)B"})
	assert.NoError(err)
	assert.NotNil(splitter)
	return splitter
}

func _test(t *testing.T, splitter *MetricSplitter, valuesIn []bitflow.Value, headerIn []string, valuesOut [][]bitflow.Value, headerOut [][]string, tagsOut []map[string]string) {
	assert := testAssert.New(t)

	sample := &bitflow.Sample{Values: valuesIn}

	// Expect the default tags everywhere
	sample.SetTag("k1", "v1")
	sample.SetTag("k2", "v2")
	for _, m := range tagsOut {
		m["k1"] = "v1"
		m["k2"] = "v2"
	}

	// Repeat to test the cache
	for i := 0; i < 3; i++ {
		res := splitter.Split(sample, &bitflow.Header{Fields: headerIn})
		assert.Len(res, len(valuesOut))

		for i, expectedValues := range valuesOut {
			assert.Equal(expectedValues, res[i].Sample.Values)
			assert.Equal(headerOut[i], res[i].Header.Fields)
			assert.Equal(tagsOut[i], res[i].Sample.TagMap())
		}
	}
}

func TestMetricSplitterNoMatch(t *testing.T) {
	splitter := _makeSplitter(t)
	_test(t, splitter,
		[]bitflow.Value{1, 2, 3},
		[]string{"1", "2", "3"},
		[][]bitflow.Value{{1, 2, 3}},
		[][]string{{"1", "2", "3"}},
		[]map[string]string{{}},
	)
	_test(t, splitter,
		[]bitflow.Value{4, 5, 6},
		[]string{"4", "5", "6"},
		[][]bitflow.Value{{4, 5, 6}},
		[][]string{{"4", "5", "6"}},
		[]map[string]string{{}},
	)
}

func TestMetricSplitterOneMatch(t *testing.T) {
	splitter := _makeSplitter(t)
	_test(t, splitter,
		[]bitflow.Value{1, 2, 3},
		[]string{"1", "axxbxxxc", "3"},
		[][]bitflow.Value{{1, 3}, {2}},
		[][]string{{"1", "3"}, {"axxbxxxc"}},
		[]map[string]string{{}, {"key1": "xx", "key2": "xxx"}},
	)
	_test(t, splitter,
		[]bitflow.Value{4, 5, 6},
		[]string{"4", "5", "AxxxxB"},
		[][]bitflow.Value{{4, 5}, {6}},
		[][]string{{"4", "5"}, {"AxxxxB"}},
		[]map[string]string{{}, {"key3": "xxxx"}},
	)
}

func TestMetricSplitterMultiMatch(t *testing.T) {
	splitter := _makeSplitter(t)
	_test(t, splitter,
		[]bitflow.Value{1, 2, 3, 4, 5, 6, 7},
		[]string{"1", "axxbxxxc", "AxxxxB", "4", "axxxxxbxxxxxxc", "AxxxxB", "axxbxxxc"},
		[][]bitflow.Value{{1, 4}, {2, 7}, {3, 6}, {5}},
		[][]string{{"1", "4"}, {"axxbxxxc", "axxbxxxc"}, {"AxxxxB", "AxxxxB"}, {"axxxxxbxxxxxxc"}},
		[]map[string]string{{}, {"key1": "xx", "key2": "xxx"}, {"key3": "xxxx"}, {"key1": "xxxxx", "key2": "xxxxxx"}},
	)
	_test(t, splitter,
		[]bitflow.Value{11, 12, 13, 14, 15, 16, 17},
		[]string{"AxxxxB", "AxxxxB", "13", "axxbxxxc", "axxxxxbxxxxxxc", "axxbxxxc", "AxxxxB"},
		[][]bitflow.Value{{13}, {11, 12, 17}, {14, 16}, {15}},
		[][]string{{"13"}, {"AxxxxB", "AxxxxB", "AxxxxB"}, {"axxbxxxc", "axxbxxxc"}, {"axxxxxbxxxxxxc"}},
		[]map[string]string{{}, {"key3": "xxxx"}, {"key1": "xx", "key2": "xxx"}, {"key1": "xxxxx", "key2": "xxxxxx"}},
	)
}
