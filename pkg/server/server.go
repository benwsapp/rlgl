package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/benwsapp/rlgl/pkg/embed"
)

func IndexHandler(configPath string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, req *http.Request) {
		slog.Info("request received", "method", req.Method, "path", req.URL.Path, "remote_addr", req.RemoteAddr)

		content, err := embed.Index(configPath)
		if err != nil {
			slog.Error("failed rendering template", "error", err)
			http.Error(responseWriter, "internal server error", http.StatusInternalServerError)

			return
		}

		_, writeErr := responseWriter.Write(content)
		if writeErr != nil {
			slog.Error("failed writing response", "error", writeErr)

			return
		}
	}
}

func ConfigHandler(configPath string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, _ *http.Request) {
		cfg, err := embed.LoadSiteConfig(configPath)
		if err != nil {
			slog.Error("failed loading config", "error", err)
			http.Error(responseWriter, "internal server error", http.StatusInternalServerError)

			return
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.Header().Set("Cache-Control", "no-store")

		err = json.NewEncoder(responseWriter).Encode(cfg)
		if err != nil {
			slog.Error("failed encoding config", "error", err)
		}
	}
}

func EventsHandler(configPath string) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, req *http.Request) {
		flusher, ok := responseWriter.(http.Flusher)
		if !ok {
			http.Error(responseWriter, "streaming unsupported", http.StatusInternalServerError)

			return
		}

		SetupSSEHeaders(responseWriter)
		StreamEvents(req.Context(), responseWriter, flusher, configPath)
	}
}

// StreamEvents sends Server-Sent Events periodically until context is done.
func StreamEvents(ctx context.Context, responseWriter http.ResponseWriter, flusher http.Flusher, configPath string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := SendEventData(responseWriter, flusher, configPath)
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// SetupSSEHeaders sets HTTP headers for Server-Sent Events.
func SetupSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
}

// SendEventData sends a single Server-Sent Event with the current config.
func SendEventData(responseWriter http.ResponseWriter, flusher http.Flusher, configPath string) error {
	cfg, err := embed.LoadSiteConfig(configPath)
	if err != nil {
		slog.Error("failed loading config for event stream", "error", err)

		return fmt.Errorf("failed to load config: %w", err)
	}

	payload, err := json.Marshal(cfg)
	if err != nil {
		slog.Error("failed marshaling event payload", "error", err)

		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, writeErr := responseWriter.Write([]byte("data: " + string(payload) + "\n\n"))
	if writeErr != nil {
		slog.Error("failed writing event payload", "error", writeErr)

		return fmt.Errorf("failed to write event: %w", writeErr)
	}

	flusher.Flush()

	return nil
}

func Run(addr, configPath string, trustedOrigins []string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler(configPath))
	mux.HandleFunc("/config", ConfigHandler(configPath))
	mux.HandleFunc("/events", EventsHandler(configPath))

	const (
		readHeaderTimeout = 5 * time.Second
		readTimeout       = 10 * time.Second
		writeTimeout      = 30 * time.Second
		idleTimeout       = 60 * time.Second
	)

	server := &http.Server{
		Addr:              addr,
		Handler:           CSRFMiddleware(mux, trustedOrigins...),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	slog.Info("http server listening", "addr", addr)

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
