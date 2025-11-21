package imparrot

import (
	"github.com/youling/im-parrot/types"
)

// Re-export types for convenience
type (
	MessageType = types.MessageType
	ChatType    = types.ChatType
	Message     = types.Message
	SendOptions = types.SendOptions
	IMParrot    = types.IMParrot
	Config      = types.Config
)

// Re-export constants
const (
	MessageTypeText     = types.MessageTypeText
	MessageTypeMarkdown = types.MessageTypeMarkdown
	MessageTypeCard     = types.MessageTypeCard

	ChatTypePrivate = types.ChatTypePrivate
	ChatTypeGroup   = types.ChatTypeGroup
)
