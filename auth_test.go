package cloudns

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuth_GetParams_None(t *testing.T) {
	// given
	auth := NewAuth()

	// when
	params := auth.GetParams()

	// then
	assert.Len(t, params, 0, "should return zero parameters")
}

func TestAuth_GetParams_AuthUserID(t *testing.T) {
	// given
	const userID int = 13
	const password string = "test"

	auth := AuthUserID(userID, password)
	client, err := New(auth)
	assert.NoError(t, err)

	// when
	params := client.auth.GetParams()

	// then
	assert.Len(t, params, 2, "should return two parameters")
	assert.Equal(t, userID, params["auth-id"], "parameter `auth-id` should match")
	assert.Equal(t, password, params["auth-password"], "parameter `auth-password` should match")
}

func TestAuth_GetParams_SubUserID(t *testing.T) {
	// given
	const subUserID int = 42
	const password string = "dummy"

	auth := AuthSubUserID(subUserID, password)
	client, err := New(auth)
	assert.NoError(t, err)

	// when
	params := client.auth.GetParams()

	// then
	assert.Len(t, params, 2, "should return two parameters")
	assert.Equal(t, subUserID, params["sub-auth-id"], "parameter `auth-id` should match")
	assert.Equal(t, password, params["auth-password"], "parameter `auth-password` should match")
}

func TestAuth_GetParams_SubUserName(t *testing.T) {
	// given
	const subUserName string = "hello"
	const password string = "world"

	auth := AuthSubUserName("hello", "world")
	client, err := New(auth)
	assert.NoError(t, err)

	// when
	params := client.auth.GetParams()

	// then
	assert.Len(t, params, 2, "should return two parameters")
	assert.Equal(t, subUserName, params["sub-auth-user"], "parameter `auth-id` should match")
	assert.Equal(t, password, params["auth-password"], "parameter `auth-password` should match")
}

func TestAuth_GetParams_Invalid(t *testing.T) {
	// given
	auth := NewAuth()
	auth.Type = -1

	// then
	assert.Panics(t, func() {
		auth.GetParams()
	}, "should panic with invalid auth type")
}
