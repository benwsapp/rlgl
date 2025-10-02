package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/benwsapp/rlgl/pkg/embed"
)

func Run(addr, configPath string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request received", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)

		content, err := embed.Index(configPath)
		if err != nil {
			slog.Error("failed rendering template", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(content); err != nil {
			slog.Error("failed writing response", "error", err)
			return
		}
	})

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		cfg, err := embed.LoadSiteConfig(configPath)
		if err != nil {
			slog.Error("failed loading config", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if err := json.NewEncoder(w).Encode(cfg); err != nil {
			slog.Error("failed encoding config", "error", err)
		}
	})

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("Connection", "keep-alive")

		notify := r.Context().Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cfg, err := embed.LoadSiteConfig(configPath)
				if err != nil {
					slog.Error("failed loading config for event stream", "error", err)
					return
				}

				payload, err := json.Marshal(cfg)
				if err != nil {
					slog.Error("failed marshaling event payload", "error", err)
					return
				}

				if _, err := w.Write([]byte("data: " + string(payload) + "\n\n")); err != nil {
					slog.Error("failed writing event payload", "error", err)
					return
				}
				flusher.Flush()
			case <-notify:
				return
			}
		}
	})

	slog.Info("http server listening", "addr", addr)
	return http.ListenAndServe(addr, mux)
}
