package wsclient_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/benwsapp/rlgl/pkg/embed"
	"github.com/benwsapp/rlgl/pkg/wsclient"
	"github.com/gorilla/websocket"
)

func createTestConfig(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	content := `name: "Test Site"
description: "Test Description"
user: "testuser"
contributor:
  active: true
  focus: "Testing"
  queue:
    - "Task 1"
    - "Task 2"
`

	err := os.WriteFile(configPath, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	return configPath
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	client := wsclient.NewClient("ws://localhost:8080/ws", "test-client")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestClientConnectAndClose(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := wsclient.NewClient(wsURL, "test-client")

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestClientPushConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var msg wsclient.Message

		err = conn.ReadJSON(&msg)
		if err != nil {
			t.Fatal(err)
		}

		if msg.Type != "push" {
			t.Errorf("expected type push, got %s", msg.Type)
		}

		response := wsclient.Message{
			Type:     "ack",
			ClientID: msg.ClientID,
		}

		err = conn.WriteJSON(response)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := wsclient.NewClient(wsURL, "test-client")

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	config := embed.SiteConfig{
		Name:        "Test Site",
		Description: "Test Description",
		User:        "testuser",
	}

	err = client.PushConfig(config)
	if err != nil {
		t.Fatalf("PushConfig failed: %v", err)
	}
}

func TestClientPushConfigNotConnected(t *testing.T) {
	t.Parallel()

	client := wsclient.NewClient("ws://localhost:8080/ws", "test-client")

	config := embed.SiteConfig{
		Name: "Test Site",
	}

	err := client.PushConfig(config)
	if !errors.Is(err, wsclient.ErrNotConnected) {
		t.Errorf("expected ErrNotConnected, got %v", err)
	}
}

func TestClientPing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var msg wsclient.Message

		err = conn.ReadJSON(&msg)
		if err != nil {
			t.Fatal(err)
		}

		if msg.Type != "ping" {
			t.Errorf("expected type ping, got %s", msg.Type)
		}

		response := wsclient.Message{
			Type:     "pong",
			ClientID: msg.ClientID,
		}

		err = conn.WriteJSON(response)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := wsclient.NewClient(wsURL, "test-client")

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	err = client.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestGetStatusJSON(t *testing.T) {
	t.Parallel()

	testConfigs := map[string]embed.SiteConfig{
		"client1": {
			Name:        "Site 1",
			Description: "Description 1",
			User:        "user1",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/status" {
			http.NotFound(writer, req)

			return
		}

		writer.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(writer).Encode(testConfigs)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	result, err := wsclient.GetStatusJSON(server.URL)
	if err != nil {
		t.Fatalf("GetStatusJSON failed: %v", err)
	}

	if !strings.Contains(result, "Site 1") {
		t.Errorf("expected result to contain 'Site 1', got %s", result)
	}
}

func TestClientCloseNilConnection(t *testing.T) {
	t.Parallel()

	client := wsclient.NewClient("ws://localhost:8080/ws", "test-client")

	err := client.Close()
	if err != nil {
		t.Fatalf("Close with nil connection should not error, got: %v", err)
	}
}

func TestRunOnce(t *testing.T) {
	t.Parallel()

	configPath := createTestConfig(t)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var msg wsclient.Message

		err = conn.ReadJSON(&msg)
		if err != nil {
			t.Fatal(err)
		}

		response := wsclient.Message{
			Type:     "ack",
			ClientID: msg.ClientID,
		}

		err = conn.WriteJSON(response)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	err := wsclient.RunOnce(wsURL, configPath, "test-client")
	if err != nil {
		t.Fatalf("RunOnce failed: %v", err)
	}
}

func TestRunOnceInvalidConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	err := wsclient.RunOnce(wsURL, "/nonexistent/config.yaml", "test-client")
	if err == nil {
		t.Fatal("expected error for invalid config path")
	}
}

func TestRunOnceConnectError(t *testing.T) {
	t.Parallel()

	configPath := createTestConfig(t)

	err := wsclient.RunOnce("ws://localhost:9999/ws", configPath, "test-client")
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestClientPushConfigServerError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var msg wsclient.Message

		err = conn.ReadJSON(&msg)
		if err != nil {
			t.Fatal(err)
		}

		response := wsclient.Message{
			Type:     "error",
			ClientID: msg.ClientID,
			Error:    "test error",
		}

		err = conn.WriteJSON(response)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := wsclient.NewClient(wsURL, "test-client")

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	config := embed.SiteConfig{
		Name: "Test Site",
		User: "testuser",
	}

	err = client.PushConfig(config)
	if err == nil {
		t.Fatal("expected error response from server")
	}

	if !errors.Is(err, wsclient.ErrServerError) {
		t.Errorf("expected ErrServerError, got %v", err)
	}
}

func TestClientPushConfigUnexpectedMessageType(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var msg wsclient.Message

		err = conn.ReadJSON(&msg)
		if err != nil {
			t.Fatal(err)
		}

		response := wsclient.Message{
			Type:     "unknown",
			ClientID: msg.ClientID,
		}

		err = conn.WriteJSON(response)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := wsclient.NewClient(wsURL, "test-client")

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	config := embed.SiteConfig{
		Name: "Test Site",
		User: "testuser",
	}

	err = client.PushConfig(config)
	if err == nil {
		t.Fatal("expected error for unexpected message type")
	}

	if !errors.Is(err, wsclient.ErrUnexpectedMessageType) {
		t.Errorf("expected ErrUnexpectedMessageType, got %v", err)
	}
}

func TestClientPingNotConnected(t *testing.T) {
	t.Parallel()

	client := wsclient.NewClient("ws://localhost:8080/ws", "test-client")

	err := client.Ping()
	if !errors.Is(err, wsclient.ErrNotConnected) {
		t.Errorf("expected ErrNotConnected, got %v", err)
	}
}

func TestClientPingUnexpectedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var msg wsclient.Message

		err = conn.ReadJSON(&msg)
		if err != nil {
			t.Fatal(err)
		}

		response := wsclient.Message{
			Type:     "unexpected",
			ClientID: msg.ClientID,
		}

		err = conn.WriteJSON(response)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	client := wsclient.NewClient(wsURL, "test-client")

	err := client.Connect()
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	err = client.Ping()
	if err == nil {
		t.Fatal("expected error for unexpected response")
	}

	if !errors.Is(err, wsclient.ErrUnexpectedMessageType) {
		t.Errorf("expected ErrUnexpectedMessageType, got %v", err)
	}
}

func TestGetStatusJSONHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := wsclient.GetStatusJSON(server.URL)
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}

	if !errors.Is(err, wsclient.ErrUnexpectedStatusCode) {
		t.Errorf("expected ErrUnexpectedStatusCode, got %v", err)
	}
}

func TestRunConnectError(t *testing.T) {
	t.Parallel()

	configPath := createTestConfig(t)

	err := wsclient.Run("ws://localhost:9999/ws", configPath, "test-client", 1*time.Second)
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestRunInitialPushError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(writer, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	err := wsclient.Run(wsURL, "/nonexistent/config.yaml", "test-client", 1*time.Second)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}
