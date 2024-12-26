# Redis

## 为什么每次使用redis都要传入一个context？

在Redis操作中使用context（上下文）有以下几个重要原因：

#### 1. 超时控制

context可以设置操作的超时时间，防止某个操作因为网络问题或Redis服务器问题而无限期等待。

```go
// 设置一个5秒超时的context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 如果操作超过5秒还没完成，将会返回超时错误
val, err := rdb.Get(ctx, "key").Result()
```

#### 2. 取消操作

context可以用于取消正在进行的Redis操作，比如在处理请求时，如果用户取消请求，可以立即停止正在进行的Redis操作。

```go
// 创建一个可取消的context
ctx, cancel := context.WithCancel(context.Background())

go func() {
    // 在某些条件下取消操作
    time.Sleep(time.Second)
    cancel()
}()

// 如果context被取消，这个操作会立即返回
err := rdb.Set(ctx, "key", "value", 0).Err()
```

#### 3. 请求追踪

context可以用于追踪请求，比如在处理请求时，可以记录请求的开始时间、结束时间、请求的ID等信息。

```go
// 创建带有追踪ID的context
ctx := context.WithValue(context.Background(), "trace_id", "abc-123")

// 在分布式系统中可以通过trace_id追踪请求链路
rdb.Get(ctx, "key")
```

#### 4. 资源管理

context可以用于管理Redis连接池，比如在处理请求时，可以记录请求的开始时间、结束时间、请求的ID等信息。

```go
// 当主程序退出时，确保Redis操作也能优雅退出
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 如果主程序退出，所有使用这个context的Redis操作都会收到通知
rdb.Subscribe(ctx, "channel")
```

#### 5. 并发控制

context可以用于控制并发操作，比如在处理请求时，可以控制并发操作的数量，防止过多的并发操作导致Redis服务器崩溃。

```go
// 使用context控制多个并发操作
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    // 操作1
    rdb.Get(ctx, "key1")
}()

go func() {
    // 操作2
    rdb.Get(ctx, "key2")
}()

// 调用cancel()可以同时取消所有操作
```

#### 实际使用

```go
// 在Web服务中，通常从请求中获取context
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // 获取请求的context
    
    // 使用请求的context进行Redis操作
    // 如果客户端断开连接，Redis操作也会被取消
    val, err := global.GVA_REDIS.Get(ctx, "key").Result()
    if err != nil {
        // 处理错误
    }
}

// 在后台任务中
func BackgroundTask() {
    // 创建带有超时的context
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // 如果操作超时，会自动取消
    err := global.GVA_REDIS.Set(ctx, "key", "value", time.Hour).Err()
    if err != nil {
        // 处理错误
    }
}
```

