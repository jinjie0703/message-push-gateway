package application

import (
	"encoding/json"
	"log"
	"time"
	"web_websocket/internal/domain"

	"github.com/gorilla/websocket"
)

const (
	// 写入超时时间
	writeWait = 10 * time.Second

	// Pong 等待时间（必须大于 pingPeriod）
	pongWait = 60 * time.Second

	// Ping 发送周期（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// Client WebSocket 客户端
type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	send           chan *domain.PushMessage
	messageService *MessageService

	// 用户信息
	UserID   string
	Username string

	// 订阅的项目ID
	SubscribedProjects map[string]bool
}

// NewClient 创建新的客户端实例
func NewClient(hub *Hub, conn *websocket.Conn, userID, username string) *Client {
	return &Client{
		hub:              hub,
		conn:             conn,
		send:             make(chan *domain.PushMessage, 256),
		messageService:   NewMessageService(hub),
		UserID:           userID,
		Username:         username,
		SubscribedProjects: make(map[string]bool),
	}
}

// ReadPump 从 WebSocket 读取消息（处理订阅指令）
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
		log.Printf("[Client] ReadPump closed: UserID=%s", c.UserID)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Client] ReadPump error: UserID=%s, Err=%v", c.UserID, err)
			}
			break
		}

		// 打印收到的原始消息（调试用）
		log.Printf("[Client] 收到原始消息: UserID=%s, Message=%s", c.UserID, string(message))

		// 解析订阅指令
		var cmd domain.SubscribeCommand
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("[Client] JSON解析错误: UserID=%s, Err=%v", c.UserID, err)
			continue
		}

		// 使用 MessageService 处理订阅/取消订阅（带验证）
		var serviceErr error
		if cmd.Action == "subscribe" {
			serviceErr = c.messageService.SubscribeProjects(c, cmd.ProjectIDs)
		} else if cmd.Action == "unsubscribe" {
			serviceErr = c.messageService.UnsubscribeProjects(c, cmd.ProjectIDs)
		}

		if serviceErr != nil {
			log.Printf("[Client] Subscription error: UserID=%s, Err=%v", c.UserID, serviceErr)
		}
	}
}

// WritePump 向 WebSocket 写入消息（推送报警）
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Printf("[Client] WritePump closed: UserID=%s", c.UserID)
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了 send 通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 发送 JSON 消息
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("[Client] WriteJSON error: UserID=%s, Err=%v", c.UserID, err)
				return
			}

		case <-ticker.C:
			// 发送 Ping 保活
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[Client] Ping error: UserID=%s, Err=%v", c.UserID, err)
				return
			}
		}
	}
}
