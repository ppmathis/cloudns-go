package cloudns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HttpParams map[string]interface{}

type Client struct {
	Account *accountService
	Zones   *zoneService
	Records *recordService

	baseURL    string
	userAgent  string
	auth       *Auth
	headers    http.Header
	params     HttpParams
	httpClient *http.Client
}

type BaseResult struct {
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
	StatusMessage     string `json:"statusMessage"`
}

func New(options ...Option) (*Client, error) {
	client := &Client{
		baseURL:   "https://api.cloudns.net",
		userAgent: "cloudns-go",

		auth:       NewAuth(),
		headers:    make(http.Header),
		params:     make(HttpParams),
		httpClient: http.DefaultClient,
	}

	if err := client.processOptions(options...); err != nil {
		return nil, ErrInvalidOptions.wrap(err)
	}

	client.Account = &accountService{api: client}
	client.Zones = &zoneService{api: client}
	client.Records = &recordService{api: client}

	return client, nil
}

func (c *Client) processOptions(options ...Option) error {
	for _, option := range options {
		if err := option(c); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) request(ctx context.Context, method, endpoint string, params HttpParams, headers http.Header, target interface{}) error {
	req, err := c.makeRequest(ctx, method, endpoint, params, headers)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req, target)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, params HttpParams, headers http.Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, ErrHttpRequest.wrap(err)
	}

	mergedHeaders := make(http.Header)
	copyHeaders(mergedHeaders, c.headers)
	copyHeaders(mergedHeaders, headers)

	req.Header = mergedHeaders
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	mergedParams := make(map[string]interface{})
	copyParams(mergedParams, c.params)
	copyParams(mergedParams, c.auth.GetParams())
	copyParams(mergedParams, params)

	if containsString(method, []string{"HEAD", "GET", "DELETE"}) {
		queryValues := make(url.Values)
		for key, value := range mergedParams {
			queryValues.Set(key, fmt.Sprintf("%s", value))
		}

		req.URL.RawQuery = queryValues.Encode()
	} else {
		jsonBody, err := json.Marshal(mergedParams)
		if err != nil {
			return nil, ErrHttpRequest.wrap(err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Body = ioutil.NopCloser(bytes.NewBuffer(jsonBody))
	}

	return req, nil
}

func (c *Client) doRequest(req *http.Request, target interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrHttpRequest.wrap(err)
	}

	fmt.Println(string(respBody))
	if err := c.checkBaseResult(respBody); err != nil {
		return nil, err
	}

	if target != nil {
		if err := json.Unmarshal(respBody, target); err != nil {
			return nil, ErrHttpRequest.wrap(err)
		}
	}

	return resp, nil
}

func (c *Client) checkBaseResult(respBody []byte) error {
	respBody = bytes.TrimLeft(respBody, " \t\r\n") // whitespace according to RFC7159.2

	switch {
	// If JSON response contains top-level object
	case len(respBody) > 0 && respBody[0] == '{':
		// Attempt to unmarshal JSON response into `BaseResult`
		var result BaseResult
		if err := json.Unmarshal(respBody, &result); err != nil {
			return ErrApiInvocation.wrap(err)
		}

		// Skip further processing if API response does not indicate failure
		if result.Status != "Failed" {
			return nil
		}

		// Return an API error in all other cases, based on either `StatusDescription` or `StatusMessage`
		if result.StatusDescription != "" {
			return ErrApiInvocation.wrap(errors.New(result.StatusDescription))
		} else if result.StatusMessage != "" {
			return ErrApiInvocation.wrap(errors.New(result.StatusMessage))
		} else {
			return ErrApiInvocation.wrap(errors.New(string(respBody)))
		}
	}

	return nil
}

func copyHeaders(target, source http.Header) {
	if source == nil {
		return
	}

	for key, value := range source {
		target[key] = value
	}
}

func copyParams(target, source map[string]interface{}) {
	if source == nil {
		return
	}

	for key, value := range source {
		target[key] = value
	}
}
