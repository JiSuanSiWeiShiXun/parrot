package imparrot_test

import (
	"context"
	"testing"

	imparrot "github.com/JiSuanSiWeiShiXun/parrot"
	"github.com/JiSuanSiWeiShiXun/parrot/telegram"
	"github.com/JiSuanSiWeiShiXun/parrot/types"
)

// TestStrategyPattern demonstrates the strategy pattern
func TestStrategyPattern(t *testing.T) {
	// Create a client using factory method
	client, err := imparrot.NewTelegramClient("test-token")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify it implements IMParrot interface
	var _ imparrot.IMParrot = client

	// Verify platform name
	if client.GetPlatformName() != "telegram" {
		t.Errorf("Expected platform name 'telegram', got '%s'", client.GetPlatformName())
	}
}

// TestFactoryMethod demonstrates the factory method pattern
func TestFactoryMethod(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		config   imparrot.Config
		wantErr  bool
	}{
		{
			name:     "telegram client",
			platform: imparrot.PlatformTelegram,
			config: &telegram.Config{
				BotToken: "test-token",
			},
			wantErr: false,
		},
		{
			name:     "invalid config",
			platform: imparrot.PlatformTelegram,
			config: &telegram.Config{
				BotToken: "", // Empty token
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := imparrot.NewIMClient(tt.platform, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIMClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("Expected non-nil client")
			}
		})
	}
}

// TestMessageTypes tests different message types
func TestMessageTypes(t *testing.T) {
	messages := []imparrot.Message{
		{
			Type:    imparrot.MessageTypeText,
			Content: "Plain text message",
		},
		{
			Type:    imparrot.MessageTypeMarkdown,
			Content: "**Bold** text",
		},
		{
			Type:    imparrot.MessageTypeCard,
			Content: `{"title":"Card"}`,
		},
	}

	for _, msg := range messages {
		if msg.Type == "" {
			t.Error("Message type should not be empty")
		}
		if msg.Content == "" {
			t.Error("Message content should not be empty")
		}
	}
}

// TestSendOptions tests send options structure
func TestSendOptions(t *testing.T) {
	opts := &types.SendOptions{
		Targets: []types.Target{{ID: "user123", ChatType: types.ChatTypePrivate}},
		AtUsers: []string{"user1", "user2"},
		Extra: map[string]interface{}{
			"priority": "high",
		},
	}

	if opts.Targets[0].ChatType != types.ChatTypePrivate {
		t.Error("ChatType should be private")
	}
	if len(opts.AtUsers) != 2 {
		t.Error("Should have 2 AtUsers")
	}
	if opts.Extra["priority"] != "high" {
		t.Error("Extra priority should be high")
	}
}

// BenchmarkMessageCreation benchmarks message creation
func BenchmarkMessageCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &imparrot.Message{
			Type:    imparrot.MessageTypeText,
			Content: "Benchmark message",
			Data: map[string]interface{}{
				"timestamp": 123456789,
			},
		}
	}
}

// ExampleNewTelegramClient demonstrates creating a Telegram client
func ExampleNewTelegramClient() {
	client, err := imparrot.NewTelegramClient("your-bot-token")
	if err != nil {
		panic(err)
	}

	msg := &imparrot.Message{
		Type:    imparrot.MessageTypeText,
		Content: "Hello, World!",
	}

	_ = client.SendPrivateMessage(context.Background(), "chat-id", msg)
}
