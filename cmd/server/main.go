package main

import (
	"fmt"
	"log"
	"web_websocket/internal/api"
	"web_websocket/internal/application"
	"web_websocket/internal/config"
)

func main() {
	// 1. 加载配置 (从文件、环境变量加载，并应用默认值)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}

	// 2. 初始化 Hub (WebSocket 核心消息中心)
	hub := application.NewHub()

	// 3. 启动 Hub 主循环 (必须在 goroutine 中运行)
	go hub.Run()

	// 4. 配置路由 (HTTP & WebSocket)
	router := api.SetupRouter(hub, cfg.JWT.Secret)

	// 5. 启动 HTTP Server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	wsURL := fmt.Sprintf("%s%s?token=YOUR_JWT_TOKEN", cfg.Endpoints.PublicBaseURL, cfg.Endpoints.WSPath)
	pushURL := fmt.Sprintf("%s%s", cfg.Endpoints.PublicBaseURL, cfg.Endpoints.AlarmPushPath)

	log.Printf("服务启动，监听地址: %s", addr)
	log.Printf("WebSocket 入口: %s", wsURL)
	log.Printf("告警推送 Webhook: POST %s", pushURL)

	if err := router.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
