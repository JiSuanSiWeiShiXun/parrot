package main

import (
	"context"
	"fmt"
	"log"
	"time"

	imparrot "github.com/JiSuanSiWeiShiXun/parrot"
	"github.com/JiSuanSiWeiShiXun/parrot/lark"
	"github.com/JiSuanSiWeiShiXun/parrot/telegram"
	"github.com/JiSuanSiWeiShiXun/parrot/types"
)

// Example: Using ClientPool for a message forwarding server
// This demonstrates how to prevent resource leaks when managing many bot clients

func main() {
	ctx := context.Background()

	// Create a client pool with custom config
	poolConfig := &imparrot.PoolConfig{
		MaxIdleTime:         10 * time.Minute, // Close idle clients after 10 minutes
		CleanupInterval:     2 * time.Minute,  // Check for idle clients every 2 minutes
		HTTPTimeout:         30 * time.Second,
		MaxIdleConns:        100, // Maximum idle connections across all hosts
		MaxIdleConnsPerHost: 10,  // Maximum idle connections per host
	}
	pool := imparrot.NewClientPool(poolConfig)
	defer pool.Close() // IMPORTANT: Always close the pool when done

	// Example 1: Get or create clients dynamically
	fmt.Println("=== Example 1: Dynamic Client Management ===")

	// Simulate receiving multiple requests for different bots
	requests := []struct {
		botKey   string
		platform string
		config   types.Config
		userID   string
		message  string
	}{
		{
			botKey:   "lark:bot1",
			platform: imparrot.PlatformLark,
			config: &lark.Config{
				AppID:     "cli_a3b23fdb2438d00c",
				AppSecret: "fAX5un2l0S9ZcuIjoruDie1ilrKWYoDJ",
			},
			userID:  "ou_9b1df8208ac284e95151b6e938115234",
			message: "Hello from bot1!",
		},
		{
			botKey:   "telegram:bot1",
			platform: imparrot.PlatformTelegram,
			config: &telegram.Config{
				BotToken: "your-bot-token-1",
			},
			userID:  "123456789",
			message: "Hello from Telegram bot1!",
		},
		{
			botKey:   "lark:bot2",
			platform: imparrot.PlatformLark,
			config: &lark.Config{
				AppID:     "cli_another_bot",
				AppSecret: "another_secret",
			},
			userID:  "ou_another_user",
			message: "Hello from bot2!",
		},
	}

	// Process each request
	for i, req := range requests {
		fmt.Printf("\nRequest %d: %s\n", i+1, req.botKey)

		// GetOrCreate will reuse existing client or create new one
		client, err := pool.GetOrCreate(ctx, req.botKey, req.platform, req.config)
		if err != nil {
			log.Printf("Failed to get/create client: %v", err)
			continue
		}

		msg := &types.Message{
			Type:    types.MessageTypeText,
			Content: req.message,
		}

		if err := client.SendPrivateMessage(ctx, req.userID, msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		} else {
			fmt.Printf("✓ Message sent successfully via %s\n", req.botKey)
		}

		fmt.Printf("Pool size: %d clients\n", pool.Size())
	}

	// Example 2: Reusing clients (no new creation)
	fmt.Println("\n=== Example 2: Client Reuse ===")

	// This will reuse the existing lark:bot1 client (no new HTTP connections)
	client, err := pool.GetOrCreate(ctx, "lark:bot1", imparrot.PlatformLark, &lark.Config{
		AppID:     "cli_a3b23fdb2438d00c",
		AppSecret: "fAX5un2l0S9ZcuIjoruDie1ilrKWYoDJ",
	})
	if err != nil {
		log.Printf("Failed to get client: %v", err)
	} else {
		msg := &types.Message{
			Type:    types.MessageTypeText,
			Content: "This message reuses the existing client!",
		}
		if err := client.SendPrivateMessage(ctx, "ou_9b1df8208ac284e95151b6e938115234", msg); err != nil {
			log.Printf("Failed to send: %v", err)
		} else {
			fmt.Println("✓ Message sent with reused client")
		}
	}

	fmt.Printf("Pool still has %d clients (no new client was created)\n", pool.Size())

	// Example 3: Manual client removal
	fmt.Println("\n=== Example 3: Manual Cleanup ===")

	// Remove a specific client when you know it won't be needed
	if err := pool.Remove("telegram:bot1"); err != nil {
		log.Printf("Failed to remove client: %v", err)
	} else {
		fmt.Println("✓ Manually removed telegram:bot1")
	}
	fmt.Printf("Pool size after removal: %d clients\n", pool.Size())

	// Example 4: Direct client usage (without pool)
	fmt.Println("\n=== Example 4: Direct Client Usage (Manual Management) ===")

	// If you're only using a single bot, you can create and manage it directly
	directClient, err := imparrot.NewLarkClient("cli_a3b23fdb2438d00c", "fAX5un2l0S9ZcuIjoruDie1ilrKWYoDJ")
	if err != nil {
		log.Printf("Failed to create direct client: %v", err)
	} else {
		defer directClient.Close() // IMPORTANT: Always close when done

		msg := &types.Message{
			Type:    types.MessageTypeText,
			Content: "Direct client usage",
		}
		if err := directClient.SendPrivateMessage(ctx, "ou_9b1df8208ac284e95151b6e938115234", msg); err != nil {
			log.Printf("Failed to send: %v", err)
		} else {
			fmt.Println("✓ Message sent with direct client")
		}
	}

	// The pool will automatically clean up idle clients based on MaxIdleTime
	fmt.Println("\n=== Automatic Cleanup ===")
	fmt.Println("Idle clients will be automatically cleaned up after MaxIdleTime")
	fmt.Println("You can also call pool.Close() to immediately close all clients")

	fmt.Println("\n=== Summary ===")
	fmt.Println("✓ Use ClientPool for message forwarding servers with many bots")
	fmt.Println("✓ Always call defer pool.Close() or defer client.Close()")
	fmt.Println("✓ Clients are automatically cleaned up when idle")
	fmt.Println("✓ Shared HTTP connection pool prevents resource exhaustion")
}
