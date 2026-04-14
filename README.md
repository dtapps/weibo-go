# Weibo Go SDK

微博 Go 语言 SDK，支持 WebSocket 实时消息收发。

参考 [微博私信通道插件](https://github.com/wecode-ai/openclaw-weibo) 实现

## 功能特性

- WebSocket 实时消息收发
- 自动心跳保活和重连

## 安装

```bash
go get github.com/dtapps/weibo-go
```

## 快速开始

```go
package main

import (
	"log"

	weibo "github.com/dtapps/weibo-go"
	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
)

func main() {

	// 日志
	logger.SetLevel(logger.LevelDebug)
	l := logger.GetLogger("test")

	// 创建配置
	cfg := &types.Config{
		Weibo: &types.WeiboConfig{
			AppId:     "your-app-id",
			AppSecret: "your-app-secret",
		},
	}

	// 创建客户端
	client, err := weibo.NewClient("default", cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// 设置消息处理
	client.OnMessage(func(msg *types.WsMessageMsg) {
		l.Info("收到消息",
			logger.F("FromUserId", msg.Payload.FromUserId),
			logger.F("Text", msg.Payload.Text),
		)
	})

	// 连接状态
	client.OnConnected(func() {
		l.Info("已连接")
	})

	client.OnDisconnected(func() {
		l.Info("已断开")
	})

	l.Info("正在连接...")

	select {}
}
```

## 配置

| 参数 | 必填 | 说明 |
|------|------|------|
| AppId | 是 | 应用ID |
| AppSecret | 是 | 应用密钥 |
| WSEndpoint | 否 | WebSocket 地址 |
| TokenEndpoint | 否 | Token 地址 |

## API

```go
// 创建客户端
client, err := weibo.NewClient("default", cfg)

// 设置消息处理
client.OnMessage(func(msg *types.WsMessageMsg) {})

// 连接成功
client.OnConnected(func() {})

// 连接断开
client.OnDisconnected(func() {})

// 发送消息
client.SendMessage(toUserId, "你好")

// 获取状态
client.GetState()

// 停止客户端
client.Stop()
```

## 日志

```go
logger.SetLevel(logger.LevelDebug)
l := logger.GetLogger("test")

l.Info("信息", logger.F("key", "value"))
l.Error("错误", logger.F("error", err.Error()))
```

## License

MIT
