package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader WebSocket 升级器配置
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许任何 Origin 的网页来连你的 WS（不做来源限制）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
