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

// HTTPParams represents a map with string keys and a freely-chosen type. It is used to collect either GET or POST
// parameters for the ClouDNS API.
type HTTPParams map[string]interface{}

// Client provides the main object for interacting with the ClouDNS API. All service objects and settings are being
// stored underneath within this structure.
type Client struct {
	Account *AccountService
	Zones   *ZoneService
	Records *RecordService

	baseURL    string
	userAgent  string
	auth       *Auth
	headers    http.Header
	params     HTTPParams
	httpClient *http.Client
}

// StatusResult is a common result used by all ClouDNS API methods for either

type DataResult struct {
	ID int `json:"id"`
}
type StatusResult struct {
	Status            string     `json:"status"`
	StatusDescription string     `json:"statusDescription"`
	StatusMessage     string     `json:"statusMessage"`
	StatusData        DataResult `json:"data"`
}

// New instantiates a new ClouDNS client for interacting with the API
func New(options ...Option) (*Client, error) {
	client := &Client{
		baseURL:   "https://api.cloudns.net",
		userAgent: "cloudns-go",

		auth:       NewAuth(),
		headers:    make(http.Header),
		params:     make(HTTPParams),
		httpClient: http.DefaultClient,
	}

	if err := client.processOptions(options...); err != nil {
		return nil, ErrInvalidOptions.wrap(err)
	}

	client.Account = &AccountService{api: client}
	client.Zones = &ZoneService{api: client}
	client.Records = &RecordService{api: client}

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

func (c *Client) request(ctx context.Context, method, endpoint string, params HTTPParams, headers http.Header, target interface{}) error {
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

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, params HTTPParams, headers http.Header) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, ErrHTTPRequest.wrap(err)
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
			return nil, ErrHTTPRequest.wrap(err)
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
		return nil, ErrHTTPRequest.wrap(err)
	}
	if err := c.checkBaseResult(respBody); err != nil {
		return nil, err
	}

	if target != nil {
		if err := json.Unmarshal(respBody, target); err != nil {
			return nil, ErrHTTPRequest.wrap(err)
		}
	}

	return resp, nil
}

func (c *Client) checkBaseResult(respBody []byte) error {
	respBody = bytes.TrimLeft(respBody, " \t\r\n") // whitespace according to RFC7159.2

	switch {
	// If JSON response contains top-level object
	case len(respBody) > 0 && respBody[0] == '{':
		// Attempt to unmarshal JSON response into `StatusResult`
		var result StatusResult
		if err := json.Unmarshal(respBody, &result); err != nil {
			return ErrAPIInvocation.wrap(err)
		}

		// Skip further processing if API response does not indicate failure
		if result.Status != "Failed" {
			return nil
		}

		// Return an API error in all other cases, based on either `StatusDescription` or `StatusMessage`
		if result.StatusDescription != "" {
			return ErrAPIInvocation.wrap(errors.New(result.StatusDescription))
		} else if result.StatusMessage != "" {
			return ErrAPIInvocation.wrap(errors.New(result.StatusMessage))
		} else {
			return ErrAPIInvocation.wrap(errors.New(string(respBody)))
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
