package main

import (
	"context"
	"fmt"
	"log"

	imparrot "github.com/JiSuanSiWeiShiXun/parrot"
	"github.com/JiSuanSiWeiShiXun/parrot/dingtalk"
	"github.com/JiSuanSiWeiShiXun/parrot/lark"
	"github.com/JiSuanSiWeiShiXun/parrot/types"
	"github.com/JiSuanSiWeiShiXun/parrot/wechat"
)

func main() {
	ctx := context.Background()

	// Example 1: Using factory method with Lark
	fmt.Println("=== Example 1: Lark (Feishu) ===")
	larkConfig := &lark.Config{
		AppID:     "cli_a3b23fdb2438d00c",
		AppSecret: "fAX5un2l0S9ZcuIjoruDie1ilrKWYoDJ",
	}
	larkClient, err := imparrot.NewIMClient(imparrot.PlatformLark, larkConfig)
	if err != nil {
		log.Printf("Failed to create Lark client: %v", err)
	} else {
		userOpenID := "ou_9b1df8208ac284e95151b6e938115234" // 请在飞书后台获取真实的 open_id

		msg := &types.Message{
			Type:    types.MessageTypeText,
			Content: "Hello from Lark!",
		}
		if err := larkClient.SendPrivateMessage(ctx, userOpenID, msg); err != nil {
			log.Printf("Failed to send Lark message: %v", err)
			log.Printf("提示: 请在飞书开发者后台添加测试用户，获取正确的 open_id (格式: ou_xxxxxxxx)")
		} else {
			fmt.Println("✓ Lark message sent successfully")
		}
	}

	// Example 1.5: Using Lark Webhook (群机器人)
	fmt.Println("\n=== Example 1.5: Lark Webhook ===")
	larkWebhookClient, err := imparrot.NewLarkWebhookClient("https://open.feishu.cn/open-apis/bot/v2/hook/f42726a9-9e8d-4e33-af4a-6dd52ee3af97")
	if err != nil {
		log.Printf("Failed to create Lark webhook client: %v", err)
	} else {
		msg := &types.Message{
			Type:    types.MessageTypeText,
			Content: "Hello from Lark Webhook! 这是通过群机器人 Webhook 发送的消息",
		}
		// Webhook 模式不需要指定 target，消息会发到配置了该 webhook 的群
		opts := &types.SendOptions{
			ChatType: types.ChatTypeGroup,
			Target:   "", // Webhook 模式忽略此参数
		}
		if err := larkWebhookClient.SendMessage(ctx, msg, opts); err != nil {
			log.Printf("Failed to send Lark webhook message: %v", err)
			log.Printf("提示: 请在飞书群聊中添加机器人，获取 Webhook URL")
		} else {
			fmt.Println("✓ Lark webhook message sent successfully")
		}
	}

	// Example 2: Using convenience method for Telegram
	fmt.Println("\n=== Example 2: Telegram ===")
	telegramClient, err := imparrot.NewTelegramClient("your-bot-token")
	if err != nil {
		log.Printf("Failed to create Telegram client: %v", err)
	} else {
		msg := &types.Message{
			Type:    types.MessageTypeText,
			Content: "Hello from Telegram!",
		}
		if err := telegramClient.SendPrivateMessage(ctx, "123456789", msg); err != nil {
			log.Printf("Failed to send Telegram message: %v", err)
		} else {
			fmt.Println("✓ Telegram message sent successfully")
		}
	}

	// Example 3: DingTalk group message with @ mention
	fmt.Println("\n=== Example 3: DingTalk ===")
	dingTalkConfig := &dingtalk.Config{
		AccessToken: "your-access-token",
		Secret:      "your-secret",
	}
	dingTalkClient, err := imparrot.NewIMClient(imparrot.PlatformDingTalk, dingTalkConfig)
	if err != nil {
		log.Printf("Failed to create DingTalk client: %v", err)
	} else {
		msg := &types.Message{
			Type:    types.MessageTypeMarkdown,
			Content: "## Hello from DingTalk\n\nThis is a **markdown** message!",
		}
		opts := &types.SendOptions{
			ChatType: types.ChatTypeGroup,
			Target:   "webhook-group",
			AtUsers:  []string{"138xxxxxxxx"}, // Phone numbers to mention
		}
		if err := dingTalkClient.SendMessage(ctx, msg, opts); err != nil {
			log.Printf("Failed to send DingTalk message: %v", err)
		} else {
			fmt.Println("✓ DingTalk message sent successfully")
		}
	}

	// Example 4: WeChat Work message
	fmt.Println("\n=== Example 4: WeChat Work ===")
	wechatConfig := &wechat.Config{
		CorpID:     "your-corp-id",
		CorpSecret: "your-corp-secret",
		AgentID:    1000002,
	}
	wechatClient, err := imparrot.NewIMClient(imparrot.PlatformWeChat, wechatConfig)
	if err != nil {
		log.Printf("Failed to create WeChat client: %v", err)
	} else {
		msg := &types.Message{
			Type:    types.MessageTypeText,
			Content: "Hello from WeChat Work!",
		}
		if err := wechatClient.SendPrivateMessage(ctx, "UserID", msg); err != nil {
			log.Printf("Failed to send WeChat message: %v", err)
		} else {
			fmt.Println("✓ WeChat message sent successfully")
		}
	}

	// Example 5: Strategy pattern demonstration - using unified interface
	fmt.Println("\n=== Example 5: Strategy Pattern Demo ===")
	demonstrateStrategyPattern(ctx)

	fmt.Println("\n=== All examples completed ===")
}

// demonstrateStrategyPattern shows how different IM platforms can be used interchangeably
func demonstrateStrategyPattern(ctx context.Context) {
	// Create multiple clients (using mock configs)
	clients := []types.IMParrot{}

	// Add Telegram client
	if telegramClient, err := imparrot.NewTelegramClient("mock-token"); err == nil {
		clients = append(clients, telegramClient)
	}

	// Strategy pattern: Send the same message through different platforms
	msg := &types.Message{
		Type:    types.MessageTypeText,
		Content: "This message is sent through multiple platforms using strategy pattern!",
	}

	for _, client := range clients {
		platform := client.GetPlatformName()
		fmt.Printf("Sending via %s... ", platform)

		// Same interface, different implementations
		opts := &types.SendOptions{
			ChatType: types.ChatTypePrivate,
			Target:   "test-user",
		}

		if err := client.SendMessage(ctx, msg, opts); err != nil {
			fmt.Printf("Failed: %v\n", err)
		} else {
			fmt.Printf("Success!\n")
		}
	}
}

// Example 6: Custom message builder helper
func buildRichMessage() *types.Message {
	return &types.Message{
		Type:    types.MessageTypeMarkdown,
		Content: "# Important Notification\n\n- Point 1\n- Point 2\n- Point 3",
		Data: map[string]interface{}{
			"disable_notification": false,
			"priority":             "high",
		},
	}
}
