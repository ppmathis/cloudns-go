package cloudns

import (
	"net/http"
	"strings"
)

// Option represents functional options which can be specified when instantiating a new API client
type Option func(api *Client) error

// BaseURL modifies the base URL of the API client
func BaseURL(baseURL string) Option {
	return func(api *Client) error {
		api.baseURL = strings.TrimRight(baseURL, "/")
		return nil
	}
}

// Headers adds a set of headers to every sent API request. These headers can be overridden by request-specific headers.
func Headers(headers http.Header) Option {
	return func(api *Client) error {
		api.headers = headers
		return nil
	}
}

// Params adds a set of parameters (either GET or POST) to every sent API request. These are overridden by auth as well
// as request-specific parameters.
func Params(params HTTPParams) Option {
	return func(api *Client) error {
		api.params = params
		return nil
	}
}

// HTTPClient overrides the HTTPClient used by the API client, useful for mocking in unit tests.
func HTTPClient(httpClient *http.Client) Option {
	return func(api *Client) error {
		api.httpClient = httpClient
		return nil
	}
}

// UserAgent overrides the default user agent of cloudns-go.
func UserAgent(userAgent string) Option {
	return func(api *Client) error {
		api.userAgent = userAgent
		return nil
	}
}

// AuthUserID setups user-id based authentication against the ClouDNS API
func AuthUserID(id int, password string) Option {
	return func(api *Client) error {
		if api.auth.Type != AuthTypeNone {
			return ErrMultipleCredentials
		}

		api.auth.Type = AuthTypeUserID
		api.auth.UserID = id
		api.auth.Password = password

		return nil
	}
}

// AuthSubUserID setups sub-user-id based authentication against the ClouDNS API
func AuthSubUserID(id int, password string) Option {
	return func(api *Client) error {
		if api.auth.Type != AuthTypeNone {
			return ErrMultipleCredentials
		}

		api.auth.Type = AuthTypeSubUserID
		api.auth.SubUserID = id
		api.auth.Password = password

		return nil
	}
}

// AuthSubUserName setups the sub-user-name based authentication against the ClouDNS API
func AuthSubUserName(user string, password string) Option {
	return func(api *Client) error {
		if api.auth.Type != AuthTypeNone {
			return ErrMultipleCredentials
		}

		api.auth.Type = AuthTypeSubUserName
		api.auth.SubUserName = user
		api.auth.Password = password

		return nil
	}
}
