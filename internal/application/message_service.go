package application

import (
	"errors"
	"log"
	"strings"
	"web_websocket/internal/domain"
)

// MessageService 消息服务（业务逻辑层）
type MessageService struct {
	hub *Hub
}

// NewMessageService 创建消息服务实例
func NewMessageService(hub *Hub) *MessageService {
	return &MessageService{hub: hub}
}

// PublishMessage 发布推送消息（带验证）
func (s *MessageService) PublishMessage(message *domain.PushMessage) error {
	// 验证消息
	if err := s.validateMessage(message); err != nil {
		log.Printf("[MessageService] Invalid message: %v", err)
		return err
	}

	// 发送到 Hub
	s.hub.Broadcast(message)
	
	log.Printf("[MessageService] Message published: ProjectID=%s, Type=%s", message.ProjectID, message.Type)
	return nil
}

// SubscribeProjects 订阅项目（带验证）
func (s *MessageService) SubscribeProjects(client *Client, projectIDs []string) error {
	// 验证项目 ID
	if err := s.validateProjectIDs(projectIDs); err != nil {
		log.Printf("[MessageService] Invalid project IDs: %v", err)
		return err
	}

	// 调用 Hub 处理订阅
	s.hub.HandleSubscription(client, "subscribe", projectIDs)
	
	return nil
}

// UnsubscribeProjects 取消订阅项目
func (s *MessageService) UnsubscribeProjects(client *Client, projectIDs []string) error {
	// 验证项目 ID
	if err := s.validateProjectIDs(projectIDs); err != nil {
		log.Printf("[MessageService] Invalid project IDs: %v", err)
		return err
	}

	// 调用 Hub 处理取消订阅
	s.hub.HandleSubscription(client, "unsubscribe", projectIDs)
	
	return nil
}

// GetClientStats 获取客户端统计信息
func (s *MessageService) GetClientStats() map[string]interface{} {
	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	// 统计订阅信息
	projectStats := make(map[string]int)
	for projectID, subscribers := range s.hub.projects {
		projectStats[projectID] = len(subscribers)
	}

	return map[string]interface{}{
		"total_clients":      len(s.hub.clients),
		"total_projects":     len(s.hub.projects),
		"project_subscribers":  projectStats,
	}
}

// validateMessage 验证消息格式
func (s *MessageService) validateMessage(message *domain.PushMessage) error {
	if message == nil {
		return errors.New("message is nil")
	}

	if strings.TrimSpace(message.ProjectID) == "" {
		return errors.New("project_id is required")
	}

	if strings.TrimSpace(message.Type) == "" {
		return errors.New("type is required")
	}

	if len(message.Data) == 0 {
		return errors.New("data is required")
	}

	return nil
}

// validateProjectIDs 验证项目 ID 列表
func (s *MessageService) validateProjectIDs(projectIDs []string) error {
	if len(projectIDs) == 0 {
		return errors.New("project IDs list is empty")
	}

	for _, projectID := range projectIDs {
		if strings.TrimSpace(projectID) == "" {
			return errors.New("project ID cannot be empty")
		}

		if len(projectID) > 100 {
			return errors.New("project ID length exceeds 100 characters")
		}
	}

	return nil
}

// BroadcastToAll 广播消息到所有在线客户端（不考虑订阅）
func (s *MessageService) BroadcastToAll(message *domain.PushMessage) error {
	if err := s.validateMessage(message); err != nil {
		return err
	}

	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	log.Printf("[MessageService] Broadcasting to all clients: %d", len(s.hub.clients))

	for client := range s.hub.clients {
		select {
		case client.send <- message:
			// 发送成功
		default:
			// 客户端阻塞，跳过
			log.Printf("[MessageService] Client send buffer full, skipping: UserID=%s", client.UserID)
		}
	}

	return nil
}

// GetProjectSubscribers 获取指定项目的订阅者数量
func (s *MessageService) GetProjectSubscribers(projectID string) int {
	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	if subscribers, exists := s.hub.projects[projectID]; exists {
		return len(subscribers)
	}
	return 0
}

// IsClientSubscribed 检查客户端是否订阅了指定项目
func (s *MessageService) IsClientSubscribed(client *Client, projectID string) bool {
	return client.SubscribedProjects[projectID]
}
