# 飞书消息类型使用指南

本文档详细说明了 Parrot 库支持的所有飞书消息类型及其使用方法。

## 消息类型概览

Parrot 支持以下飞书消息类型：

| 类型 | 常量 | 说明 | API msg_type |
|------|------|------|--------------|
| 文本消息 | `types.MessageTypeText` | 支持基本格式化标签 | `text` |
| 富文本消息(简化) | `types.MessageTypeMarkdown` | 支持 Markdown 语法，自动包装 | `post` |
| 富文本消息(原始) | `types.MessageTypePost` | 原始 post 格式，需完整 JSON | `post` |
| 图片消息 | `types.MessageTypeImage` | 发送图片 | `image` |
| 卡片消息 | `types.MessageTypeCard` | 交互式卡片 | `interactive` |
| 群名片 | `types.MessageTypeShareChat` | 分享群聊 | `share_chat` |
| 用户名片 | `types.MessageTypeShareUser` | 分享用户 | `share_user` |
| 音频消息 | `types.MessageTypeAudio` | 发送音频 | `audio` |
| 视频消息 | `types.MessageTypeMedia` | 发送视频 | `media` |
| 文件消息 | `types.MessageTypeFile` | 发送文件 | `file` |
| 表情包 | `types.MessageTypeSticker` | 发送表情包 | `sticker` |
| 系统消息 | `types.MessageTypeSystem` | 系统分割线 | `system` |

## 1. 文本消息 (Text)

### 支持的功能

文本消息支持以下格式化功能：

#### 1.1 换行
使用 `\n` 换行符：
```go
msg := &types.Message{
    Type:    types.MessageTypeText,
    Content: "第一行\n第二行\n第三行",
}
```

#### 1.2 @提到用户
```go
content := `<at user_id="ou_xxx">用户名</at> 你好！`
```

@所有人：
```go
content := `<at user_id="all"></at> 大家好！`
```

#### 1.3 样式标签

- **粗体**：`<b>文本</b>`
- **斜体**：`<i>文本</i>`
- **下划线**：`<u>文本</u>`
- **删除线**：`<s>文本</s>`

标签可以嵌套：
```go
content := "<b>粗体<i>加斜体</i></b>"
```

#### 1.4 超链接
```go
content := "[飞书开放平台](https://open.feishu.cn)"
```

### 组合示例

```go
msg := &types.Message{
    Type:    types.MessageTypeText,
    Content: `<b>通知</b>：<at user_id="ou_xxx">张三</at>
项目进度：<u>已完成</u>
详情请查看：[项目文档](https://example.com)`,
}
```

### 注意事项

- 样式标签需要正确嵌套，否则会显示原始内容
- 标签会增加消息体大小，请适当使用
- 不支持自定义机器人和批量发送接口

## 2. 富文本消息 (Post/Markdown)

富文本消息使用飞书的 post 格式，内部使用 `md` 标签支持 Markdown 语法。

### 支持的 Markdown 语法

#### 2.1 标题
```go
content := "# 一级标题\n## 二级标题\n### 三级标题"
```

#### 2.2 文本样式

- **粗体**：`**文本**`
- **斜体**：`*文本*`
- **粗体+斜体**：`***文本***`
- **下划线**：`~文本~`
- **删除线**：`~~文本~~`

```go
content := "**粗体** *斜体* ***粗体加斜体***\n~下划线~ ~~删除线~~"
```

#### 2.3 超链接
```go
content := "[飞书开放平台](https://open.feishu.cn)"
```

#### 2.4 @提到用户
```go
content := `<at user_id="ou_xxx">用户</at> 你好！`
```

#### 2.5 列表

**有序列表**：
```go
content := "1. 第一项\n2. 第二项\n3. 第三项"
```

**无序列表**：
```go
content := "- 项目1\n- 项目2\n- 项目3"
```

**嵌套列表**（每级缩进4个空格）：
```go
content := `1. 一级列表
    1. 二级列表1
    2. 二级列表2
2. 一级列表2`
```

#### 2.6 代码块
````go
content := "```GO\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```"
````

支持的语言类型：PYTHON, C, CPP, GO, JAVA, KOTLIN, SWIFT, PHP, RUBY, RUST, JAVASCRIPT, TYPESCRIPT, BASH, SHELL, SQL, JSON, XML, YAML, HTML, THRIFT 等（不区分大小写）

#### 2.7 引用
```go
content := "> 这是一段引用文本"
```

#### 2.8 分隔线
```go
content := "上方内容\n\n---\n\n下方内容"
```

### 综合示例

```go
msg := &types.Message{
    Type: types.MessageTypeMarkdown,
    Content: `# 项目进度报告

## 完成情况

**项目名称**：Parrot IM 库
**负责人**：<at user_id="ou_xxx">张三</at>

### 已完成功能

1. 文本消息支持
    - 格式化标签
    - @提到用户
2. 富文本消息支持
    - Markdown 语法
    - 代码高亮

### 代码示例

` + "```GO\nfunc SendMessage() {\n    // 实现代码\n}\n```" + `

> 注：所有功能已通过测试

---

详情请查看：[项目文档](https://example.com)`,
}
```

### 注意事项

- 粗体、斜体、下划线、删除线的文本不支持解析其他组件（如超链接）
- 粗体和斜体可以组合使用，但下划线和删除线不支持与其他样式组合
- 列表嵌套时每级缩进 4 个空格
- 代码块语言类型不区分大小写

## 3. 卡片消息 (Interactive Card)

卡片消息提供丰富的交互式界面，需要传入完整的卡片 JSON 结构。

### 基本卡片结构

```go
card := map[string]interface{}{
    "config": map[string]interface{}{
        "wide_screen_mode": true,
    },
    "header": map[string]interface{}{
        "title": map[string]interface{}{
            "tag":     "plain_text",
            "content": "卡片标题",
        },
        "template": "blue", // blue, green, orange, red, purple, etc.
    },
    "elements": []map[string]interface{}{
        // 卡片元素
    },
}

cardJSON, _ := json.Marshal(card)
msg := &types.Message{
    Type:    types.MessageTypeCard,
    Content: string(cardJSON),
}
```

### 常用元素

#### 3.1 Markdown 内容
```go
{
    "tag":     "markdown",
    "content": "这是**Markdown**内容\n- 列表项1\n- 列表项2",
}
```

#### 3.2 按钮
```go
{
    "tag": "action",
    "actions": []map[string]interface{}{
        {
            "tag": "button",
            "text": map[string]interface{}{
                "tag":     "plain_text",
                "content": "按钮文字",
            },
            "type": "primary", // primary, default, danger
            "url":  "https://example.com",
        },
    },
}
```

#### 3.3 分隔线
```go
{
    "tag": "hr",
}
```

#### 3.4 备注
```go
{
    "tag": "note",
    "elements": []map[string]interface{}{
        {
            "tag":     "plain_text",
            "content": "这是备注信息",
        },
    },
}
```

### 完整示例

```go
card := map[string]interface{}{
    "config": map[string]interface{}{
        "wide_screen_mode": true,
    },
    "header": map[string]interface{}{
        "title": map[string]interface{}{
            "tag":     "plain_text",
            "content": "系统通知",
        },
        "template": "blue",
    },
    "elements": []map[string]interface{}{
        {
            "tag":     "markdown",
            "content": "您有一条新消息\n\n**发送人**：张三\n**时间**：2024-01-01 10:00",
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
                    "url":  "https://example.com",
                },
            },
        },
        {
            "tag": "note",
            "elements": []map[string]interface{}{
                {
                    "tag":     "plain_text",
                    "content": "点击按钮查看完整内容",
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

### 使用卡片构建工具

推荐使用飞书提供的[卡片构建工具](https://open.feishu.cn/cardkit)来设计卡片，然后复制 JSON 代码。

## 消息类型选择建议

| 场景 | 推荐类型 | 原因 |
|------|---------|------|
| 简单通知 | Text | 轻量、简洁 |
| 需要格式化的文档 | Markdown/Post | 支持丰富的 Markdown 语法 |
| 需要精确控制格式 | Post (原始) | 完全控制富文本结构 |
| 需要用户交互 | Card | 支持按钮、表单等交互元素 |
| 需要精美排版 | Card | 提供更灵活的布局控制 |
| 分享群聊或用户 | ShareChat/ShareUser | 专用名片格式 |
| 发送媒体文件 | Image/Audio/Media/File | 媒体资源 |

## 4. 图片消息 (Image)

发送图片需要先通过[上传图片接口](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/image/create)获取 `image_key`。

```go
// 构造图片消息内容
content := map[string]string{
    "image_key": "img_7ea74629-9191-4176-998c-2e603c9c5e8g",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeImage,
    Content: string(contentJSON),
}
```

## 5. 原始富文本消息 (Post)

如果需要精确控制富文本格式，可以使用 `MessageTypePost` 提供完整的 post JSON 结构。

### 支持的标签

- `text`: 文本标签，支持样式（bold, underline, lineThrough, italic）
- `a`: 超链接标签
- `at`: @提到用户
- `img`: 图片标签
- `media`: 视频标签
- `emotion`: 表情标签
- `code_block`: 代码块标签
- `hr`: 分隔线标签
- `md`: Markdown 标签

### 示例

```go
postContent := map[string]interface{}{
    "zh_cn": map[string]interface{}{
        "title": "项目通知",
        "content": [][]map[string]interface{}{
            { // 第一段
                {
                    "tag":   "text",
                    "text":  "第一行：",
                    "style": []string{"bold", "underline"},
                },
                {
                    "tag":   "a",
                    "href":  "https://open.feishu.cn",
                    "text":  "飞书开放平台",
                    "style": []string{"bold", "italic"},
                },
            },
            { // 第二段
                {
                    "tag":  "text",
                    "text": "第二行文本",
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
                    "language": "GO",
                    "text":     "func main() {\n    fmt.Println(\"Hello\")\n}",
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

## 6. 群名片消息 (ShareChat)

分享群聊名片。机器人必须在要分享的群中。

```go
content := map[string]string{
    "chat_id": "oc_0dd200d32fda15216d2c2ef1ddb32f76",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeShareChat,
    Content: string(contentJSON),
}
```

## 7. 用户名片消息 (ShareUser)

分享用户名片。`user_id` 只支持设置用户的 `open_id`，且用户需要在机器人的可用范围内。

```go
content := map[string]string{
    "user_id": "ou_0dd200d32fda15216d2c2ef1ddb32f76",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeShareUser,
    Content: string(contentJSON),
}
```

## 8. 音频消息 (Audio)

发送音频需要先通过[上传文件接口](https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/im-v1/file/create)获取 `file_key`。

```go
content := map[string]string{
    "file_key": "75235e0c-4f92-430a-a99b-8446610223cg",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeAudio,
    Content: string(contentJSON),
}
```

## 9. 视频消息 (Media)

发送视频需要先上传视频文件（mp4 格式）获取 `file_key`，可选配置视频封面 `image_key`。

```go
content := map[string]string{
    "file_key":  "file_v2_0dcdd7d9-fib0-4432-a519-41d25aca542j",
    "image_key": "img_7ea74629-9191-4176-998c-2e603c9c5e8g", // 可选
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeMedia,
    Content: string(contentJSON),
}
```

## 10. 文件消息 (File)

发送文件需要先通过上传文件接口获取 `file_key`。

```go
content := map[string]string{
    "file_key": "75235e0c-4f92-430a-a99b-8446610223cg",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeFile,
    Content: string(contentJSON),
}
```

## 11. 表情包消息 (Sticker)

目前仅支持发送机器人接收到的表情包，通过接收消息事件获取 `file_key`。

```go
content := map[string]string{
    "file_key": "75235e0c-4f92-430a-a99b-8446610223cg",
}
contentJSON, _ := json.Marshal(content)

msg := &types.Message{
    Type:    types.MessageTypeSticker,
    Content: string(contentJSON),
}
```

## 12. 系统消息 (System)

系统消息用于在单聊中显示分割线。需要使用 `tenant_access_token` 并具有相应权限。

```go
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
        "need_rollup": true, // 是否滚动清屏
    },
}
contentJSON, _ := json.Marshal(systemContent)

msg := &types.Message{
    Type:    types.MessageTypeSystem,
    Content: string(contentJSON),
}
```

**注意事项**：
- 仅支持在单聊（p2p）中使用
- 需要飞书客户端 V7.20 或以上版本
- 文本长度不能超过 20 个字符或 10 个汉字

## 消息类型选择建议

## API 参考

完整的飞书消息 API 文档：
- [发送消息内容结构](https://open.feishu.cn/document/server-docs/im-v1/message-content-description/create_json)
- [飞书卡片设计指南](https://open.feishu.cn/document/uAjLw4CM/ukzMukzMukzM/feishu-cards/card-json-v2-structure)

## 测试示例

查看 `lark_message_types_test.go` 文件获取更多测试示例。

运行测试：
```bash
# 测试文本消息格式
go test -v -run TestLarkTextMessageFormats

# 测试富文本消息格式
go test -v -run TestLarkPostMessageFormats

# 测试卡片消息格式
go test -v -run TestLarkCardMessageFormats
```
