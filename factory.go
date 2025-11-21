package imparrot

import (
	"fmt"
	"net/http"
	"time"

	"github.com/youling/im-parrot/dingtalk"
	"github.com/youling/im-parrot/lark"
	"github.com/youling/im-parrot/telegram"
	"github.com/youling/im-parrot/types"
	"github.com/youling/im-parrot/wechat"
)

// Platform constants
const (
	PlatformLark     = "lark"
	PlatformTelegram = "telegram"
	PlatformDingTalk = "dingtalk"
	PlatformWeChat   = "wechat"
	PlatformWPSXZ    = "wpsxz"
)

// Factory method pattern implementation
// NewIMClient creates a new IM client based on the platform and config
func NewIMClient(platform string, config types.Config) (types.IMParrot, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if config.GetPlatform() != platform {
		return nil, fmt.Errorf("config platform %s does not match requested platform %s",
			config.GetPlatform(), platform)
	}

	// Create shared HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Factory method - create different implementations based on platform
	switch platform {
	case PlatformLark:
		cfg, ok := config.(*lark.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for lark platform")
		}
		return lark.NewClient(cfg, httpClient)

	case PlatformTelegram:
		cfg, ok := config.(*telegram.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for telegram platform")
		}
		return telegram.NewClient(cfg, httpClient)

	case PlatformDingTalk:
		cfg, ok := config.(*dingtalk.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for dingtalk platform")
		}
		return dingtalk.NewClient(cfg, httpClient)

	case PlatformWeChat:
		cfg, ok := config.(*wechat.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for wechat platform")
		}
		return wechat.NewClient(cfg, httpClient)

	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// NewLarkClient is a convenience method for creating Lark client
func NewLarkClient(appID, appSecret string) (types.IMParrot, error) {
	config := &lark.Config{
		AppID:     appID,
		AppSecret: appSecret,
	}
	return NewIMClient(PlatformLark, config)
}

// NewLarkWebhookClient is a convenience method for creating Lark webhook client
func NewLarkWebhookClient(webhookURL string) (types.IMParrot, error) {
	config := &lark.Config{
		WebhookURL: webhookURL,
	}
	return NewIMClient(PlatformLark, config)
}

// NewTelegramClient is a convenience method for creating Telegram client
func NewTelegramClient(botToken string) (types.IMParrot, error) {
	config := &telegram.Config{
		BotToken: botToken,
	}
	return NewIMClient(PlatformTelegram, config)
}

// NewDingTalkClient is a convenience method for creating DingTalk client
func NewDingTalkClient(accessToken, secret string) (types.IMParrot, error) {
	config := &dingtalk.Config{
		AccessToken: accessToken,
		Secret:      secret,
	}
	return NewIMClient(PlatformDingTalk, config)
}

// NewWeChatClient is a convenience method for creating WeChat Work client
func NewWeChatClient(corpID, corpSecret string) (types.IMParrot, error) {
	config := &wechat.Config{
		CorpID:     corpID,
		CorpSecret: corpSecret,
	}
	return NewIMClient(PlatformWeChat, config)
}
