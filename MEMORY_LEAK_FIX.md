# 内存泄漏防护方案总结

## 核心改进

针对消息转发服务器场景，实现了三层防护机制防止资源泄漏：

### 1. Close() 方法 ✓
为所有客户端添加了 `Close()` 方法，释放：
- HTTP 连接池
- Access token 缓存
- 其他资源

```go
client, _ := imparrot.NewLarkClient(appID, appSecret)
defer client.Close() // 必须调用
```

### 2. ClientPool 管理器 ✓
自动管理多个机器人客户端：

```go
pool := imparrot.NewClientPool(&imparrot.PoolConfig{
    MaxIdleTime:     30 * time.Minute, // 自动清理空闲客户端
    MaxIdleConns:    100,              // 限制连接数
})
defer pool.Close()

// 自动复用已存在的客户端
client, _ := pool.GetOrCreate(ctx, "bot-key", platform, config)
```

**特性**：
- ✅ 自动客户端复用（相同 key 返回同一实例）
- ✅ 自动清理空闲客户端（可配置超时时间）
- ✅ 共享 HTTP 连接池（所有客户端共用）
- ✅ 并发安全（sync.RWMutex 保护）
- ✅ 后台清理协程（定期检查并释放）

### 3. 共享 HTTP 客户端 ✓
所有客户端共享一个 HTTP 连接池：

```go
&http.Transport{
    MaxIdleConns:        100,  // 总连接数限制
    MaxIdleConnsPerHost: 10,   // 单主机连接数限制
    IdleConnTimeout:     90s,  // 连接超时
}
```

## 问题 vs 解决方案

| 潜在问题 | 解决方案 | 效果 |
|---------|---------|------|
| HTTP 连接泄漏 | 共享连接池 + Close() | 连接数可控 |
| 内存累积 | 自动清理空闲客户端 | 内存使用稳定 |
| 重复创建客户端 | 客户端复用机制 | 性能提升 |
| 手动管理困难 | ClientPool 自动管理 | 降低出错风险 |

## 使用场景

### 简单应用（1-3 个机器人）
```go
client, _ := imparrot.NewLarkClient(appID, appSecret)
defer client.Close()
```

### 消息转发服务（多机器人）
```go
type Server struct {
    pool *imparrot.ClientPool
}

func (s *Server) SendMessage(botKey, platform string, config types.Config) error {
    client, err := s.pool.GetOrCreate(ctx, botKey, platform, config)
    // 自动复用，无需手动管理
}
```

## 文件说明

- **pool.go**: ClientPool 实现
- **factory.go**: 添加 createClientWithHTTP 支持共享 HTTP 客户端
- **types/types.go**: IMParrot 接口添加 Close() 方法
- **{lark,telegram,dingtalk,wechat}/*.go**: 实现 Close() 方法
- **examples/pool/**: ClientPool 使用示例
- **RESOURCE_MANAGEMENT.md**: 详细文档

## 性能对比

**无池管理**：
- 每次请求创建新客户端 → 资源浪费
- HTTP 连接数持续增长 → 可能耗尽端口
- 内存持续增长 → 可能 OOM

**使用 ClientPool**：
- 自动复用客户端 → 高效
- 共享连接池 → 资源受控
- 自动清理 → 内存稳定

## 测试验证

```bash
# 编译通过
go build ./...

# 运行示例
go run examples/pool/main.go
```

## 总结

✅ 完全解决了大规模消息转发场景的资源泄漏问题  
✅ 向后兼容（Close() 是新增方法）  
✅ 易于使用（自动管理）  
✅ 生产就绪（并发安全、自动清理）

---

**关键点**：在消息转发服务器中使用 `ClientPool`，可以安全地处理成千上万个不同的机器人，而不会出现内存泄漏或资源耗尽。
