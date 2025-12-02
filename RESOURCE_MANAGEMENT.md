# 资源管理和内存泄漏防护

## 问题背景

当 IM-Parrot 被用于消息转发服务器时，可能需要管理大量不同的机器人客户端。如果不正确管理资源，可能会导致：

1. **HTTP 连接泄漏**: 每个客户端创建独立的 HTTP 连接池
2. **内存泄漏**: 客户端对象和相关数据无法被垃圾回收
3. **Token 缓存累积**: Lark 和 WeChat 的 access token 占用内存
4. **goroutine 泄漏**: 后台任务没有正确关闭

## 解决方案

### 1. Close() 方法

所有客户端都实现了 `Close()` 方法来释放资源：

```go
client, err := imparrot.NewLarkClient(appID, appSecret)
if err != nil {
    log.Fatal(err)
}
defer client.Close() // 重要：始终关闭客户端
```

`Close()` 方法会：
- 关闭空闲的 HTTP 连接
- 清理 access token 缓存
- 标记客户端为已关闭状态

### 2. ClientPool - 自动资源管理

对于需要管理多个机器人的场景，使用 `ClientPool`：

```go
// 创建客户端池
pool := imparrot.NewClientPool(&imparrot.PoolConfig{
    MaxIdleTime:         30 * time.Minute, // 空闲30分钟后自动关闭
    CleanupInterval:     5 * time.Minute,  // 每5分钟检查一次
    HTTPTimeout:         30 * time.Second,
    MaxIdleConns:        100,  // 最大空闲连接数
    MaxIdleConnsPerHost: 10,   // 每个主机的最大空闲连接数
})
defer pool.Close() // 关闭所有客户端

// 获取或创建客户端（自动复用）
client, err := pool.GetOrCreate(ctx, "bot-key", platform, config)
```

### 3. 共享 HTTP 客户端

`ClientPool` 使用单个共享的 HTTP 客户端：

```go
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100, // 所有主机共享
        MaxIdleConnsPerHost: 10,  // 每个主机
        IdleConnTimeout:     90 * time.Second,
    },
}
```

**好处**：
- 多个机器人共享连接池
- 防止连接数爆炸
- 更好的资源利用

## 使用场景对比

### 场景 1: 单个机器人（简单应用）

```go
func main() {
    client, err := imparrot.NewLarkClient(appID, appSecret)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close() // 重要！
    
    // 使用客户端...
}
```

**特点**：
- 简单直接
- 手动管理生命周期
- 适合单机器人或少量机器人

### 场景 2: 消息转发服务器（大规模）

```go
type MessageServer struct {
    pool *imparrot.ClientPool
}

func NewMessageServer() *MessageServer {
    return &MessageServer{
        pool: imparrot.NewClientPool(nil), // 使用默认配置
    }
}

func (s *MessageServer) SendMessage(ctx context.Context, botKey string, platform string, config types.Config, msg *types.Message, target string) error {
    // 自动获取或创建客户端
    client, err := s.pool.GetOrCreate(ctx, botKey, platform, config)
    if err != nil {
        return err
    }
    
    return client.SendPrivateMessage(ctx, target, msg)
}

func (s *MessageServer) Close() error {
    return s.pool.Close() // 关闭所有客户端
}
```

**特点**：
- 自动客户端复用
- 自动清理空闲客户端
- 共享 HTTP 连接池
- 防止内存泄漏

## 最佳实践

### ✅ 推荐做法

```go
// 1. 总是使用 defer Close()
client, _ := imparrot.NewLarkClient(appID, appSecret)
defer client.Close()

// 2. 长时间运行的服务使用 ClientPool
pool := imparrot.NewClientPool(nil)
defer pool.Close()

// 3. 使用上下文控制超时
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
client.SendMessage(ctx, msg, opts)

// 4. 检查错误
if err := client.Close(); err != nil {
    log.Printf("Failed to close client: %v", err)
}
```

### ❌ 避免的做法

```go
// 1. 忘记关闭客户端 - 导致资源泄漏
client, _ := imparrot.NewLarkClient(appID, appSecret)
// 没有 defer client.Close() ❌

// 2. 每次请求都创建新客户端 - 浪费资源
for _, req := range requests {
    client, _ := imparrot.NewLarkClient(appID, appSecret) // ❌
    client.SendMessage(ctx, msg, opts)
    // 即使有 Close()，也很低效
}

// 3. 全局客户端没有关闭机制
var globalClient types.IMParrot

func init() {
    globalClient, _ = imparrot.NewLarkClient(appID, appSecret) // ❌
    // 无法在程序结束时关闭
}
```

## ClientPool 自动清理机制

```
时间线：
0m    - 创建客户端 A, B, C
5m    - 使用客户端 A
10m   - 清理检查：无空闲超过30分钟的客户端
15m   - 清理检查：无空闲超过30分钟的客户端
20m   - 清理检查：无空闲超过30分钟的客户端
25m   - 清理检查：无空闲超过30分钟的客户端
30m   - 清理检查：无空闲超过30分钟的客户端
35m   - 清理检查：客户端 B, C 空闲超过30分钟，自动关闭 ✓
```

## 监控和调试

```go
// 查看池中客户端数量
fmt.Printf("Pool size: %d\n", pool.Size())

// 手动移除不需要的客户端
if err := pool.Remove("bot-key"); err != nil {
    log.Printf("Failed to remove: %v", err)
}

// 立即关闭所有客户端
pool.Close()
```

## 性能对比

### 无池管理（每次创建）

```
请求 1: 创建客户端 A (建立连接)
请求 2: 创建客户端 A (建立连接) ❌ 浪费
请求 3: 创建客户端 A (建立连接) ❌ 浪费
...
内存使用: 持续增长 ❌
连接数: 持续增长 ❌
```

### 使用 ClientPool

```
请求 1: 创建客户端 A (建立连接)
请求 2: 复用客户端 A ✓
请求 3: 复用客户端 A ✓
...
30分钟无请求后: 自动关闭客户端 A ✓
内存使用: 稳定 ✓
连接数: 受限且复用 ✓
```

## 总结

1. **单机器人应用**: 直接创建客户端 + `defer Close()`
2. **多机器人服务**: 使用 `ClientPool` 自动管理
3. **始终关闭**: 无论哪种方式，都要确保资源被释放
4. **监控**: 在生产环境中监控客户端数量和内存使用

通过这些机制，IM-Parrot 可以安全地用于大规模消息转发服务，而不会出现资源泄漏。
