# Parrot 消息类型完整示例

本文档提供所有支持的消息类型的完整代码示例。

## 基本设置

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    
    "github.com/JiSuanSiWeiShiXun/parrot"
    "github.com/JiSuanSiWeiShiXun/parrot/lark"
    "github.com/JiSuanSiWeiShiXun/parrot/types"
)

func main() {
    // 创建客户端
    config := &lark.Config{
        AppID:     "your_app_id",
        AppSecret: "your_app_secret",
    }
    
    client, err := imparrot.NewIMClient(imparrot.PlatformLark, config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    userID := "ou_xxx" // 接收者的 open_id
    
    // 发送消息示例见下文
}
```

## 1. 文本消息 (Text)

### 基本文本
```go
msg := &types.Message{
    Type:    types.MessageTypeText,
    Content: "这是一条文本消息",
}

opt := &types.SendOptions{
    Targets: []types.Target{{ID: userID, ChatType: types.ChatTypePrivate}},
}

client.SendMessage(context.Background(), msg, opt)
```

### 带格式的文本
```go
msg := &types.Message{
    Type:    types.MessageTypeText,
    Content: `<b>重要通知</b>

项目状态：<u>进行中</u>
负责人：<at user_id="ou_xxx">张三</at>

详情：[查看文档](https://example.com)`,
}
```

## 2. Markdown 消息（简化版）

```go
msg := &types.Message{
    Type: types.MessageTypeMarkdown,
    Content: `# 项目进度报告

## 本周完成

1. 完成核心功能开发
2. 编写单元测试
3. 更新文档

## 下周计划

- [ ] 性能优化
- [ ] 集成测试
- [ ] 上线准备

**负责人**：<at user_id="ou_xxx">张三</at>
**详情**：[查看项目](https://example.com)

` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

> 注：所有任务按计划进行中`,
}
```

## 3. Post 消息（原始格式）

```go
postContent := map[string]interface{}{
    "zh_cn": map[string]interface{}{
        "title": "系统通知",
        "content": [][]map[string]interface{}{
            { // 第一段：标题行
                {
                    "tag":   "text",
                    "text":  "重要更新：",
                    "style": []string{"bold", "underline"},
                },
            },
            { // 第二段：正文
                {
                    "tag":  "text",
                    "text": "系统将在今晚 ",
                },
                {
                    "tag":   "text",
                    "text":   "22:00",
                    "style":  []string{"bold"},
                },
                {
                    "tag":  "text",
                    "text": " 进行升级维护，预计持续 2 小时。",
                },
            },
            { // 第三段：链接
                {
                    "tag":  "text",
                    "text": "详情请查看：",
                },
                {
                    "tag":   "a",
                    "href":  "https://example.com/notice",
                    "text":  "维护公告",
                    "style": []string{"italic"},
                },
            },
            { // 分隔线
                {
                    "tag": "hr",
                },
            },
            { // 代码块
                {
                    "tag":      "code_block",
                    "language": "bash",
                    "text":     "# 升级步骤\nsudo systemctl stop service\nsudo apt update\nsudo apt upgrade",
                },
            },
        },
    },
}

contentJSON, _ := json.Marshal(postContent)
msg := &types.Message{
    Type:    types.MessageTypePost,
    Content: string(contentJSON),
}
```

## 4. 卡片消息 (Card)

### 简单通知卡片
```go
card := map[string]interface{}{
    "config": map[string]interface{}{
        "wide_screen_mode": true,
    },
    "header": map[string]interface{}{
        "title": map[string]interface{}{
            "tag":     "plain_text",
            "content": "代码审查通知",
        },
        "template": "blue",
    },
    "elements": []map[string]interface{}{
        {
            "tag": "markdown",
            "content": `**项目**：Parrot IM 库
**提交者**：张三
**时间**：2024-01-01 10:00

本次更新包含以下内容：
- 新增消息类型支持
- 优化错误处理
- 更新文档`,
        },
        {
            "tag": "hr",
        },
        {
            "tag": "action",
            "actions": []map[string]interface{}{
                {
                    "tag": "button",
                    "text": map[string]interface{}{
                        "tag":     "plain_text",
                        "content": "查看详情",
                    },
                    "type": "primary",
                    "url":  "https://github.com/user/repo/pull/123",
                },
                {
                    "tag": "button",
                    "text": map[string]interface{}{
                        "tag":     "plain_text",
                        "content": "批准",
                    },
                    "type": "primary",
                },
            },
        },
    },
}

cardJSON, _ := json.Marshal(card)
msg := &types.Message{
    Type:    types.MessageTypeCard,
    Content: string(cardJSON),
}
```

### 丰富内容卡片
```go
card := map[string]interface{}{
    "config": map[string]interface{}{
        "wide_screen_mode": true,
    },
    "header": map[string]interface{}{
        "title": map[string]interface{}{
            "tag":     "plain_text",
            "content": "📊 每周数据报告",
        },
        "template": "green",
    },
    "elements": []map[string]interface{}{
        {
            "tag": "markdown",
            "content": "### 本周关键指标\n\n📈 **用户增长**：+15%\n💰 **收入**：+20%\n🎯 **目标达成率**：85%",
        },
        {
            "tag": "hr",
        },
        {
            "tag": "note",
            "elements": []map[string]interface{}{
                {
                    "tag":     "plain_text",
                    "content": "数据统计时间：2024-01-01 至 2024-01-07",
                },
            },
        },
        {
            "tag": "action",
            "actions": []map[string]interface{}{
                {
                    "tag": "button",
                    "text": map[string]interface{}{
                        "tag":     "plain_text",
                        "content": "查看完整报告",
                    },
                    "type": "primary",
                    "url":  "https://dashboard.example.com",
                },
            },
        },
    },
}

cardJSON, _ := json.Marshal(card)
msg := &types.Message{
    Type:    types.MessageTypeCard,
    Content: string(cardJSON),
}
```

## 5. 图片消息 (Image)

```go
// 先上传图片获取 image_key
// 参考：https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/image/create

content := map[string]string{
    "image_key": "img_v2_041b28e3-5680-48c2-9d2a-3b7d5a0f4e2g",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeImage,
    Content: string(contentJSON),
}
```

## 6. 文件消息 (File)

```go
// 先上传文件获取 file_key
// 参考：https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/file/create

content := map[string]string{
    "file_key": "file_v2_a3c4b5d6-e7f8-9a0b-1c2d-3e4f5a6b7c8d",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeFile,
    Content: string(contentJSON),
}
```

## 7. 音频消息 (Audio)

```go
content := map[string]string{
    "file_key": "file_v2_audio_xxx",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeAudio,
    Content: string(contentJSON),
}
```

## 8. 视频消息 (Media)

```go
// 视频需要 mp4 格式，可选配置封面图
content := map[string]string{
    "file_key":  "file_v2_video_xxx",
    "image_key": "img_v2_cover_xxx", // 可选
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeMedia,
    Content: string(contentJSON),
}
```

## 9. 分享群名片 (ShareChat)

```go
// 机器人必须在要分享的群中
content := map[string]string{
    "chat_id": "oc_a1b2c3d4e5f6g7h8i9j0",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeShareChat,
    Content: string(contentJSON),
}
```

## 10. 分享用户名片 (ShareUser)

```go
// user_id 必须是 open_id 格式
content := map[string]string{
    "user_id": "ou_a1b2c3d4e5f6g7h8i9j0",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeShareUser,
    Content: string(contentJSON),
}
```

## 11. 系统消息 (System)

```go
// 仅支持单聊，需要特殊权限
systemContent := map[string]interface{}{
    "type": "divider",
    "params": map[string]interface{}{
        "divider_text": map[string]interface{}{
            "text": "新会话",
            "i18n_text": map[string]string{
                "zh_CN": "新会话",
                "en_US": "New Session",
            },
        },
    },
    "options": map[string]bool{
        "need_rollup": true,
    },
}
contentJSON, _ := json.Marshal(systemContent)

msg := &types.Message{
    Type:    types.MessageTypeSystem,
    Content: string(contentJSON),
}
```

## 12. 表情包 (Sticker)

```go
// 仅支持转发接收到的表情包
content := map[string]string{
    "file_key": "sticker_file_key_xxx",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeSticker,
    Content: string(contentJSON),
}
```

## 批量发送示例

### 发送给多个用户
```go
msg := &types.Message{
    Type:    types.MessageTypeText,
    Content: "批量通知消息",
}

opt := &types.SendOptions{
    Targets: []types.Target{
        {ID: "ou_user1", ChatType: types.ChatTypePrivate},
        {ID: "ou_user2", ChatType: types.ChatTypePrivate},
        {ID: "ou_user3", ChatType: types.ChatTypePrivate},
    },
}

err := client.SendMessage(context.Background(), msg, opt)
if sendErr, ok := err.(*types.SendError); ok {
    log.Printf("成功：%d/%d", sendErr.SuccessCount, sendErr.TotalCount)
    for _, failed := range sendErr.FailedTargets {
        log.Printf("失败：%v - %v", failed.Target, failed.Error)
    }
}
```

### 发送给用户和群
```go
opt := &types.SendOptions{
    Targets: []types.Target{
        {ID: "ou_user1", ChatType: types.ChatTypePrivate},
        {ID: "oc_group1", ChatType: types.ChatTypeGroup},
    },
}
```

## 错误处理

```go
err := client.SendMessage(context.Background(), msg, opt)
if err != nil {
    // 检查是否是批量发送错误
    if sendErr, ok := err.(*types.SendError); ok {
        log.Printf("发送结果：成功 %d/%d", sendErr.SuccessCount, sendErr.TotalCount)
        
        // 处理失败的目标
        for _, failed := range sendErr.FailedTargets {
            log.Printf("目标 %s 发送失败：%v", failed.Target.ID, failed.Error)
            // 可以实现重试逻辑
        }
    } else {
        // 其他错误
        log.Printf("发送失败：%v", err)
    }
}
```

## 完整示例程序

```go
package main

import (
    "context"
    "log"
    
    "github.com/JiSuanSiWeiShiXun/parrot"
    "github.com/JiSuanSiWeiShiXun/parrot/lark"
    "github.com/JiSuanSiWeiShiXun/parrot/types"
)

func main() {
    // 1. 创建客户端
    config := &lark.Config{
        AppID:     "your_app_id",
        AppSecret: "your_app_secret",
    }
    
    client, err := imparrot.NewIMClient(imparrot.PlatformLark, config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // 2. 准备消息
    msg := &types.Message{
        Type: types.MessageTypeMarkdown,
        Content: `# 系统通知

**内容**：系统升级完成
**时间**：2024-01-01 12:00
**状态**：✅ 成功`,
    }
    
    // 3. 设置接收者
    opt := &types.SendOptions{
        Targets: []types.Target{
            {ID: "ou_user123", ChatType: types.ChatTypePrivate},
        },
    }
    
    // 4. 发送消息
    if err := client.SendMessage(context.Background(), msg, opt); err != nil {
        log.Printf("发送失败：%v", err)
    } else {
        log.Println("发送成功")
    }
}
```

## 测试代码

查看以下测试文件获取更多示例：
- `lark_test.go` - 基本功能测试
- `lark_message_types_test.go` - 各种消息类型测试
- `lark_advanced_types_test.go` - 高级消息类型测试

运行测试：
```bash
# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestLarkTextMessageFormats
go test -v -run TestLarkPostMessageFormats
go test -v -run TestLarkCardMessageFormats
```
