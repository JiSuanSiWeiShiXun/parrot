# im-parrot

封装飞书、钉钉、企业微信、Telegram 等 IM 的 SDK，提供统一的接口作为第三方库。

## 特性

- 🎯 **统一接口**: 使用策略模式，所有平台实现统一的 `IMParrot` 接口
- 🏭 **工厂方法**: 通过工厂模式创建不同平台的客户端
- 🚀 **简单易用**: 提供便捷方法快速创建客户端
- 📦 **零依赖**: 仅使用 Go 标准库，无第三方依赖
- 🔒 **类型安全**: 完整的类型定义和错误处理
- 🌐 **多平台支持**: 飞书、Telegram、钉钉、企业微信

## 设计模式

### 策略模式 (Strategy Pattern)
所有 IM 平台客户端实现统一的 `IMParrot` 接口：
```go
type IMParrot interface {
    SendMessage(ctx context.Context, msg *Message, opts *SendOptions) error
    SendPrivateMessage(ctx context.Context, userID string, msg *Message) error
    SendGroupMessage(ctx context.Context, groupID string, msg *Message) error
    GetPlatformName() string
}
```

### 工厂方法模式 (Factory Method Pattern)
使用工厂函数根据平台类型创建相应的客户端：
```go
client, err := imparrot.NewIMClient(imparrot.PlatformLark, config)
```

## 安装

```bash
go get github.com/JiSuanSiWeiShiXun/parrot
```

## 快速开始

### 1. 飞书 (Lark/Feishu)

```go
import (
    imparrot "github.com/JiSuanSiWeiShiXun/parrot"
)

// 使用便捷方法
client, err := imparrot.NewLarkClient("app-id", "app-secret")
if err != nil {
    log.Fatal(err)
}

msg := &imparrot.Message{
    Type:    imparrot.MessageTypeText,
    Content: "Hello from Lark!",
}

// 发送私聊消息
err = client.SendPrivateMessage(context.Background(), "user-open-id", msg)

// 发送群聊消息
err = client.SendGroupMessage(context.Background(), "chat-id", msg)
```

### 2. Telegram

```go
client, err := imparrot.NewTelegramClient("bot-token")
if err != nil {
    log.Fatal(err)
}

msg := &imparrot.Message{
    Type:    imparrot.MessageTypeMarkdown,
    Content: "**Hello** from Telegram!",
}

err = client.SendPrivateMessage(context.Background(), "chat-id", msg)
```

### 3. 钉钉 (DingTalk)

```go
client, err := imparrot.NewDingTalkClient("access-token", "secret")
if err != nil {
    log.Fatal(err)
}

msg := &imparrot.Message{
    Type:    imparrot.MessageTypeText,
    Content: "Hello from DingTalk!",
}

opts := &imparrot.SendOptions{
    ChatType: imparrot.ChatTypeGroup,
    Target:   "webhook",
    AtUsers:  []string{"138xxxxxxxx"}, // @特定用户
}

err = client.SendMessage(context.Background(), msg, opts)
```

### 4. 企业微信 (WeChat Work)

```go
import "github.com/JiSuanSiWeiShiXun/parrot/wechat"

config := &wechat.Config{
    CorpID:     "corp-id",
    CorpSecret: "corp-secret",
    AgentID:    1000002,
}

client, err := imparrot.NewIMClient(imparrot.PlatformWeChat, config)
if err != nil {
    log.Fatal(err)
}

msg := &imparrot.Message{
    Type:    imparrot.MessageTypeText,
    Content: "Hello from WeChat Work!",
}

err = client.SendPrivateMessage(context.Background(), "user-id", msg)
```

## 使用工厂方法

```go
import (
    imparrot "github.com/JiSuanSiWeiShiXun/parrot"
    "github.com/JiSuanSiWeiShiXun/parrot/lark"
)

// 创建配置
config := &lark.Config{
    AppID:     "your-app-id",
    AppSecret: "your-app-secret",
}

// 使用工厂方法创建客户端
client, err := imparrot.NewIMClient(imparrot.PlatformLark, config)
if err != nil {
    log.Fatal(err)
}

// 使用统一接口
msg := &imparrot.Message{
    Type:    imparrot.MessageTypeText,
    Content: "Hello!",
}

err = client.SendPrivateMessage(context.Background(), "user-id", msg)
```

## 消息类型

支持多种消息类型：

```go
// 文本消息
msg := &imparrot.Message{
    Type:    imparrot.MessageTypeText,
    Content: "纯文本消息",
}

// Markdown 消息
msg := &imparrot.Message{
    Type:    imparrot.MessageTypeMarkdown,
    Content: "## 标题\n\n**粗体** *斜体*",
}

// 自定义数据
msg := &imparrot.Message{
    Type:    imparrot.MessageTypeText,
    Content: "消息内容",
    Data: map[string]interface{}{
        "priority": "high",
        "custom_field": "value",
    },
}
```

## 发送选项

```go
opts := &imparrot.SendOptions{
    ChatType: imparrot.ChatTypeGroup,    // 群聊
    Target:   "group-id",                  // 目标ID
    AtUsers:  []string{"user1", "user2"}, // @用户（钉钉支持）
    Extra: map[string]interface{}{        // 平台特定参数
        "disable_notification": true,
    },
}

err := client.SendMessage(context.Background(), msg, opts)
```

## 策略模式示例

不同平台可互换使用：

```go
func sendToAllPlatforms(clients []imparrot.IMParrot, content string) {
    msg := &imparrot.Message{
        Type:    imparrot.MessageTypeText,
        Content: content,
    }
    
    for _, client := range clients {
        platform := client.GetPlatformName()
        log.Printf("Sending via %s...", platform)
        
        opts := &imparrot.SendOptions{
            ChatType: imparrot.ChatTypePrivate,
            Target:   "user-id",
        }
        
        if err := client.SendMessage(context.Background(), msg, opts); err != nil {
            log.Printf("Failed to send via %s: %v", platform, err)
        }
    }
}
```

## 项目结构

```
im-parrot/
├── interface.go          # IMParrot 接口定义
├── factory.go            # 工厂方法实现
├── go.mod
├── README.md
├── lark/                 # 飞书实现
│   └── lark.go
├── telegram/             # Telegram 实现
│   └── telegram.go
├── dingtalk/             # 钉钉实现
│   └── dingtalk.go
├── wechat/               # 企业微信实现
│   └── wechat.go
└── examples/             # 示例代码
    └── main.go
```

## 支持的平台

| 平台 | 状态 | 私聊 | 群聊 | 认证方式 |
|------|------|------|------|----------|
| 飞书 (Lark) | ✅ | ✅ | ✅ | App ID + Secret |
| Telegram | ✅ | ✅ | ✅ | Bot Token |
| 钉钉 (DingTalk) | ✅ | ❌ | ✅ | Webhook + Secret |
| 企业微信 (WeChat Work) | ✅ | ✅ | ✅ | Corp ID + Secret |

## 开发计划

- [ ] 添加单元测试
- [ ] 支持更多消息类型（图片、文件等）
- [ ] 添加消息模板功能
- [ ] 支持批量发送
- [ ] 添加重试机制
- [ ] 支持 WPS 协作

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License