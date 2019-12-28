package cloudns

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsString(t *testing.T) {
	result1 := containsString("yes", []string{"no", "yes", "maybe"})
	assert.True(t, result1, "`yes` should be found inside string slice")

	result2 := containsString("what", []string{"no", "yes", "maybe"})
	assert.False(t, result2, "`what` should not be found inside string slice")
}

func TestAPIBool_MarshalJSON(t *testing.T) {
	trueResult, err := json.Marshal(APIBool(true))
	assert.NoError(t, err, "JSON marshalling of APIBool with `true` value should not fail")
	assert.Equal(t, []byte("1"), trueResult, "JSON for APIBool with `true` should return `1`")

	falseResult, err := json.Marshal(APIBool(false))
	assert.NoError(t, err, "JSON marshalling of APIBool with `false` value should not fail")
	assert.Equal(t, []byte("0"), falseResult, "JSON for APIBool with `false` should return `0`")
}

func TestAPIBool_UnmarshalJSON(t *testing.T) {
	test := func(value string, expected APIBool) {
		var actual APIBool
		err := json.Unmarshal([]byte(value), &actual)
		assert.NoError(t, err, "JSON unmarshalling of APIBool(%s) should not fail", value)
		assert.Equal(t, expected, actual, "Unmarshalled APIBool(%s) should return %t", expected)
	}

	test(`true`, true)
	test(`1`, true)
	test(`"1"`, true)
	test(`"true"`, true)

	test(`false`, false)
	test(`0`, false)
	test(`"0"`, false)
	test(`"false"`, false)
}
