package cloudns

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConstError_Error(t *testing.T) {
	// when
	err := constError("Hello World")

	// then
	assert.Error(t, err, "constError() should return error")
	assert.Equal(t, "Hello World", err.Error(), "constError() should be described as `Hello World`")
}

func TestWrapError_Error(t *testing.T) {
	// given
	innerErr := constError("World")
	outerErr := constError("Hello")

	// when
	wrapErr := outerErr.wrap(innerErr)

	// then
	assert.Error(t, wrapErr, "wrapErr should contain err")
	assert.Equal(t, "Hello: World", wrapErr.Error(), "wrapErr should be described as `Hello: World`")
}

func TestWrapError_Error_InnerNil(t *testing.T) {
	// given
	var innerErr error = nil
	outerErr := constError("Hello")

	// when
	noWrapErr := outerErr.wrap(innerErr)

	// then
	assert.Error(t, noWrapErr, "noWrapErr should contain err")
	assert.Equal(t, "Hello", noWrapErr.Error(), "noWrapErr should be described as `Hello`")
}

func TestWrapError_Unwrap(t *testing.T) {
	// given
	innerErr := constError("World")
	outerErr := constError("Hello")
	wrapErr := outerErr.wrap(innerErr)

	// then
	assert.Error(t, wrapErr, "wrapErr should contain err")
	assert.True(t, errors.Is(wrapErr, outerErr), "errors.Is(wrapErr, outerErr) should return true")
	assert.True(t, errors.Is(wrapErr, innerErr), "errors.Is(wrapErr, innerErr) should return true")
}
