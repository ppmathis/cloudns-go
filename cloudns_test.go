package cloudns

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
)

const testDomain string = "api-example.com"
const testTTL int = 3600

var (
	vcr    *recorder.Recorder
	client *Client
	ctx    context.Context
)

func setup(t *testing.T) func() {
	var err error

	// Determine recorder mode based on environment
	recorderMode := recorder.ModeReplayWithNewEpisodes
	if os.Getenv("CLOUDNS_SKIP_FIXTURES") != "" {
		recorderMode = recorder.ModePassthrough
	}

	// Initialize test fixtures with go-vcr for automated recording
	vcr, err = recorder.NewWithOptions(&recorder.Options{
		CassetteName:       "fixtures/" + t.Name(),
		Mode:               recorderMode,
		SkipRequestLatency: true,
	})
	if err != nil {
		log.Fatalf("could not initialize test fixtures: %v", err)
	}
	vcr.AddHook(filterCookies, recorder.AfterCaptureHook)
	vcr.AddHook(filterCredentials, recorder.AfterCaptureHook)
	vcr.AddHook(filterZoneListing, recorder.AfterCaptureHook)

	// Initialize API client with go-vcr as HTTP client transport
	client, err = New(
		buildAuthFromEnv(),
		HTTPClient(&http.Client{Transport: vcr}),
		UserAgent("cloudns-go/test"),
	)
	if err != nil {
		panic(err)
	}

	// Initialize default context
	ctx = context.Background()

	// Return teardown function
	return func() {
		if err := vcr.Stop(); err != nil {
			log.Fatalf("could not stop test recorer: %v", err)
		}
	}
}

func buildAuthFromEnv() Option {
	if os.Getenv("CLOUDNS_USER_ID") == "" || os.Getenv("CLOUDNS_PASSWORD") == "" {
		return func(api *Client) error {
			return nil
		}
	}

	userPassword := os.Getenv("CLOUDNS_PASSWORD")
	userID, err := strconv.Atoi(os.Getenv("CLOUDNS_USER_ID"))
	if err != nil {
		log.Fatalf("could not convert CLOUDNS_USER_ID to integer: %v", err)
	}

	return AuthUserID(userID, userPassword)
}

func filterCookies(i *cassette.Interaction) error {
	delete(i.Request.Headers, "Cookie")
	delete(i.Response.Headers, "Set-Cookie")

	return nil
}

func filterCredentials(i *cassette.Interaction) error {
	var jsonData map[string]interface{}

	if err := json.Unmarshal([]byte(i.Request.Body), &jsonData); err != nil {
		return fmt.Errorf("could not unmarshal request body as JSON for filtering: %w", err)
	}

	for _, key := range client.auth.getAllParamKeys() {
		if _, ok := jsonData[key]; ok {
			jsonData[key] = "[filtered]"
		}
	}

	jsonBody, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("could not marshal filtered request body into JSON: %w", err)
	}

	i.Request.Body = string(jsonBody)
	return nil
}

func filterZoneListing(i *cassette.Interaction) error {
	var retrievedZones []map[string]interface{}
	var filteredZones []map[string]interface{}

	if !strings.HasSuffix(i.Request.URL, "/list-zones.json") {
		return nil
	}

	if err := json.Unmarshal([]byte(i.Response.Body), &retrievedZones); err != nil {
		return fmt.Errorf("could not unmarshal request body as JSON for filtering: %w", err)
	}

	for _, zone := range retrievedZones {
		if zone["name"] == testDomain {
			filteredZones = append(filteredZones, zone)
		}
	}

	jsonBody, err := json.Marshal(filteredZones)
	if err != nil {
		return fmt.Errorf("could not marshal filtered request body into JSON: %w", err)
	}

	i.Response.Body = string(jsonBody)
	return nil
}
