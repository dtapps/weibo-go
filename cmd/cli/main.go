package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	weibo "github.com/dtapps/weibo-go"
)

func main() {
	// 解析命令行参数
	mode := flag.String("mode", "cli", "运行模式: cli, hotsearch, search, status")
	appId := flag.String("app-id", "", "微博 App ID")
	appSecret := flag.String("app-secret", "", "微博 App Secret")
	category := flag.String("category", "主榜", "热搜榜单类型")
	query := flag.String("query", "", "搜索关键词")
	configFile := flag.String("config", "", "配置文件路径")
	flag.Parse()

	// 创建客户端
	var client *weibo.Client

	if *configFile != "" {
		// 从配置文件加载
		data, err := os.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("读取配置文件失败: %v", err)
		}
		config, err := weibo.ConfigFromJSON(string(data))
		if err != nil {
			log.Fatalf("解析配置文件失败: %v", err)
		}
		client = weibo.NewClient(config)
	} else if *appId != "" && *appSecret != "" {
		// 使用命令行参数
		client = weibo.NewClientWithAppCredentials(*appId, *appSecret)
	} else {
		// 尝试从环境变量读取
		appIdEnv := os.Getenv("WEIBO_APP_ID")
		appSecretEnv := os.Getenv("WEIBO_APP_SECRET")
		if appIdEnv != "" && appSecretEnv != "" {
			client = weibo.NewClientWithAppCredentials(appIdEnv, appSecretEnv)
		} else {
			fmt.Println("请提供以下参数之一:")
			fmt.Println("  1. -config <配置文件路径>")
			fmt.Println("  2. -app-id <App ID> -app-secret <App Secret>")
			fmt.Println("  3. 环境变量 WEIBO_APP_ID 和 WEIBO_APP_SECRET")
			flag.Usage()
			os.Exit(1)
		}
	}

	// 根据模式执行
	switch *mode {
	case "cli":
		runCLI(client)
	case "hotsearch":
		runHotSearch(client, *category)
	case "search":
		if *query == "" {
			fmt.Println("搜索模式需要提供 -query 参数")
			os.Exit(1)
		}
		runSearch(client, *query)
	case "status":
		runStatus(client)
	default:
		fmt.Printf("未知模式: %s\n", *mode)
		os.Exit(1)
	}
}

func runCLI(client *weibo.Client) {
	fmt.Println("微博 CLI 客户端")
	fmt.Println("输入消息发送，按 Ctrl+C 退出")
	fmt.Println("---")

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 连接
	connected := make(chan bool)
	errChan := make(chan error)

	go func() {
		err := client.Connect(&weibo.ConnectOptions{
			OnOpen: func() {
				fmt.Println("\n已连接到微博服务")
				connected <- true
			},
			OnMessage: func(msg *weibo.InboundMessage) {
				handleMessage(msg)
			},
			OnError: func(err error) {
				fmt.Printf("\n错误: %v\n", err)
				errChan <- err
			},
			OnClose: func(code int, reason string) {
				fmt.Printf("\n连接已关闭: %d - %s\n", code, reason)
			},
		})
		if err != nil {
			errChan <- err
		}
	}()

	// 等待连接或错误
	select {
	case <-connected:
		// 连接成功，显示提示
		fmt.Println("> ")
	case err := <-errChan:
		log.Fatalf("连接失败: %v", err)
	}

	// 等待信号
	<-sigChan
	fmt.Println("\n正在关闭...")
	client.Close()
}

func runHotSearch(client *weibo.Client, category string) {
	fmt.Printf("获取 %s 热搜...\n", category)

	result, err := client.GetHotSearch(category, nil)
	if err != nil {
		log.Fatalf("获取热搜失败: %v", err)
	}

	if !result.Success {
		log.Fatalf("获取热搜失败: %s", result.Error)
	}

	fmt.Printf("\n%s 热搜榜\n", result.Category)
	fmt.Printf("数据来源: %s %s\n", result.CallTime, result.Source)
	fmt.Println("---")

	for _, item := range result.Items {
		fmt.Printf("%2d. %s\n", item.Rank, item.Word)
		fmt.Printf("    热度: %d\n", item.HotValue)
		if item.H5Link != "" {
			fmt.Printf("    链接: %s\n", item.H5Link)
		}
		fmt.Println()
	}

	fmt.Printf("共 %d 条热搜\n", result.Total)
}

func runSearch(client *weibo.Client, query string) {
	fmt.Printf("搜索: %s\n", query)

	result, err := client.Search(query)
	if err != nil {
		log.Fatalf("搜索失败: %v", err)
	}

	if !result.Success {
		log.Fatalf("搜索失败: %s", result.Error)
	}

	fmt.Println("---")
	if result.CallTime != "" {
		fmt.Printf("来源: %s %s\n", result.CallTime, result.Source)
		fmt.Println("---")
	}
	fmt.Println(result.Content)

	if result.ReferenceCount > 0 {
		fmt.Printf("\n(参考来源: %d)\n", result.ReferenceCount)
	}
}

func runStatus(client *weibo.Client) {
	fmt.Println("获取用户微博...")

	result, err := client.GetStatus(nil)
	if err != nil {
		log.Fatalf("获取微博失败: %v", err)
	}

	if !result.Success {
		log.Fatalf("获取微博失败: %s", result.Error)
	}

	fmt.Printf("\n用户微博 (共 %d 条)\n", result.Total)
	fmt.Println("---")

	for i, status := range result.Statuses {
		fmt.Printf("[%d] @%s\n", i+1, status.User.ScreenName)
		fmt.Printf("    %s\n", status.Text)
		fmt.Printf("    转发: %d  评论: %d  点赞: %d\n",
			status.RepostsCount, status.CommentsCount, status.AttitudesCount)
		fmt.Printf("    时间: %s\n", status.CreatedAt)
		if status.HasImage && len(status.Images) > 0 {
			fmt.Printf("    图片: %d 张\n", len(status.Images))
		}
		fmt.Println()
	}
}

func handleMessage(msg *weibo.InboundMessage) {
	fmt.Printf("\n[收到消息] from=%s\n", msg.Payload.FromUserId)

	var text string
	for _, item := range msg.Payload.Input {
		if item.Type == "message" && item.Role == "user" {
			for _, part := range item.Content {
				if part.Type == "input_text" && part.Text != nil {
					text = *part.Text
					break
				}
			}
		}
	}

	if text == "" && msg.Payload.Text != nil {
		text = *msg.Payload.Text
	}

	if text != "" {
		fmt.Printf("内容: %s\n", text)
	}

	// 显示附件信息
	for _, item := range msg.Payload.Input {
		if item.Type == "message" && item.Role == "user" {
			for _, part := range item.Content {
				if part.Type == "input_image" && part.Source != nil {
					fmt.Printf("[图片] %s (%s)\n", derefString(part.Filename), part.Source.MediaType)
				}
				if part.Type == "input_file" && part.Source != nil {
					fmt.Printf("[文件] %s (%s)\n", derefString(part.Filename), part.Source.MediaType)
				}
			}
		}
	}

	fmt.Print("> ")
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// 打印 JSON 辅助函数
func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("JSON 格式化失败: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
