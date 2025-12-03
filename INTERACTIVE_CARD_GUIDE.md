# 飞书交互式卡片实现指南

## 需求
创建一个带按钮的消息卡片，点击按钮发送 POST 请求到 `http://dumpinfo.xoyo.com/dump_api/get_upload_info`，根据返回状态码更新按钮状态。

## 重要说明

⚠️ **飞书按钮的工作原理**：
飞书的交互式按钮**不能**直接发送 HTTP 请求并根据返回码更新。正确的流程是：

```
用户点击按钮 
    ↓
飞书服务器发送回调到你的服务器
    ↓
你的服务器处理回调，调用目标 API
    ↓
根据 API 返回结果更新卡片
```

## 实现步骤

### 1. 发送交互式卡片

已在 `lark_test.go` 中实现，运行测试：

```bash
go test -v -run TestParrotLarkInteractiveCard
```

卡片包含：
- 标题："API 请求工具"
- 目标 API 说明
- 🚀 发送请求按钮（主按钮）
- 📋 查看文档按钮（跳转链接）
- 底部提示信息

### 2. 配置飞书开发者后台

#### 2.1 配置请求地址（事件订阅）

1. 登录 [飞书开发者后台](https://open.feishu.cn/)
2. 选择你的应用
3. 进入「事件订阅」→「请求地址配置」
4. 填写你的服务器地址（必须是 HTTPS 公网可访问）
   ```
   https://your-domain.com/feishu/callback
   ```
5. 飞书会发送验证请求，你的服务器需要返回 challenge

#### 2.2 订阅事件

在「事件订阅」中添加：
- ✅ `im.message.card_action_triggered` - 消息卡片回传交互

#### 2.3 权限配置

确保应用有以下权限：
- ✅ 发送消息
- ✅ 获取用户信息
- ✅ 接收消息事件

### 3. 实现回调服务器

创建一个 HTTP 服务器处理飞书回调：

```go
package main

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
    "bytes"
)

// 回调处理器
func handleFeishuCallback(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    var callback map[string]interface{}
    json.Unmarshal(body, &callback)

    // 1. 处理 URL 验证（首次配置时）
    if challenge, ok := callback["challenge"].(string); ok {
        json.NewEncoder(w).Encode(map[string]string{
            "challenge": challenge,
        })
        return
    }

    // 2. 处理按钮点击事件
    if callback["type"] == "card.action.trigger" {
        // 异步处理，立即返回
        go processButtonClick(callback)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// 处理按钮点击
func processButtonClick(callback map[string]interface{}) {
    action := callback["action"].(map[string]interface{})
    value := action["value"].(map[string]interface{})
    
    // 获取按钮配置的数据
    targetURL := value["target_url"].(string)
    
    // 发送 POST 请求到目标 API
    resp, err := http.Post(targetURL, "application/json", 
        bytes.NewBuffer([]byte("{}")))
    
    // 根据结果更新卡片
    messageID := callback["open_message_id"].(string)
    
    if err != nil || resp.StatusCode != 200 {
        updateCardToError(messageID, resp.StatusCode, err)
    } else {
        updateCardToSuccess(messageID, resp.StatusCode)
    }
}

// 更新卡片为成功状态
func updateCardToSuccess(messageID string, statusCode int) {
    // 构建新卡片
    newCard := map[string]interface{}{
        "header": map[string]interface{}{
            "title": map[string]interface{}{
                "tag": "plain_text",
                "content": "✅ 请求成功",
            },
            "template": "green",
        },
        "elements": []interface{}{
            map[string]interface{}{
                "tag": "div",
                "text": map[string]interface{}{
                    "tag": "lark_md",
                    "content": "**状态码**: 200\n\n✓ API 请求执行成功",
                },
            },
        },
    }
    
    // 调用飞书 API 更新卡片
    // PATCH https://open.feishu.cn/open-apis/im/v1/messages/:message_id
    // 需要 tenant_access_token
}

func main() {
    http.HandleFunc("/feishu/callback", handleFeishuCallback)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 4. 更新卡片 API

使用飞书的更新消息 API：

```bash
PATCH https://open.feishu.cn/open-apis/im/v1/messages/{message_id}
Authorization: Bearer {tenant_access_token}
Content-Type: application/json

{
  "msg_type": "interactive",
  "content": "{...新卡片JSON...}"
}
```

### 5. 完整的卡片状态流转

#### 初始状态（蓝色）
```json
{
  "header": {
    "title": {"content": "API 请求工具"},
    "template": "blue"
  },
  "elements": [
    {
      "tag": "action",
      "actions": [{
        "tag": "button",
        "text": {"content": "🚀 发送请求"},
        "type": "primary",
        "value": {
          "action": "send_post_request",
          "target_url": "http://dumpinfo.xoyo.com/dump_api/get_upload_info"
        }
      }]
    }
  ]
}
```

#### 成功状态（绿色）
```json
{
  "header": {
    "title": {"content": "✅ 请求成功"},
    "template": "green"
  },
  "elements": [
    {
      "tag": "div",
      "text": {
        "tag": "lark_md",
        "content": "**状态码**: 200\n\n✓ 请求执行成功"
      }
    }
  ]
}
```

#### 失败状态（红色）
```json
{
  "header": {
    "title": {"content": "❌ 请求失败"},
    "template": "red"
  },
  "elements": [
    {
      "tag": "div",
      "text": {
        "tag": "lark_md",
        "content": "**状态码**: {code}\n\n错误信息..."
      }
    },
    {
      "tag": "action",
      "actions": [{
        "tag": "button",
        "text": {"content": "🔄 重试"},
        "type": "primary"
      }]
    }
  ]
}
```

## 测试流程

1. **发送卡片**
   ```bash
   go test -v -run TestParrotLarkInteractiveCard
   ```

2. **启动回调服务器**
   ```bash
   # 需要部署到公网服务器
   go run callback_server.go
   ```

3. **配置飞书后台**
   - 填写回调地址
   - 订阅事件

4. **点击按钮测试**
   - 在飞书客户端查看收到的卡片
   - 点击"发送请求"按钮
   - 观察卡片状态变化

## 开发建议

### 本地开发调试

使用内网穿透工具（如 ngrok）将本地服务暴露到公网：

```bash
# 启动本地服务
go run callback_server.go

# 在另一个终端启动 ngrok
ngrok http 8080

# 将 ngrok 提供的 HTTPS 地址配置到飞书后台
# 例如: https://abc123.ngrok.io/feishu/callback
```

### 日志记录

在回调服务器中记录详细日志：
```go
log.Printf("收到回调: %s", string(body))
log.Printf("按钮点击: action=%s", action)
log.Printf("API 请求: url=%s, status=%d", targetURL, statusCode)
```

### 错误处理

1. 超时处理（API 请求设置超时）
2. 重试机制（失败后允许重试）
3. 错误信息展示（清晰的错误提示）

## 参考文档

- [飞书开放平台 - 消息卡片](https://open.feishu.cn/document/ukTMukTMukTM/uczM3QjL3MzN04yNzcDN)
- [飞书开放平台 - 事件订阅](https://open.feishu.cn/document/ukTMukTMukTM/uUTNz4SN1MjL1UzM)
- [飞书开放平台 - 卡片消息](https://open.feishu.cn/document/ukTMukTMukTM/uEjNwUjLxYDM14SM2ATN)

## 总结

✅ 已实现：发送带按钮的交互式卡片  
⏳ 需要实现：回调服务器 + 卡片更新逻辑  
📋 需要配置：飞书开发者后台事件订阅

**关键点**：飞书的交互式按钮通过回调机制工作，不能直接在客户端发送 HTTP 请求。你需要实现一个服务器来接收回调、调用目标 API、并更新卡片状态。
