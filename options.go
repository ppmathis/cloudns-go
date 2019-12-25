package cloudns

import (
	"net/http"
	"strings"
)

type Option func(api *API) error

func BaseURL(baseURL string) Option {
	return func(api *API) error {
		api.baseURL = strings.TrimRight(baseURL, "/")
		return nil
	}
}

func Headers(headers http.Header) Option {
	return func(api *API) error {
		api.headers = headers
		return nil
	}
}

func Params(params HttpParams) Option {
	return func(api *API) error {
		api.params = params
		return nil
	}
}

func UserAgent(userAgent string) Option {
	return func(api *API) error {
		api.userAgent = userAgent
		return nil
	}
}

func AuthUserID(id int, password string) Option {
	return func(api *API) error {
		if api.auth.Type != AuthTypeNone {
			return ErrMultipleCredentials
		}

		api.auth.Type = AuthTypeUserID
		api.auth.UserID = id
		api.auth.Password = password

		return nil
	}
}

func AuthSubUserID(id int, password string) Option {
	return func(api *API) error {
		if api.auth.Type != AuthTypeNone {
			return ErrMultipleCredentials
		}

		api.auth.Type = AuthTypeSubUserID
		api.auth.SubUserID = id
		api.auth.Password = password

		return nil
	}
}

func AuthSubUserName(user string, password string) Option {
	return func(api *API) error {
		if api.auth.Type != AuthTypeNone {
			return ErrMultipleCredentials
		}

		api.auth.Type = AuthTypeSubUserName
		api.auth.SubUserName = user
		api.auth.Password = password

		return nil
	}
}
