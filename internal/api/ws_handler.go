package api

import (
	"log"
	"net/http"
	"web_websocket/internal/application"
	jwtParser "web_websocket/internal/infrastructure/jwt"
	"web_websocket/internal/infrastructure/websocket"

	"github.com/gin-gonic/gin"
)

// WSHandler WebSocket 连接处理器
type WSHandler struct {
	hub       *application.Hub
	jwtSecret string
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(hub *application.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{
		hub:       hub,
		jwtSecret: jwtSecret,
	}
}

// HandleWebSocket 处理 WebSocket 握手
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	// 1. 从 URL Query 获取 Token
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// 2. 校验 JWT Token
	claims, err := jwtParser.ParseToken(token, h.jwtSecret)
	if err != nil {
		log.Printf("[WS] JWT validation failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// 3. Upgrade HTTP to WebSocket
	conn, err := websocket.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] Upgrade failed: %v", err)
		return
	}

	// 4. 创建 Client 并注册到 Hub
	client := application.NewClient(h.hub, conn, claims.UserID, claims.Username)
	h.hub.Register(client)

	log.Printf("[WS] Client connected: UserID=%s, Username=%s", claims.UserID, claims.Username)

	// 5. 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}
