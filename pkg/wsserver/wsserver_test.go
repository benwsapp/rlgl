package wsserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benwsapp/rlgl/pkg/embed"
	"github.com/benwsapp/rlgl/pkg/wsserver"
	"github.com/gorilla/websocket"
)

const testToken = "test-token-12345"

func TestNewStore(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()
	if store == nil {
		t.Fatal("NewStore returned nil")
	}
}

func TestStoreSetAndGet(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config := embed.SiteConfig{
		Name:        "Test Site",
		Description: "Test Description",
		User:        "testuser",
	}

	store.Set("client1", config)

	retrieved, found := store.Get("client1")
	if !found {
		t.Fatal("expected to find config for client1")
	}

	if retrieved.Name != "Test Site" {
		t.Errorf("expected name to be 'Test Site', got %s", retrieved.Name)
	}

	_, found = store.Get("nonexistent")
	if found {
		t.Error("expected to not find config for nonexistent client")
	}
}

func TestStoreGetAll(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config1 := embed.SiteConfig{
		Name: "Site 1",
		User: "user1",
	}

	config2 := embed.SiteConfig{
		Name: "Site 2",
		User: "user2",
	}

	store.Set("client1", config1)
	store.Set("client2", config2)

	all := store.GetAll()

	if len(all) != 2 {
		t.Errorf("expected 2 configs, got %d", len(all))
	}

	if all["client1"].Name != "Site 1" {
		t.Errorf("expected client1 name to be 'Site 1', got %s", all["client1"].Name)
	}

	if all["client2"].Name != "Site 2" {
		t.Errorf("expected client2 name to be 'Site 2', got %s", all["client2"].Name)
	}
}

func TestHandlerWebSocket(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	server := httptest.NewServer(wsserver.Handler(store, testToken))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn := dialWebSocket(t, wsURL, testToken)
	defer conn.Close()

	config := embed.SiteConfig{
		Name:        "Test Site",
		Description: "Test Description",
		User:        "testuser",
	}

	msg := wsserver.Message{
		Type:     "push",
		ClientID: "test-client",
		Config:   &config,
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	var response wsserver.Message

	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if response.Type != "ack" {
		t.Errorf("expected type 'ack', got %s", response.Type)
	}

	if response.ClientID != "test-client" {
		t.Errorf("expected clientID 'test-client', got %s", response.ClientID)
	}

	retrieved, found := store.Get("test-client")
	if !found {
		t.Fatal("expected to find config in store")
	}

	if retrieved.Name != "Test Site" {
		t.Errorf("expected name 'Test Site', got %s", retrieved.Name)
	}
}

func dialWebSocket(t *testing.T, wsURL, token string) *websocket.Conn {
	t.Helper()

	dialer := websocket.DefaultDialer

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)

	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	return conn
}

func TestHandlerPing(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	server := httptest.NewServer(wsserver.Handler(store, testToken))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.DefaultDialer

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+testToken)

	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	defer conn.Close()

	msg := wsserver.Message{
		Type:     "ping",
		ClientID: "test-client",
	}

	err = conn.WriteJSON(msg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	var response wsserver.Message

	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if response.Type != "pong" {
		t.Errorf("expected type 'pong', got %s", response.Type)
	}
}

func TestHandlerUnknownMessageType(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	server := httptest.NewServer(wsserver.Handler(store, testToken))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.DefaultDialer

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+testToken)

	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	defer conn.Close()

	msg := wsserver.Message{
		Type:     "unknown",
		ClientID: "test-client",
	}

	err = conn.WriteJSON(msg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	var response wsserver.Message

	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if response.Type != "error" {
		t.Errorf("expected type 'error', got %s", response.Type)
	}

	if response.Error != "unknown message type" {
		t.Errorf("expected error 'unknown message type', got %s", response.Error)
	}
}

func TestStatusHandler(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config1 := embed.SiteConfig{
		Name: "Site 1",
		User: "user1",
	}

	config2 := embed.SiteConfig{
		Name: "Site 2",
		User: "user2",
	}

	store.Set("client1", config1)
	store.Set("client2", config2)

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	recorder := httptest.NewRecorder()

	wsserver.StatusHandler(store)(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var result map[string]embed.SiteConfig

	err := json.NewDecoder(recorder.Body).Decode(&result)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 configs, got %d", len(result))
	}

	if result["client1"].Name != "Site 1" {
		t.Errorf("expected client1 name 'Site 1', got %s", result["client1"].Name)
	}
}

func TestGetStore(t *testing.T) {
	t.Parallel()

	store := wsserver.GetStore()
	if store == nil {
		t.Fatal("GetStore returned nil")
	}
}

func TestHandleMessagePush(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	server := httptest.NewServer(wsserver.Handler(store, testToken))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.DefaultDialer

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+testToken)

	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	defer conn.Close()

	config := embed.SiteConfig{
		Name: "Test",
		User: "test",
	}

	msg := wsserver.Message{
		Type:     "push",
		ClientID: "test",
		Config:   &config,
	}

	err = conn.WriteJSON(msg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	var response wsserver.Message

	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if response.Type != "ack" {
		t.Errorf("expected ack, got %s", response.Type)
	}
}

func TestStoreSetWithoutSlackConfig(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config := embed.SiteConfig{
		Name:        "Test",
		Description: "Test config without Slack",
		User:        "testuser",
		Contributor: embed.Contributor{
			Active: true,
			Focus:  "Testing",
			Queue:  []string{"task1"},
		},
	}

	store.Set("test-client", config)

	retrieved, ok := store.Get("test-client")
	if !ok {
		t.Fatal("expected to retrieve stored config")
	}

	if retrieved.Name != "Test" {
		t.Errorf("expected name 'Test', got %s", retrieved.Name)
	}

	if retrieved.Slack.Enabled {
		t.Error("expected Slack to be disabled by default")
	}
}

func TestStoreSetWithSlackDisabled(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config := embed.SiteConfig{
		Name:        "Test",
		Description: "Test config with Slack disabled",
		User:        "testuser",
		Contributor: embed.Contributor{
			Active: true,
			Focus:  "Testing",
			Queue:  []string{"task1"},
		},
		Slack: embed.SlackConfig{
			Enabled:   false,
			UserToken: "xoxp-test",
		},
	}

	store.Set("test-client", config)

	retrieved, ok := store.Get("test-client")
	if !ok {
		t.Fatal("expected to retrieve stored config")
	}

	if retrieved.Slack.Enabled {
		t.Error("expected Slack to remain disabled")
	}
}

func TestStoreSetWithSlackEnabledInvalidToken(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config := embed.SiteConfig{
		Name:        "Test",
		Description: "Test config with Slack enabled but invalid token",
		User:        "testuser",
		Contributor: embed.Contributor{
			Active: true,
			Focus:  "Testing feature",
			Queue:  []string{"task1"},
		},
		Slack: embed.SlackConfig{
			Enabled:             true,
			UserToken:           "xoxp-invalid-token",
			StatusEmojiActive:   ":large_green_circle:",
			StatusEmojiInactive: ":red_circle:",
		},
	}

	store.Set("test-client", config)

	retrieved, ok := store.Get("test-client")
	if !ok {
		t.Fatal("expected to retrieve stored config")
	}

	if !retrieved.Slack.Enabled {
		t.Error("expected Slack to be enabled")
	}
}

func TestStoreSetWithSlackEnabledActiveTrue(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config := embed.SiteConfig{
		Name:        "Test",
		Description: "Test with active status",
		User:        "testuser",
		Contributor: embed.Contributor{
			Active: true,
			Focus:  "Working on feature",
			Queue:  []string{"task1"},
		},
		Slack: embed.SlackConfig{
			Enabled:             true,
			UserToken:           "xoxp-test-token",
			StatusEmojiActive:   ":large_green_circle:",
			StatusEmojiInactive: ":red_circle:",
		},
	}

	store.Set("test-client", config)

	retrieved, ok := store.Get("test-client")
	if !ok {
		t.Fatal("expected to retrieve stored config")
	}

	if !retrieved.Contributor.Active {
		t.Error("expected active to be true")
	}
}

func TestStoreSetWithSlackEnabledActiveFalse(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config := embed.SiteConfig{
		Name:        "Test",
		Description: "Test with inactive status",
		User:        "testuser",
		Contributor: embed.Contributor{
			Active: false,
			Focus:  "In meeting",
			Queue:  []string{"task1"},
		},
		Slack: embed.SlackConfig{
			Enabled:             true,
			UserToken:           "xoxp-test-token",
			StatusEmojiActive:   ":large_green_circle:",
			StatusEmojiInactive: ":red_circle:",
		},
	}

	store.Set("test-client", config)

	retrieved, ok := store.Get("test-client")
	if !ok {
		t.Fatal("expected to retrieve stored config")
	}

	if retrieved.Contributor.Active {
		t.Error("expected active to be false")
	}
}

func TestStoreSetWithSlackEnabledActiveFalseEmptyFocus(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config := embed.SiteConfig{
		Name:        "Test",
		Description: "Test with inactive status and empty focus",
		User:        "testuser",
		Contributor: embed.Contributor{
			Active: false,
			Focus:  "",
			Queue:  []string{"task1"},
		},
		Slack: embed.SlackConfig{
			Enabled:   true,
			UserToken: "xoxp-test-token",
		},
	}

	store.Set("test-client", config)

	retrieved, ok := store.Get("test-client")
	if !ok {
		t.Fatal("expected to retrieve stored config")
	}

	if retrieved.Contributor.Focus != "" {
		t.Error("expected focus to be empty")
	}
}

func TestStoreSetWithSlackEnabledDefaultEmojis(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	config := embed.SiteConfig{
		Name:        "Test",
		Description: "Test with default emojis",
		User:        "testuser",
		Contributor: embed.Contributor{
			Active: true,
			Focus:  "Working",
			Queue:  []string{"task1"},
		},
		Slack: embed.SlackConfig{
			Enabled:   true,
			UserToken: "xoxp-test-token",
		},
	}

	store.Set("test-client", config)

	retrieved, ok := store.Get("test-client")
	if !ok {
		t.Fatal("expected to retrieve stored config")
	}

	if retrieved.Slack.StatusEmojiActive != "" {
		t.Errorf("expected empty active emoji, got %s", retrieved.Slack.StatusEmojiActive)
	}
}

func TestHandlerMissingToken(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	server := httptest.NewServer(wsserver.Handler(store, testToken))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.DefaultDialer

	_, resp, err := dialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected connection to fail without token")
	}

	if resp != nil {
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}

		_ = resp.Body.Close()
	}
}

func TestHandlerInvalidToken(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()
	testToken := "correct-token-12345"

	server := httptest.NewServer(wsserver.Handler(store, testToken))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.DefaultDialer

	headers := http.Header{}
	headers.Add("Authorization", "Bearer wrong-token")

	_, resp, err := dialer.Dial(wsURL, headers)
	if err == nil {
		t.Fatal("expected connection to fail with wrong token")
	}

	if resp != nil {
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}

		_ = resp.Body.Close()
	}
}

func TestHandlerTokenViaQueryParam(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	server := httptest.NewServer(wsserver.Handler(store, testToken))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + testToken

	conn := dialWebSocket(t, wsURL, testToken)
	defer conn.Close()

	msg := wsserver.Message{
		Type:     "ping",
		ClientID: "test-client",
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	var response wsserver.Message

	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if response.Type != "pong" {
		t.Errorf("expected type 'pong', got %s", response.Type)
	}
}

func TestHandlerMalformedAuthHeader(t *testing.T) {
	t.Parallel()

	store := wsserver.NewStore()

	server := httptest.NewServer(wsserver.Handler(store, testToken))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	dialer := websocket.DefaultDialer

	headers := http.Header{}
	headers.Add("Authorization", testToken)

	_, resp, err := dialer.Dial(wsURL, headers)
	if err == nil {
		t.Fatal("expected connection to fail with malformed auth header")
	}

	if resp != nil {
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", resp.StatusCode)
		}

		_ = resp.Body.Close()
	}
}
