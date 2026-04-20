# Weibo Go SDK

微博 Go 语言 SDK，支持 WebSocket 实时消息收发。

参考 [微博私信通道插件](https://github.com/wecode-ai/openclaw-weibo) 实现

## 安装

```bash
go get github.com/dtapps/weibo-go
```

## 获取凭证

1. 打开微博客户端，私信 **@微博龙虾助手**
2. 发送消息：`连接龙虾`
3. 收到回复示例：

```
您的应用凭证信息如下：

AppId: your-app-id
AppSecret: your-app-secret
```

如需重置凭证，发送 `重置凭证`

## 使用 Demo

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	weibo "github.com/dtapps/weibo-go"
	"github.com/dtapps/weibo-go/config"
	"github.com/dtapps/weibo-go/logger"
	"github.com/dtapps/weibo-go/types"
)

func main() {
	// 日志
	logger.SetLevel(logger.LevelDebug)
	l := logger.GetLogger("demo")

	// 创建配置
	defaultCfg := config.DefaultConfig()
	defaultCfg.AppID = "your-app-id"
	defaultCfg.AppSecret = "your-app-secret"

	// 创建客户端
	client, err := weibo.NewClient("default", &types.Config{
		Weibo: defaultCfg,
	})
	if err != nil {
		panic(err)
	}
	defer client.Stop()

	// 设置消息处理
	client.OnMessage(func(msg *types.InboundMessage) {
		fmt.Println("----------------")
		jsonMsg, _ := json.MarshalIndent(msg, "", "  ")
		fmt.Println(string(jsonMsg))
		fmt.Println("----------------")

		content := ""
		for _, segment := range msg.Content {
			content += segment.Text
		}

		// 自动回复
		reply := fmt.Sprintf("收到: %s", content)
		messageID, err := client.SendMessage(&types.OutboundMessage{
			ToUserID: msg.SenderID,
			Text:     reply,
		})
		if err != nil {
			l.Error("发送消息失败", logger.F("error", err.Error()))
		}
		l.Info("已回复", logger.F("reply", reply), logger.F("messageID", messageID))
	})

	// 设置连接状态
	client.OnConnected(func() {
		l.Info("已连接到微博服务器")

		// 列出所有成员
		members := client.GetMember().ListUsers(&types.MemberListUsersRequest{})
		l.Info("列出所有成员", logger.F("members", members))

		// 获取缓存的Token
		token := client.GetTokenManager().GetCachedToken()
		l.Info("获取缓存的Token",
			logger.F("token", token.Token),
			logger.F("acquiredAt", time.Unix(token.AcquiredAt, 0).Format(time.DateTime)),
			logger.F("expiresAt", time.Unix(token.ExpiresAt, 0).Format(time.DateTime)),
		)
	})

	client.OnDisconnected(func() {
		l.Info("已断开连接")
	})

	l.Info("正在连接...")

	select {}
}
```

## 配置

| 参数 | 必填 | 说明 |
|------|------|------|
| AppID | 是 | 应用ID |
| AppSecret | 是 | 应用密钥 |
| TokenEndpoint | 否 | Token 端点 |
| WSEndpoint | 否 | WebSocket 端点 |

## License

MIT
