package api

import (
	"web_websocket/internal/application"
	"web_websocket/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter(hub *application.Hub, jwtSecret string) *gin.Engine {
	router := gin.Default()

	// CORS 中间件
	router.Use(middleware.CORS())

	// 创建处理器
	httpHandler := NewHTTPHandler(hub)
	wsHandler := NewWSHandler(hub, jwtSecret)

	// API 路由组
	api := router.Group("/api")
	{
		// HTTP 接口：接收推送消息 Webhook
		api.POST("/push", httpHandler.HandlePushMessage)

		// 自导自演：生成一条测试推送（会按 project_id 精准推给已订阅的 WS 客户端）
		api.GET("/test/push", httpHandler.HandleTestPush)

		// 健康检查
		api.GET("/health", httpHandler.HandleHealth)

		// 系统统计
		api.GET("/stats", httpHandler.HandleStats)

		// WebSocket 路由（支持多种路径）
		api.GET("/ws/notifications", wsHandler.HandleWebSocket)
	}

	// 兼容旧路径
	router.GET("/ws", wsHandler.HandleWebSocket)

	return router
}
