package wsserver

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/benwsapp/rlgl/pkg/embed"
	"github.com/benwsapp/rlgl/pkg/slack"
	"github.com/gorilla/websocket"
)

const (
	readBufferSize  = 1024
	writeBufferSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

type Store struct {
	mu      sync.RWMutex
	configs map[string]embed.SiteConfig
}

func NewStore() *Store {
	return &Store{
		configs: make(map[string]embed.SiteConfig),
	}
}

func (s *Store) Set(clientID string, config embed.SiteConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs[clientID] = config
	slog.Info("stored config", "client_id", clientID, "name", config.Name)

	if config.Slack.Enabled && config.Slack.UserToken != "" {
		go syncToSlack(config)
	}
}

func syncToSlack(config embed.SiteConfig) {
	client := slack.NewClient(config.Slack.UserToken)

	var statusText string

	var statusEmoji string

	if config.Contributor.Active {
		statusEmoji = config.Slack.StatusEmojiActive
		if statusEmoji == "" {
			statusEmoji = ":large_green_circle:"
		}

		statusText = config.Contributor.Focus
	} else {
		statusEmoji = config.Slack.StatusEmojiInactive
		if statusEmoji == "" {
			statusEmoji = ":red_circle:"
		}

		statusText = config.Contributor.Focus
		if statusText == "" {
			statusText = "Busy"
		}
	}

	ttl := config.Slack.TTLSeconds
	if ttl == 0 {
		ttl = 3600
	}

	expirationSeconds := int(time.Now().Add(time.Duration(ttl) * time.Second).Unix())

	err := client.SetStatus(statusText, statusEmoji, expirationSeconds)
	if err != nil {
		slog.Error("failed to sync status to Slack", "error", err, "user", config.User)
	} else {
		slog.Info("synced status to Slack", "user", config.User, "status", statusText, "emoji", statusEmoji)
	}
}

func (s *Store) Get(clientID string) (embed.SiteConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, ok := s.configs[clientID]

	return config, ok
}

func (s *Store) GetAll() map[string]embed.SiteConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]embed.SiteConfig, len(s.configs))
	maps.Copy(result, s.configs)

	return result
}

type Message struct {
	Type     string            `json:"type"`
	ClientID string            `json:"clientId"`
	Config   *embed.SiteConfig `json:"config,omitempty"`
	Error    string            `json:"error,omitempty"`
}

func Handler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			slog.Error("failed to upgrade connection", "error", err)

			return
		}
		defer conn.Close()

		slog.Info("websocket connection established", "remote_addr", req.RemoteAddr)

		handleConnection(conn, store)
	}
}

func handleConnection(conn *websocket.Conn, store *Store) {
	for {
		var msg Message

		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err)
			}

			break
		}

		slog.Info("received message", "type", msg.Type, "client_id", msg.ClientID)

		handleErr := handleMessage(conn, store, msg)
		if handleErr != nil {
			return
		}
	}
}

func handleMessage(conn *websocket.Conn, store *Store, msg Message) error {
	switch msg.Type {
	case "push":
		store.Set(msg.ClientID, *msg.Config)

		response := Message{
			Type:     "ack",
			ClientID: msg.ClientID,
		}

		writeErr := conn.WriteJSON(response)
		if writeErr != nil {
			slog.Error("failed to send ack", "error", writeErr)

			return fmt.Errorf("failed to send ack: %w", writeErr)
		}

	case "ping":
		response := Message{
			Type:     "pong",
			ClientID: msg.ClientID,
		}

		writeErr := conn.WriteJSON(response)
		if writeErr != nil {
			slog.Error("failed to send pong", "error", writeErr)

			return fmt.Errorf("failed to send pong: %w", writeErr)
		}

	default:
		response := Message{
			Type:     "error",
			ClientID: msg.ClientID,
			Error:    "unknown message type",
		}

		writeErr := conn.WriteJSON(response)
		if writeErr != nil {
			slog.Error("failed to send error", "error", writeErr)

			return fmt.Errorf("failed to send error: %w", writeErr)
		}
	}

	return nil
}

func StatusHandler(store *Store) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		configs := store.GetAll()

		writer.Header().Set("Content-Type", "application/json")

		encodeErr := json.NewEncoder(writer).Encode(configs)
		if encodeErr != nil {
			slog.Error("failed to encode status", "error", encodeErr)
			http.Error(writer, "internal server error", http.StatusInternalServerError)
		}
	}
}

func GetStore() *Store {
	return NewStore()
}

func Run(addr string, store *Store, _ []string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", Handler(store))
	mux.HandleFunc("/status", StatusHandler(store))

	const (
		readHeaderTimeout = 5 * time.Second
		readTimeout       = 10 * time.Second
		writeTimeout      = 30 * time.Second
		idleTimeout       = 60 * time.Second
	)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	slog.Info("websocket server listening", "addr", addr)

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
