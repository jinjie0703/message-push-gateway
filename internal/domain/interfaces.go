package domain

// MessageBroadcaster 消息广播器接口
type MessageBroadcaster interface {
	// Broadcast 广播消息到指定 ProjectID
	Broadcast(message *PushMessage)
}

// ClientManager 客户端管理器接口
type ClientManager interface {
	// Register 注册客户端
	Register(client ClientConnection)
	
	// Unregister 注销客户端
	Unregister(client ClientConnection)
	
	// HandleSubscription 处理订阅/取消订阅
	HandleSubscription(client ClientConnection, action string, projectIDs []string)
}

// ClientConnection 客户端连接接口
type ClientConnection interface {
	// GetUserID 获取用户ID
	GetUserID() string
	
	// GetUsername 获取用户名
	GetUsername() string
	
	// GetSubscribedProjects 获取已订阅的项目ID
	GetSubscribedProjects() map[string]bool
	
	// Send 发送消息到客户端
	Send(message *PushMessage) error
}

// TokenParser JWT Token 解析器接口
type TokenParser interface {
	// Parse 解析并验证 Token
	Parse(tokenString string) (*JWTClaims, error)
}

// MessageValidator 消息验证器接口
type MessageValidator interface {
	// Validate 验证消息格式
	Validate(message *PushMessage) error
}

// SubscriptionValidator 订阅验证器接口
type SubscriptionValidator interface {
	// ValidateProjectIDs 验证项目ID列表
	ValidateProjectIDs(projectIDs []string) error
	
	// ValidateAction 验证订阅动作
	ValidateAction(action string) error
}
