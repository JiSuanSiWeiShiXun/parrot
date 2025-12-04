package dingtalk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/JiSuanSiWeiShiXun/parrot/types"
)

// Config represents DingTalk robot configuration
type Config struct {
	AccessToken string // Robot webhook access token
	Secret      string // Optional: secret for signature
	BaseURL     string // Optional: custom webhook URL
}

// Validate validates the config
func (c *Config) Validate() error {
	if c.AccessToken == "" {
		return fmt.Errorf("AccessToken is required")
	}
	return nil
}

// GetPlatform returns the platform name
func (c *Config) GetPlatform() string {
	return "dingtalk"
}

// Client implements IMParrot interface for DingTalk
type Client struct {
	config     *Config
	httpClient *http.Client
	ownsHTTP   bool // Whether the client owns the http.Client and should close it
	webhookURL string
	closed     bool
	closedMu   sync.RWMutex
}

// NewClient creates a new DingTalk robot client
func NewClient(config *Config, httpClient *http.Client) (*Client, error) {
	ownsHTTP := false
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
		ownsHTTP = true
	}

	webhookURL := config.BaseURL
	if webhookURL == "" {
		webhookURL = "https://oapi.dingtalk.com/robot/send"
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		ownsHTTP:   ownsHTTP,
		webhookURL: webhookURL,
	}, nil
}

// GetPlatformName returns the platform name
func (c *Client) GetPlatformName() string {
	return "dingtalk"
}

// sign generates signature for DingTalk webhook
func (c *Client) sign(timestamp int64) string {
	if c.config.Secret == "" {
		return ""
	}

	stringToSign := fmt.Sprintf("%d\n%s", timestamp, c.config.Secret)
	h := hmac.New(sha256.New, []byte(c.config.Secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// SendMessage sends a message with options (Strategy pattern implementation)
func (c *Client) SendMessage(ctx context.Context, msg *types.Message, opts *types.SendOptions) error {
	if msg == nil || opts == nil {
		return fmt.Errorf("message and options cannot be nil")
	}

	// DingTalk webhook doesn't support multiple targets, but we still check
	// Note: For DingTalk, all messages go to the same webhook, so no retry needed for multiple targets

	// Build webhook URL with signature
	timestamp := time.Now().UnixMilli()
	webhookURL := fmt.Sprintf("%s?access_token=%s", c.webhookURL, c.config.AccessToken)

	if c.config.Secret != "" {
		sign := c.sign(timestamp)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s",
			webhookURL, timestamp, url.QueryEscape(sign))
	}

	// Build request body based on message type
	var reqBody map[string]interface{}

	switch msg.Type {
	case types.MessageTypeText:
		reqBody = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]interface{}{
				"content": msg.Content,
			},
		}

		// Add @ mentions for group messages
		if len(opts.AtUsers) > 0 {
			reqBody["at"] = map[string]interface{}{
				"atMobiles": opts.AtUsers,
				"isAtAll":   false,
			}
		}

	case types.MessageTypeMarkdown:
		reqBody = map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]interface{}{
				"title": "Message",
				"text":  msg.Content,
			},
		}

		if len(opts.AtUsers) > 0 {
			reqBody["at"] = map[string]interface{}{
				"atMobiles": opts.AtUsers,
				"isAtAll":   false,
			}
		}

	default:
		reqBody = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]interface{}{
				"content": msg.Content,
			},
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(body))
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
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return err
	}

	if apiResp.ErrCode != 0 {
		return fmt.Errorf("dingtalk API error: %s", apiResp.ErrMsg)
	}

	return nil
}

// SendPrivateMessage sends a private message (DingTalk robot doesn't support private messages directly)
func (c *Client) SendPrivateMessage(ctx context.Context, userID string, msg *types.Message) error {
	return fmt.Errorf("dingtalk robot does not support private messages")
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
