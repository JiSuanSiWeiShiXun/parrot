package imparrot

import (
	"github.com/JiSuanSiWeiShiXun/parrot/types"
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

// Re-export popular constants
const (
	MessageTypeText     = types.MessageTypeText
	MessageTypeMarkdown = types.MessageTypeMarkdown
	MessageTypeCard     = types.MessageTypeCard

	ChatTypePrivate = types.ChatTypePrivate
	ChatTypeGroup   = types.ChatTypeGroup
)
