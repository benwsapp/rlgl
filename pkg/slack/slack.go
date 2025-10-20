package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	slackAPIURL     = "https://slack.com/api/users.profile.set"
	requestTimeout  = 10 * time.Second
	maxStatusLength = 100
)

var ErrSlackAPI = errors.New("slack API error")

type Client struct {
	userToken  string
	httpClient *http.Client
	apiURL     string
}

func NewClient(userToken string) *Client {
	return &Client{
		userToken: userToken,
		apiURL:    slackAPIURL,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (c *Client) WithAPIURL(url string) *Client {
	c.apiURL = url

	return c
}

type ProfileStatus struct {
	StatusText       string `json:"status_text"`       //nolint:tagliatelle // Slack API uses snake_case
	StatusEmoji      string `json:"status_emoji"`      //nolint:tagliatelle // Slack API uses snake_case
	StatusExpiration int    `json:"status_expiration"` //nolint:tagliatelle // Slack API uses snake_case
}

type ProfileRequest struct {
	Profile ProfileStatus `json:"profile"`
}

type ProfileResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func (c *Client) SetStatus(statusText, statusEmoji string, expirationSeconds int) error {
	if len(statusText) > maxStatusLength {
		statusText = statusText[:maxStatusLength]
	}

	req := ProfileRequest{
		Profile: ProfileStatus{
			StatusText:       statusText,
			StatusEmoji:      statusEmoji,
			StatusExpiration: expirationSeconds,
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Authorization", "Bearer "+c.userToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var slackResp ProfileResponse

	err = json.Unmarshal(body, &slackResp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !slackResp.Ok {
		return fmt.Errorf("%w: %s", ErrSlackAPI, slackResp.Error)
	}

	return nil
}

func (c *Client) ClearStatus() error {
	return c.SetStatus("", "", 0)
}
