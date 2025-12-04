package types

import (
	"context"
	"fmt"
)

// MessageType defines the type of message
type MessageType string

const (
	MessageTypeText      MessageType = "text"       // 文本消息
	MessageTypeMarkdown  MessageType = "markdown"   // 富文本消息 (飞书 post 格式)
	MessageTypePost      MessageType = "post"       // 富文本消息 (飞书 用json定义的富文本格式)
	MessageTypeImage     MessageType = "image"      // 图片消息
	MessageTypeCard      MessageType = "card"       // 交互式卡片
	MessageTypeShareChat MessageType = "share_chat" // 群名片
	MessageTypeShareUser MessageType = "share_user" // 用户名片
	MessageTypeAudio     MessageType = "audio"      // 音频消息
	MessageTypeMedia     MessageType = "media"      // 视频消息
	MessageTypeFile      MessageType = "file"       // 文件消息
	MessageTypeSticker   MessageType = "sticker"    // 表情包
	MessageTypeSystem    MessageType = "system"     // 系统消息
)

// ChatType defines the type of chat
type ChatType string

const (
	ChatTypePrivate ChatType = "private"
	ChatTypeGroup   ChatType = "group"
)

// Message represents a unified message structure
type Message struct {
	Type    MessageType            // Message type: text, markdown, card
	Content string                 // Message content
	Data    map[string]interface{} // Additional platform-specific data
}

// Target represents a message destination
type Target struct {
	ID       string   // User ID or Group ID
	ChatType ChatType // Private or group chat
}

// SendOptions contains options for sending messages
type SendOptions struct {
	Targets []Target               // Multiple targets with their chat types
	AtUsers []string               // Users to @ mention (for group messages)
	Extra   map[string]interface{} // Platform-specific extra options
}

// FailedTarget represents a target that failed to receive a message
type FailedTarget struct {
	Target Target // The target that failed
	Error  error  // The error that occurred
}

func (ft FailedTarget) String() string {
	return fmt.Sprintf("{Target: %v, Error: %v}", ft.Target, ft.Error)
}

// SendError is returned when some targets fail to receive messages
type SendError struct {
	FailedTargets []FailedTarget // List of targets that failed
	SuccessCount  int            // Number of successful sends
	TotalCount    int            // Total number of targets
}

// Error implements the error interface
func (e *SendError) Error() string {
	failedInfos := ""
	for _, ft := range e.FailedTargets {
		failedInfos += ft.String() + "; "
	}
	return fmt.Sprintf("failed to send to %d/%d targets, failed targets: %v", len(e.FailedTargets), e.TotalCount, failedInfos)
}

// IMParrot is the unified interface for all IM platforms
// Using strategy pattern - different platforms implement this interface
type IMParrot interface {
	// SendMessage sends a message with options
	SendMessage(ctx context.Context, msg *Message, opts *SendOptions) error

	// SendPrivateMessage sends a private message to a user
	SendPrivateMessage(ctx context.Context, userID string, msg *Message) error

	// SendGroupMessage sends a message to a group
	SendGroupMessage(ctx context.Context, groupID string, msg *Message) error

	// GetPlatformName returns the platform name
	GetPlatformName() string

	// Close releases all resources held by the client
	// Should be called when the client is no longer needed to prevent resource leaks
	Close() error
}

// Config is the interface for platform configurations
type Config interface {
	Validate() error
	GetPlatform() string
}
