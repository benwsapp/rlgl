package slack_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benwsapp/rlgl/pkg/slack"
)

func TestSetStatusSuccess(t *testing.T) {
	t.Parallel()

	server := createMockServer(t, slack.ProfileResponse{Ok: true})
	defer server.Close()

	client := slack.NewClient("test-token").WithAPIURL(server.URL)

	err := client.SetStatus("Working on feature", ":computer:", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetStatusSlackAPIError(t *testing.T) {
	t.Parallel()

	server := createMockServer(t, slack.ProfileResponse{
		Ok:    false,
		Error: "invalid_auth",
	})
	defer server.Close()

	client := slack.NewClient("test-token").WithAPIURL(server.URL)

	err := client.SetStatus("Test status", ":test:", 0)
	if err == nil {
		t.Error("expected error but got none")
	}

	if !strings.Contains(err.Error(), "slack API error") {
		t.Errorf("expected slack API error, got: %v", err)
	}
}

func TestSetStatusEmptyStatus(t *testing.T) {
	t.Parallel()

	server := createMockServer(t, slack.ProfileResponse{Ok: true})
	defer server.Close()

	client := slack.NewClient("test-token").WithAPIURL(server.URL)

	err := client.SetStatus("", "", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetStatusLongTextTruncation(t *testing.T) {
	t.Parallel()

	const longStatusText = "This is a very long status text that exceeds" +
		" the maximum allowed length of 100 characters and should be truncated"

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
		var profileReq slack.ProfileRequest

		err := json.NewDecoder(req.Body).Decode(&profileReq)
		if err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		// Verify status text length
		if len(profileReq.Profile.StatusText) > 100 {
			t.Errorf("status text exceeds max length: %d > 100", len(profileReq.Profile.StatusText))
		}

		responseWriter.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(responseWriter).Encode(slack.ProfileResponse{Ok: true})
		if err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := slack.NewClient("test-token").WithAPIURL(server.URL)

	err := client.SetStatus(longStatusText, ":long:", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClearStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
		var profileReq slack.ProfileRequest

		err := json.NewDecoder(req.Body).Decode(&profileReq)
		if err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if profileReq.Profile.StatusText != "" {
			t.Errorf("expected empty status text, got %s", profileReq.Profile.StatusText)
		}

		if profileReq.Profile.StatusEmoji != "" {
			t.Errorf("expected empty status emoji, got %s", profileReq.Profile.StatusEmoji)
		}

		responseWriter.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(responseWriter).Encode(slack.ProfileResponse{Ok: true})
		if err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := slack.NewClient("test-token").WithAPIURL(server.URL)

	err := client.ClearStatus()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func createMockServer(t *testing.T, response slack.ProfileResponse) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", req.Method)
		}

		if req.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("expected Authorization header with Bearer token")
		}

		if req.Header.Get("Content-Type") != "application/json; charset=utf-8" {
			t.Error("expected Content-Type application/json")
		}

		responseWriter.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(responseWriter).Encode(response)
		if err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
}
