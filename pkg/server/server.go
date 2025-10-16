package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/benwsapp/rlgl/pkg/embed"
	"github.com/benwsapp/rlgl/pkg/wsserver"
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

func Run(addr string, store *wsserver.Store, trustedOrigins []string) error {
	mux := http.NewServeMux()

	// WebSocket endpoints for client push
	mux.HandleFunc("/ws", wsserver.Handler(store))

	// Status endpoint showing all stored configs
	mux.HandleFunc("/status", wsserver.StatusHandler(store))

	// HTML index page (uses first available config or shows all)
	mux.HandleFunc("/", IndexHandlerWithStore(store))

	// JSON config endpoints
	mux.HandleFunc("/config", ConfigHandlerWithStore(store))

	// SSE events endpoint
	mux.HandleFunc("/events", EventsHandlerWithStore(store))

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

func IndexHandlerWithStore(store *wsserver.Store) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, req *http.Request) {
		slog.Info("request received", "method", req.Method, "path", req.URL.Path, "remote_addr", req.RemoteAddr)

		configs := store.GetAll()
		if len(configs) == 0 {
			http.Error(responseWriter, "no configs available", http.StatusNotFound)

			return
		}

		// Use the first config for rendering
		var cfg embed.SiteConfig
		for _, c := range configs {
			cfg = c

			break
		}

		content, err := renderIndex(cfg)
		if err != nil {
			slog.Error("failed rendering template", "error", err)
			http.Error(responseWriter, "internal server error", http.StatusInternalServerError)

			return
		}

		_, writeErr := responseWriter.Write(content)
		if writeErr != nil {
			slog.Error("failed writing response", "error", writeErr)
		}
	}
}

func ConfigHandlerWithStore(store *wsserver.Store) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, _ *http.Request) {
		configs := store.GetAll()
		if len(configs) == 0 {
			http.Error(responseWriter, "no configs available", http.StatusNotFound)

			return
		}

		// Return the first config
		var cfg embed.SiteConfig
		for _, c := range configs {
			cfg = c

			break
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.Header().Set("Cache-Control", "no-store")

		err := json.NewEncoder(responseWriter).Encode(cfg)
		if err != nil {
			slog.Error("failed encoding config", "error", err)
		}
	}
}

func EventsHandlerWithStore(store *wsserver.Store) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, req *http.Request) {
		flusher, ok := responseWriter.(http.Flusher)
		if !ok {
			http.Error(responseWriter, "streaming unsupported", http.StatusInternalServerError)

			return
		}

		SetupSSEHeaders(responseWriter)
		StreamEventsFromStore(req.Context(), responseWriter, flusher, store)
	}
}

func StreamEventsFromStore(
	ctx context.Context,
	responseWriter http.ResponseWriter,
	flusher http.Flusher,
	store *wsserver.Store,
) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := SendEventDataFromStore(responseWriter, flusher, store)
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func SendEventDataFromStore(responseWriter http.ResponseWriter, flusher http.Flusher, store *wsserver.Store) error {
	configs := store.GetAll()
	if len(configs) == 0 {
		return nil
	}

	// Send the first config
	var cfg embed.SiteConfig
	for _, c := range configs {
		cfg = c

		break
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

func renderIndex(cfg embed.SiteConfig) ([]byte, error) {
	tmpl, err := embed.GetTemplate()
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	var buf []byte

	w := &bytesWriter{buf: &buf}

	err = tmpl.Execute(w, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf, nil
}

type bytesWriter struct {
	buf *[]byte
}

func (bw *bytesWriter) Write(p []byte) (int, error) {
	*bw.buf = append(*bw.buf, p...)

	return len(p), nil
}
