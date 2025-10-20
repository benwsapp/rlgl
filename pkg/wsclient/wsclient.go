package wsclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/benwsapp/rlgl/pkg/embed"
	"github.com/gorilla/websocket"
)

const (
	handshakeTimeout = 10 * time.Second
	httpTimeout      = 10 * time.Second
)

var (
	ErrNotConnected          = errors.New("not connected")
	ErrServerError           = errors.New("server error")
	ErrUnexpectedMessageType = errors.New("unexpected message type")
	ErrUnexpectedStatusCode  = errors.New("unexpected status code")
)

type Message struct {
	Type     string            `json:"type"`
	ClientID string            `json:"clientId"`
	Config   *embed.SiteConfig `json:"config,omitempty"`
	Error    string            `json:"error,omitempty"`
}

type Client struct {
	serverURL string
	clientID  string
	authToken string
	conn      *websocket.Conn
}

func NewClient(serverURL, clientID, authToken string) *Client {
	return &Client{
		serverURL: serverURL,
		clientID:  clientID,
		authToken: authToken,
	}
}

func (c *Client) Connect() error {
	slog.Info("connecting to websocket server", "url", c.serverURL)

	dialer := &websocket.Dialer{
		HandshakeTimeout: handshakeTimeout,
	}

	headers := http.Header{}
	if c.authToken != "" {
		headers.Add("Authorization", "Bearer "+c.authToken)
	}

	conn, resp, err := dialer.Dial(c.serverURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	c.conn = conn
	slog.Info("connected to server", "url", c.serverURL)

	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	return nil
}

func (c *Client) PushConfig(config embed.SiteConfig) error {
	if c.conn == nil {
		return ErrNotConnected
	}

	msg := Message{
		Type:     "push",
		ClientID: c.clientID,
		Config:   &config,
	}

	err := c.conn.WriteJSON(msg)
	if err != nil {
		return fmt.Errorf("failed to send config: %w", err)
	}

	slog.Info("sent config to server", "client_id", c.clientID)

	var response Message

	err = c.conn.ReadJSON(&response)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if response.Type == "error" {
		return fmt.Errorf("%w: %s", ErrServerError, response.Error)
	}

	if response.Type != "ack" {
		return fmt.Errorf("%w: %s", ErrUnexpectedMessageType, response.Type)
	}

	slog.Info("received acknowledgment from server")

	return nil
}

func (c *Client) Ping() error {
	if c.conn == nil {
		return ErrNotConnected
	}

	msg := Message{
		Type:     "ping",
		ClientID: c.clientID,
	}

	err := c.conn.WriteJSON(msg)
	if err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	var response Message

	err = c.conn.ReadJSON(&response)
	if err != nil {
		return fmt.Errorf("failed to read pong: %w", err)
	}

	if response.Type != "pong" {
		return fmt.Errorf("%w: %s", ErrUnexpectedMessageType, response.Type)
	}

	return nil
}

func Run(serverURL, configPath, clientID, authToken string, interval time.Duration) error {
	client := NewClient(serverURL, clientID, authToken)

	err := client.Connect()
	if err != nil {
		return err
	}
	defer client.Close()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	pushConfig := func() error {
		config, err := embed.LoadSiteConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		pushErr := client.PushConfig(config)
		if pushErr != nil {
			return fmt.Errorf("failed to push config: %w", pushErr)
		}

		return nil
	}

	err = pushConfig()
	if err != nil {
		return err
	}

	slog.Info("client started", "interval", interval, "config", configPath)

	for range ticker.C {
		err := pushConfig()
		if err != nil {
			slog.Error("failed to push config", "error", err)

			continue
		}
	}

	return nil
}

func RunOnce(serverURL, configPath, clientID, authToken string) error {
	client := NewClient(serverURL, clientID, authToken)

	err := client.Connect()
	if err != nil {
		return err
	}
	defer client.Close()

	config, configErr := embed.LoadSiteConfig(configPath)
	if configErr != nil {
		return fmt.Errorf("failed to load config: %w", configErr)
	}

	pushErr := client.PushConfig(config)
	if pushErr != nil {
		return fmt.Errorf("failed to push config: %w", pushErr)
	}

	slog.Info("config pushed successfully")

	return nil
}

func GetStatusJSON(serverURL string) (string, error) {
	client := &http.Client{
		Timeout: httpTimeout,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, serverURL+"/status", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode)
	}

	var result map[string]embed.SiteConfig

	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return "", fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}
