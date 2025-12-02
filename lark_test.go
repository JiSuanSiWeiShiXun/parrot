// File    :   lark_test.go
// Time    :   2025/12/02 16:36:33
// Contact :   youling15122511@gmail.com
// Desc    :   None
package imparrot

import (
	"context"
	"testing"

	"github.com/JiSuanSiWeiShiXun/parrot/lark"
	"github.com/JiSuanSiWeiShiXun/parrot/types"
)

func TestPirrotLark(t *testing.T) {
	config := &lark.Config{
		AppID:     "cli_a3b23fdb2438d00c",
		AppSecret: "fAX5un2l0S9ZcuIjoruDie1ilrKWYoDJ",
	}
	pirrot, err := NewIMClient(PlatformLark, config)
	if err != nil {
		t.Fatalf("Failed to create Lark client: %v", err)
	}

	msg := &types.Message{
		Type:    types.MessageTypeText,
		Content: "野猪拉屎啦!",
	}
	opt := &types.SendOptions{
		Target: "ou_9b1df8208ac284e95151b6e938115234",
	}
	if err := pirrot.SendMessage(context.TODO(), msg, opt); err != nil {
		t.Fatalf("Failed to send Lark message: %v", err)
	}
	t.Logf("Lark message sent successfully")
}
