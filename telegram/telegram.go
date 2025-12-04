package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/JiSuanSiWeiShiXun/parrot/types"
)

const (
	// Telegram Bot API base URL
	telegramAPIBase = "https://api.telegram.org/bot"
)

// Config represents Telegram bot configuration
type Config struct {
	BotToken string
	BaseURL  string // Optional: custom base URL (for proxy or test)
}

// Validate validates the config
func (c *Config) Validate() error {
	if c.BotToken == "" {
		return fmt.Errorf("BotToken is required")
	}
	return nil
}

// GetPlatform returns the platform name
func (c *Config) GetPlatform() string {
	return "telegram"
}

// Client implements IMParrot interface for Telegram
type Client struct {
	config     *Config
	httpClient *http.Client
	ownsHTTP   bool // Whether the client owns the http.Client and should close it
	apiURL     string
	closed     bool
	closedMu   sync.RWMutex
}

// NewClient creates a new Telegram bot client
func NewClient(config *Config, httpClient *http.Client) (*Client, error) {
	ownsHTTP := false
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
		ownsHTTP = true
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = telegramAPIBase
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		ownsHTTP:   ownsHTTP,
		apiURL:     baseURL + config.BotToken,
	}, nil
}

// GetPlatformName returns the platform name
func (c *Client) GetPlatformName() string {
	return "telegram"
}

// SendMessage sends a message with options (Strategy pattern implementation)
func (c *Client) SendMessage(ctx context.Context, msg *types.Message, opts *types.SendOptions) error {
	if msg == nil || opts == nil {
		return fmt.Errorf("message and options cannot be nil")
	}

	if len(opts.Targets) == 0 {
		return fmt.Errorf("at least one target is required")
	}

	// Send to multiple targets with retry
	const maxRetries = 3
	failedTargets := make([]types.FailedTarget, 0)
	successCount := 0

	for _, target := range opts.Targets {
		var lastErr error
		sent := false

		// Retry up to maxRetries times for each target
		for retry := 0; retry < maxRetries; retry++ {
			if err := c.sendToSingleTarget(ctx, msg, target); err != nil {
				lastErr = err
				// Wait a bit before retrying (exponential backoff)
				if retry < maxRetries-1 {
					time.Sleep(time.Duration(100*(retry+1)) * time.Millisecond)
				}
			} else {
				sent = true
				successCount++
				break
			}
		}

		// Record failed target after all retries exhausted
		if !sent {
			failedTargets = append(failedTargets, types.FailedTarget{
				Target: target,
				Error:  lastErr,
			})
		}
	}

	// Return error with failed targets information
	if len(failedTargets) > 0 {
		return &types.SendError{
			FailedTargets: failedTargets,
			SuccessCount:  successCount,
			TotalCount:    len(opts.Targets),
		}
	}

	return nil
}

// sendToSingleTarget sends a message to a single target
func (c *Client) sendToSingleTarget(ctx context.Context, msg *types.Message, target types.Target) error {
	// Build request body
	reqBody := map[string]interface{}{
		"chat_id": target.ID,
	}

	// Set message content based on type
	switch msg.Type {
	case types.MessageTypeText:
		reqBody["text"] = msg.Content
	case types.MessageTypeMarkdown:
		reqBody["text"] = msg.Content
		reqBody["parse_mode"] = "MarkdownV2"
	default:
		reqBody["text"] = msg.Content
	}

	// Add extra options from msg.Data
	if msg.Data != nil {
		for k, v := range msg.Data {
			reqBody[k] = v
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/sendMessage", c.apiURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return err
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram API error: %s", apiResp.Description)
	}

	return nil
}

// SendPrivateMessage sends a private message to a user
func (c *Client) SendPrivateMessage(ctx context.Context, userID string, msg *types.Message) error {
	return c.SendMessage(ctx, msg, &types.SendOptions{
		Targets: []types.Target{{ID: userID, ChatType: types.ChatTypePrivate}},
	})
}

// SendGroupMessage sends a message to a group
func (c *Client) SendGroupMessage(ctx context.Context, groupID string, msg *types.Message) error {
	return c.SendMessage(ctx, msg, &types.SendOptions{
		Targets: []types.Target{{ID: groupID, ChatType: types.ChatTypeGroup}},
	})
}

// Close releases all resources held by the client
func (c *Client) Close() error {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true

	// Close HTTP client connections if we own it
	if c.ownsHTTP && c.httpClient != nil {
		c.httpClient.CloseIdleConnections()
	}

	return nil
}
