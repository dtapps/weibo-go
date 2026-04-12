# Weibo Go SDK

微博 Go 语言 SDK，支持 WebSocket 连接、热搜查询、搜索和消息收发等功能。

参考 [微博私信通道插件](https://github.com/wecode-ai/openclaw-weibo) 实现

## 功能特性

- **WebSocket 客户端**: 支持微博实时消息推送和发送
- **自动重连**: 内置指数退避重连机制
- **心跳检测**: 自动心跳保活连接
- **Token 管理**: 自动获取和管理 API Token
- **热搜 API**: 获取微博热搜榜（主榜、文娱榜、社会榜等）
- **搜索 API**: 微博智搜功能
- **用户微博 API**: 获取用户发布的微博
- **消息分片**: 支持长消息自动分片发送

## 安装

```bash
go get github.com/dtapps/weibo-go
```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    weibo "github.com/dtapps/weibo-go"
)

func main() {
    // 创建客户端
    client := weibo.NewClientWithAppCredentials(
        "your-app-id",
        "your-app-secret",
    )

    // 连接
    err := client.Connect(&weibo.ConnectOptions{
        OnMessage: func(msg *weibo.InboundMessage) {
            fmt.Printf("收到消息 from=%s\n", msg.Payload.FromUserId)
        },
        OnOpen: func() {
            fmt.Println("已连接到微博服务")
        },
    })

    if err != nil {
        log.Fatal(err)
    }

    // 阻塞主线程
    select {}
}
```

### 获取热搜

```go
// 创建客户端
client := weibo.NewClientWithAppCredentials(appId, appSecret)

// 获取主榜热搜
result, err := client.GetHotSearch("主榜", nil)
if err != nil {
    log.Fatal(err)
}

if result.Success {
    fmt.Printf("热搜来源: %s %s\n", result.Category, result.CallTime)
    for _, item := range result.Items {
        fmt.Printf("%d. %s (热度: %d)\n", item.Rank, item.Word, item.HotValue)
    }
}
```

支持的榜单类型：
- `主榜`
- `文娱榜`
- `社会榜`
- `生活榜`
- `acg榜`
- `科技榜`
- `体育榜`

### 搜索微博

```go
result, err := client.Search("人工智能")
if err != nil {
    log.Fatal(err)
}

if result.Success {
    fmt.Printf("来源: %s %s\n", result.CallTime, result.Source)
    fmt.Println(result.Content)
}
```

### 获取用户微博

```go
result, err := client.GetStatus(nil) // nil 表示使用默认数量
if err != nil {
    log.Fatal(err)
}

if result.Success {
    for _, status := range result.Statuses {
        fmt.Printf("@%s: %s\n", status.User.ScreenName, status.Text)
    }
}
```

### 发送消息

```go
// 连接后才能发送
err := client.Connect(nil)
if err != nil {
    log.Fatal(err)
}

// 发送消息
result, err := client.Send("user123456", "你好，这是一条测试消息")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("消息已发送: %s\n", result.MessageId)
```

### 长消息分片发送

```go
// 自动按 2000 字符分片发送
results, err := client.SendChunked("user123456", longText, 2000)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("已发送 %d 个分片\n", len(results))
```

## 配置选项

### 从 JSON 创建配置

```go
configJSON := `{
    "enabled": true,
    "appId": "your-app-id",
    "appSecret": "your-app-secret",
    "wsEndpoint": "ws://open-im.api.weibo.com/ws/stream",
    "tokenEndpoint": "http://open-im.api.weibo.com/open/auth/ws_token",
    "dmPolicy": "open",
    "chunkMode": "raw"
}`

config, err := weibo.ConfigFromJSON(configJSON)
client := weibo.NewClient(config)
```

### 多账号配置

```go
config := &weibo.WeiboConfig{
    Enabled: boolPtr(true),
    AppId: strPtr("default-app-id"),
    AppSecret: strPtr("default-app-secret"),
    Accounts: map[string]interface{}{
        "account1": map[string]interface{}{
            "appId": "app-id-1",
            "appSecret": "app-secret-1",
        },
        "account2": map[string]interface{}{
            "appId": "app-id-2",
            "appSecret": "app-secret-2",
        },
    },
}

client := weibo.NewClient(config)
```

## API 参考

### Client

| 方法 | 描述 |
|------|------|
| `NewClient(config)` | 使用配置创建客户端 |
| `NewClientWithAppCredentials(appId, appSecret)` | 使用 App 凭证创建客户端 |
| `Connect(opts)` | 连接到微博 WebSocket 服务 |
| `Send(to, text)` | 发送消息 |
| `SendChunked(to, text, chunkLimit)` | 分片发送长消息 |
| `GetHotSearch(category, count)` | 获取热搜榜 |
| `Search(query)` | 搜索微博 |
| `GetStatus(count)` | 获取用户微博 |
| `GetToken()` | 获取当前 Token |
| `Close()` | 关闭客户端 |

### ConnectOptions

```go
type ConnectOptions struct {
    AccountId *string                    // 指定账号 ID
    OnMessage func(*InboundMessage)      // 消息回调
    OnError   func(error)                // 错误回调
    OnClose   func(code int, reason string) // 关闭回调
    OnOpen    func()                     // 连接成功回调
    OnStatus  func(*RuntimeStatusPatch)  // 状态更新回调
}
```

### InboundMessage

```go
type InboundMessage struct {
    Type    string
    Payload InboundMessagePayload
}

type InboundMessagePayload struct {
    MessageId  string
    FromUserId string
    Text       *string
    Timestamp  *int64
    Input      []InboundMessageInputItem
}
```

### RuntimeStatusPatch

```go
type RuntimeStatusPatch struct {
    Running           *bool
    Connected         *bool
    ConnectionState   *ConnectionState
    ReconnectAttempts *int
    NextRetryAt       *int64
    LastConnectedAt   *int64
    LastDisconnect    *DisconnectInfo
    LastError         *string
}
```

### HotSearchResult

```go
type HotSearchResult struct {
    Success   bool
    Category  string
    Total     int
    CallTime  string
    Source    string
    Items     []HotSearchItem
    Error     string
}

type HotSearchItem struct {
    Rank     int
    Word     string
    HotValue int
    Category string
    AppLink  string
    H5Link   string
}
```

## 消息分片模式

SDK 支持三种消息分片模式：

| 模式 | 描述 |
|------|------|
| `raw` | 不分片，直接发送原始文本 |
| `length` | 按字符数分片 |
| `newline` | 按换行符分片 |

## 错误处理

```go
result, err := client.GetHotSearch("主榜", nil)
if err != nil {
    // 网络或系统错误
    log.Printf("系统错误: %v", err)
    return
}

if !result.Success {
    // API 返回的错误
    log.Printf("API 错误: %s", result.Error)
    return
}

// 成功处理
fmt.Println(result.Items)
```

## License

MIT License
