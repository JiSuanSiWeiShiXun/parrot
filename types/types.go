package types

import "context"

// MessageType defines the type of message
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeMarkdown MessageType = "markdown"
	MessageTypeCard     MessageType = "card"
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

// SendOptions contains options for sending messages
type SendOptions struct {
	ChatType ChatType               // Private or group chat
	Target   string                 // User ID or Group ID
	AtUsers  []string               // Users to @ mention (for group messages)
	Extra    map[string]interface{} // Platform-specific extra options
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
}

// Config is the interface for platform configurations
type Config interface {
	Validate() error
	GetPlatform() string
}
