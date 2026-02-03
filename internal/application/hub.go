package application

import (
	"log"
	"sync"
	"web_websocket/internal/domain"
)

// Hub 管理所有 WebSocket 连接和订阅关系
type Hub struct {
	// 所有在线客户端
	clients map[*Client]bool

	// ProjectID 到 Client 的倒排索引（核心！）
	// projects["project_01"] = {client1, client2, ...}
	projects map[string]map[*Client]bool

	// UserID 到 Client 的映射（同一账号只保留最新连接）
	// userConnections["user_001"] = client
	userConnections map[string]*Client

	// 广播通道
	broadcast chan *domain.PushMessage

	// 注册通道
	register chan *Client

	// 注销通道
	unregister chan *Client

	// 读写锁保护并发访问
	mu sync.RWMutex
}

// NewHub 创建新的 Hub 实例
func NewHub() *Hub {
	return &Hub{
		clients:           make(map[*Client]bool),
		projects:          make(map[string]map[*Client]bool),
		userConnections:   make(map[string]*Client),
		broadcast:         make(chan *domain.PushMessage, 256),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
	}
}

// Run 启动 Hub 主循环（必须在 goroutine 中运行）
func (h *Hub) Run() {
	log.Println("[Hub] Starting hub main loop...")
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// handleRegister 处理客户端注册
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 检查是否存在同一 UserID 的旧连接
	if oldClient, exists := h.userConnections[client.UserID]; exists {
		log.Printf("[Hub] Replacing old connection for UserID=%s", client.UserID)
		// 关闭旧连接（发送到注销通道）
		go func(c *Client) {
			h.unregister <- c
		}(oldClient)
	}

	// 注册新连接
	h.clients[client] = true
	h.userConnections[client.UserID] = client
	log.Printf("[Hub] Client registered: UserID=%s, Total=%d", client.UserID, len(h.clients))
}

// handleUnregister 处理客户端注销
func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// 从所有订阅的项目中移除
		for projectID := range client.SubscribedProjects {
			if subscribers, exists := h.projects[projectID]; exists {
				delete(subscribers, client)
				// 如果该项目没有订阅者了,删除项目
				if len(subscribers) == 0 {
					delete(h.projects, projectID)
				}
			}
		}

		// 从客户端列表移除
		delete(h.clients, client)

		// 从用户连接映射中移除（只有当这个连接是最新的才移除）
		if h.userConnections[client.UserID] == client {
			delete(h.userConnections, client.UserID)
		}

		// 方案1：Hub 不关闭 client.send，避免并发广播时 send on closed channel
		// 通过关闭 websocket 连接让 read/write pump 自然退出。
		_ = client.conn.Close()

		log.Printf("[Hub] Client unregistered: UserID=%s, Total=%d", client.UserID, len(h.clients))
	}
}

// handleBroadcast 处理消息广播（精准推送）
func (h *Hub) handleBroadcast(message *domain.PushMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	projectID := message.ProjectID
	subscribers, exists := h.projects[projectID]

	if !exists || len(subscribers) == 0 {
		log.Printf("[Hub] No subscribers for project: %s", projectID)
		return
	}

	log.Printf("[Hub] Broadcasting to project '%s': %d subscribers", projectID, len(subscribers))

	// 只推送给订阅了该 Topic 的客户端
	for client := range subscribers {
		select {
		case client.send <- message:
			// 发送成功
		default:
			// 发送失败，客户端可能阻塞，关闭连接
			log.Printf("[Hub] Client send buffer full, closing: UserID=%s", client.UserID)
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// HandleSubscription 处理客户端订阅/取消订阅（线程安全）
func (h *Hub) HandleSubscription(client *Client, action string, projectIDs []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch action {
	case "subscribe":
		for _, projectID := range projectIDs {
			// 初始化项目订阅者集合
			if h.projects[projectID] == nil {
				h.projects[projectID] = make(map[*Client]bool)
			}
			h.projects[projectID][client] = true
			client.SubscribedProjects[projectID] = true
		}
		log.Printf("[Hub] Client subscribed: UserID=%s, ProjectIDs=%v", client.UserID, projectIDs)

	case "unsubscribe":
		for _, projectID := range projectIDs {
			if subscribers, exists := h.projects[projectID]; exists {
				delete(subscribers, client)
				if len(subscribers) == 0 {
					delete(h.projects, projectID)
				}
			}
			delete(client.SubscribedProjects, projectID)
		}
		log.Printf("[Hub] Client unsubscribed: UserID=%s, ProjectIDs=%v", client.UserID, projectIDs)
	}
}

// Broadcast 发送消息到广播通道（供外部调用）
func (h *Hub) Broadcast(message *domain.PushMessage) {
	h.broadcast <- message
}

// Register 注册客户端（供外部调用）
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端（供外部调用）
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
