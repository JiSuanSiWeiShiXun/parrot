package lark

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
	// Lark API endpoints
	tokenURL       = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	sendMessageURL = "https://open.feishu.cn/open-apis/im/v1/messages"
	batchGetIDURL  = "https://open.feishu.cn/open-apis/contact/v3/users/batch_get_id"
)

// Config represents Lark/Feishu configuration
type Config struct {
	AppID      string
	AppSecret  string
	BaseURL    string // Optional: custom base URL
	WebhookURL string // Optional: webhook URL for group robot
}

// Validate validates the config
func (c *Config) Validate() error {
	// Webhook mode: only webhook URL is required
	if c.WebhookURL != "" {
		return nil
	}
	// App mode: AppID and AppSecret are required
	if c.AppID == "" {
		return fmt.Errorf("AppID is required (or provide WebhookURL for webhook mode)")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("AppSecret is required (or provide WebhookURL for webhook mode)")
	}
	return nil
}

// GetPlatform returns the platform name
func (c *Config) GetPlatform() string {
	return "lark"
}

// Client implements IMParrot interface for Lark/Feishu
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

// NewClient creates a new Lark client
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

	// Get initial access token only if not in webhook mode
	if config.WebhookURL == "" {
		if err := client.refreshToken(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to get access token: %w", err)
		}
	}

	return client, nil
}

// GetPlatformName returns the platform name
func (c *Client) GetPlatformName() string {
	return "lark"
}

// refreshToken gets a new tenant access token
func (c *Client) refreshToken(ctx context.Context) error {
	reqBody := map[string]string{
		"app_id":     c.config.AppID,
		"app_secret": c.config.AppSecret,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewReader(body))
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

	var tokenResp struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}

	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return err
	}

	if tokenResp.Code != 0 {
		return fmt.Errorf("failed to get token: %s", tokenResp.Msg)
	}

	c.tokenMu.Lock()
	c.token = tokenResp.TenantAccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.Expire-300) * time.Second) // Refresh 5 min early
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

	// If webhook URL is configured, use webhook mode
	if c.config.WebhookURL != "" {
		return c.sendViaWebhook(ctx, msg, opts)
	}

	// Otherwise, use standard API mode
	token, err := c.getToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Determine receive_id_type based on chat type
	receiveIDType := "open_id"
	if opts.ChatType == types.ChatTypeGroup {
		receiveIDType = "chat_id"
	}

	// Build message content based on type
	var content string
	var msgType string
	switch msg.Type {
	case types.MessageTypeText:
		contentMap := map[string]string{"text": msg.Content}
		contentBytes, _ := json.Marshal(contentMap)
		content = string(contentBytes)
		msgType = "text"
	case types.MessageTypeMarkdown:
		contentMap := map[string]string{"text": msg.Content}
		contentBytes, _ := json.Marshal(contentMap)
		content = string(contentBytes)
		msgType = "post"
	case types.MessageTypeCard:
		// Interactive card - content should be the card JSON
		content = msg.Content
		msgType = "interactive"
	default:
		content = msg.Content
		msgType = string(msg.Type)
	}

	reqBody := map[string]interface{}{
		"receive_id": opts.Target,
		"msg_type":   msgType,
		"content":    content,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?receive_id_type=%s", sendMessageURL, receiveIDType)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

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
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return err
	}

	if apiResp.Code != 0 {
		return fmt.Errorf("lark API error: %s", apiResp.Msg)
	}

	return nil
}

// sendViaWebhook sends a message via webhook URL
func (c *Client) sendViaWebhook(ctx context.Context, msg *types.Message, opts *types.SendOptions) error {
	// Build webhook message body
	var reqBody map[string]interface{}

	switch msg.Type {
	case types.MessageTypeText:
		reqBody = map[string]interface{}{
			"msg_type": "text",
			"content": map[string]interface{}{
				"text": msg.Content,
			},
		}
	case types.MessageTypeMarkdown, types.MessageTypeCard:
		// For markdown and interactive cards
		reqBody = map[string]interface{}{
			"msg_type": "interactive",
			"card": map[string]interface{}{
				"elements": []map[string]interface{}{
					{
						"tag":     "markdown",
						"content": msg.Content,
					},
				},
			},
		}
	default:
		reqBody = map[string]interface{}{
			"msg_type": "text",
			"content": map[string]interface{}{
				"text": msg.Content,
			},
		}
	}

	// Add custom data if provided
	if msg.Data != nil {
		for k, v := range msg.Data {
			reqBody[k] = v
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.WebhookURL, bytes.NewReader(body))
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
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return err
	}

	if apiResp.Code != 0 {
		return fmt.Errorf("lark webhook error: %s", apiResp.Msg)
	}

	return nil
}

// SendPrivateMessage sends a private message to a user
func (c *Client) SendPrivateMessage(ctx context.Context, userID string, msg *types.Message) error {
	return c.SendMessage(ctx, msg, &types.SendOptions{
		ChatType: types.ChatTypePrivate,
		Target:   userID,
	})
}

// SendGroupMessage sends a message to a group
func (c *Client) SendGroupMessage(ctx context.Context, groupID string, msg *types.Message) error {
	return c.SendMessage(ctx, msg, &types.SendOptions{
		ChatType: types.ChatTypeGroup,
		Target:   groupID,
	})
}

// GetOpenIDByMobile gets user's open_id by mobile phone number
func (c *Client) GetOpenIDByMobile(ctx context.Context, mobile string) (string, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	reqBody := map[string]interface{}{
		"mobiles": []string{mobile},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s?user_id_type=open_id", batchGetIDURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			UserList []struct {
				UserID string `json:"user_id"`
				Mobile string `json:"mobile"`
			} `json:"user_list"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", err
	}

	if apiResp.Code != 0 {
		return "", fmt.Errorf("lark API error: %s", apiResp.Msg)
	}

	if len(apiResp.Data.UserList) == 0 {
		return "", fmt.Errorf("user not found with mobile: %s", mobile)
	}

	return apiResp.Data.UserList[0].UserID, nil
}

// GetOpenIDByEmail gets user's open_id by email address
func (c *Client) GetOpenIDByEmail(ctx context.Context, email string) (string, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	reqBody := map[string]interface{}{
		"emails": []string{email},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s?user_id_type=open_id", batchGetIDURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			UserList []struct {
				UserID string `json:"user_id"`
				Email  string `json:"email"`
			} `json:"user_list"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", err
	}

	if apiResp.Code != 0 {
		return "", fmt.Errorf("lark API error: %s", apiResp.Msg)
	}

	if len(apiResp.Data.UserList) == 0 {
		return "", fmt.Errorf("user not found with email: %s", email)
	}

	return apiResp.Data.UserList[0].UserID, nil
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
