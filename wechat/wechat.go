package wechat

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
	// WeChat Work API endpoints
	tokenURL       = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	sendMessageURL = "https://qyapi.weixin.qq.com/cgi-bin/message/send"
)

// Config represents WeChat Work configuration
type Config struct {
	CorpID     string // Enterprise ID
	CorpSecret string // Application secret
	AgentID    int    // Application agent ID
	BaseURL    string // Optional: custom base URL
}

// Validate validates the config
func (c *Config) Validate() error {
	if c.CorpID == "" {
		return fmt.Errorf("CorpID is required")
	}
	if c.CorpSecret == "" {
		return fmt.Errorf("CorpSecret is required")
	}
	return nil
}

// GetPlatform returns the platform name
func (c *Config) GetPlatform() string {
	return "wechat"
}

// Client implements IMParrot interface for WeChat Work
type Client struct {
	config      *Config
	httpClient  *http.Client
	ownsHTTP    bool // Whether the client owns the http.Client and should close it
	token       string
	tokenMu     sync.RWMutex
	tokenExpiry time.Time
	closed      bool
	closedMu    sync.RWMutex
}

// NewClient creates a new WeChat Work client
func NewClient(config *Config, httpClient *http.Client) (*Client, error) {
	ownsHTTP := false
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
		ownsHTTP = true
	}

	client := &Client{
		config:     config,
		httpClient: httpClient,
		ownsHTTP:   ownsHTTP,
	}

	// Get initial access token
	if err := client.refreshToken(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	return client, nil
}

// GetPlatformName returns the platform name
func (c *Client) GetPlatformName() string {
	return "wechat"
}

// refreshToken gets a new access token
func (c *Client) refreshToken(ctx context.Context) error {
	url := fmt.Sprintf("%s?corpid=%s&corpsecret=%s",
		tokenURL, c.config.CorpID, c.config.CorpSecret)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tokenResp struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return err
	}

	if tokenResp.ErrCode != 0 {
		return fmt.Errorf("failed to get token: %s", tokenResp.ErrMsg)
	}

	c.tokenMu.Lock()
	c.token = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second) // Refresh 5 min early
	c.tokenMu.Unlock()

	return nil
}

// getToken returns a valid access token, refreshing if necessary
func (c *Client) getToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	if time.Now().Before(c.tokenExpiry) {
		token := c.token
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	if err := c.refreshToken(ctx); err != nil {
		return "", err
	}

	c.tokenMu.RLock()
	token := c.token
	c.tokenMu.RUnlock()
	return token, nil
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
	token, err := c.getToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Build message content based on type
	reqBody := map[string]interface{}{
		"agentid": c.config.AgentID,
	}

	// Set recipient based on chat type
	if target.ChatType == types.ChatTypePrivate {
		reqBody["touser"] = target.ID
	} else {
		reqBody["toparty"] = target.ID // or toparty for department
	}

	// Set message content
	switch msg.Type {
	case types.MessageTypeText:
		reqBody["msgtype"] = "text"
		reqBody["text"] = map[string]interface{}{
			"content": msg.Content,
		}
	case types.MessageTypeMarkdown:
		reqBody["msgtype"] = "markdown"
		reqBody["markdown"] = map[string]interface{}{
			"content": msg.Content,
		}
	default:
		reqBody["msgtype"] = "text"
		reqBody["text"] = map[string]interface{}{
			"content": msg.Content,
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?access_token=%s", sendMessageURL, token)
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
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return err
	}

	if apiResp.ErrCode != 0 {
		return fmt.Errorf("wechat API error: %s", apiResp.ErrMsg)
	}

	return nil
}

// SendPrivateMessage sends a private message to a user
func (c *Client) SendPrivateMessage(ctx context.Context, userID string, msg *types.Message) error {
	return c.SendMessage(ctx, msg, &types.SendOptions{
		Targets: []types.Target{{ID: userID, ChatType: types.ChatTypePrivate}},
	})
}

// SendGroupMessage sends a message to a department/group
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

	// Clear token
	c.tokenMu.Lock()
	c.token = ""
	c.tokenMu.Unlock()

	return nil
}
