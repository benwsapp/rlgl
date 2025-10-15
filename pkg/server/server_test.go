package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/benwsapp/rlgl/pkg/server"
)

func TestIndexHandler(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")
	configContent := `name: "Test Site"
description: "Test description"
user: "testuser"
contributor:
  active: true
  focus: "testing"
  queue:
    - "task 1"
    - "task 2"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	handler := server.IndexHandler(configPath)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if rec.Body.Len() == 0 {
		t.Error("expected response body, got empty")
	}
}

func TestIndexHandlerInvalidConfig(t *testing.T) {
	t.Parallel()

	handler := server.IndexHandler("/nonexistent/path.yaml")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestConfigHandler(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")

	configContent := `name: "Test Site"
description: "Test description"
user: "testuser"
contributor:
  active: true
  focus: "testing"
  queue:
    - "task 1"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	handler := server.ConfigHandler(configPath)
	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	cacheControl := rec.Header().Get("Cache-Control")
	if cacheControl != "no-store" {
		t.Errorf("expected Cache-Control no-store, got %s", cacheControl)
	}
}

func TestConfigHandlerInvalidConfig(t *testing.T) {
	t.Parallel()

	handler := server.ConfigHandler("/nonexistent/path.yaml")
	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

type noFlusherWriter struct {
	header http.Header
	code   int
	body   []byte
}

func (w *noFlusherWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}

	return w.header
}

func (w *noFlusherWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)

	return len(b), nil
}

func (w *noFlusherWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func TestEventsHandlerNoFlusher(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")

	configContent := `name: "Test Site"
description: "Test"
user: "test"
contributor:
  active: true
  focus: "test"
  queue: []
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	handler := server.EventsHandler(configPath)
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	rec := &noFlusherWriter{}

	handler.ServeHTTP(rec, req)

	if rec.code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.code)
	}
}

func TestSetupSSEHeaders(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	server.SetupSSEHeaders(rec)

	tests := []struct {
		header   string
		expected string
	}{
		{"Content-Type", "text/event-stream"},
		{"Cache-Control", "no-cache, no-transform"},
		{"Connection", "keep-alive"},
	}

	for _, tt := range tests {
		if got := rec.Header().Get(tt.header); got != tt.expected {
			t.Errorf("header %s: expected %s, got %s", tt.header, tt.expected, got)
		}
	}
}

func TestStreamEvents(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")

	configContent := `name: "Test Site"
description: "Test"
user: "test"
contributor:
  active: true
  focus: "test"
  queue: []
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	rec := httptest.NewRecorder()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	done := make(chan bool, 1)

	go func() {
		server.StreamEvents(ctx, rec, rec, configPath)

		done <- true
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("streamEvents did not complete in time")
	}
}

func TestSendEventData(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "site.yaml")

	configContent := `name: "Test Site"
description: "Test"
user: "test"
contributor:
  active: true
  focus: "test"
  queue:
    - "task 1"
`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	rec := httptest.NewRecorder()

	err = server.SendEventData(rec, rec, configPath)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	body := rec.Body.String()
	if body == "" {
		t.Error("expected response body, got empty")
	}
}

func TestSendEventDataInvalidConfig(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	err := server.SendEventData(rec, rec, "/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}
