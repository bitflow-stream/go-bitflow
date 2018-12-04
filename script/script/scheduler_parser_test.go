package script

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAntlrBitflowScriptScheduleListener_ParseScript(t *testing.T) {
	scripts, errs := new(BitflowScriptScheduleParser).ParseScript("./in_data ->concatenate()[cpu-tag='high']->window(){asd()->bcd()}->avg()[cpu-tag='low', someOtherHint=5]->./out_data")

	assert.NoError(t, errs.NilOrError())
	assert.Len(t, scripts, 3)
	assertValues(t, scripts[0], 0, "./in_data ->{{1}}->window(){asd()->bcd()}->{{2}}->./out_data", nil)
	assertValues(t, scripts[1], 1, "concatenate()[cpu-tag='high']", map[string]string{"cpu-tag": "high"})
	assertValues(t, scripts[2], 2, "avg()[cpu-tag='low', someOtherHint=5]", map[string]string{"cpu-tag": "low", "someOtherHint": "5"})
}

func TestAntlrBitflowScriptScheduleListener_ParseScriptWithPropagateDown(t *testing.T) {
	scripts, errs := new(BitflowScriptScheduleParser).ParseScript("./in_data ->concatenate()[propagate-down='true',cpu-tag='high']->window(){asd()->bcd()}->avg()->./out_data")

	assert.Nil(t, errs.NilOrError())
	assert.Len(t, scripts, 2)
	assertValues(t, scripts[0], 0, "./in_data ->{{1}}", nil)
	assertValues(t, scripts[1], 1, "concatenate()[propagate-down='true',cpu-tag='high']->window(){asd()->bcd()}->avg()->./out_data", map[string]string{"propagate-down": "true", "cpu-tag": "high"})

}

func TestAntlrBitflowScriptScheduleListener_ParseScriptWithPropagateDownAndInterrupt(t *testing.T) {
	scripts, errs := new(BitflowScriptScheduleParser).ParseScript("./in_data ->concatenate()[propagate-down='true',cpu-tag='high']->window(){asd()->bcd()}->avg()[newHints='sampleValue']->./out_data")

	assert.Nil(t, errs.NilOrError())
	assert.Len(t, scripts, 3)
	assertValues(t, scripts[0], 0, "./in_data ->{{1}}->{{2}}->./out_data", nil)
	assertValues(t, scripts[1], 1, "concatenate()[propagate-down='true',cpu-tag='high']->window(){asd()->bcd()}", map[string]string{"propagate-down": "true", "cpu-tag": "high"})
	assertValues(t, scripts[2], 2, "avg()[newHints='sampleValue']", map[string]string{"newHints": "sampleValue"})
}

func assertValues(t *testing.T, receivedHS HintedSubscript, expectedIndex int, expectedScript string, expectedHints map[string]string) {
	t.Helper()

	assert.Equal(t, expectedIndex, int(receivedHS.Index))
	assert.Equal(t, expectedScript, receivedHS.Script)

	// expected keys have expected values
	for k, expectedVal := range expectedHints {
		assert.Equal(t, expectedVal, receivedHS.Hints[k], "Expected key was missing or did not contain expected value")
	}

	for k := range receivedHS.Hints {
		if _, ok := expectedHints[k]; !ok {
			assert.Fail(t, "Scheduling hints contained unexpected key: "+k)
		}

	}
	assert.Equal(t, receivedHS.Hints, expectedHints)
}
