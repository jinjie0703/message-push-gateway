package domain

import (
	"encoding/json"
	"time"
)

// PushMessage 通用推送消息结构（信封模式）
type PushMessage struct {
	// ===== Header: 路由与元数据 =====
	Type      string `json:"type"`       // 消息类型: gps, fence_alarm, device_status, etc.
	ProjectID string `json:"project_id"` // [路由键] 项目ID，Hub根据此字段决定推送给谁
	SourceID  string `json:"source_id"`  // [发送源] 数据来源: 车牌号/设备ID/摄像头ID
	Ts        int64  `json:"ts"`         // 时间戳(毫秒)

	// ===== Body: 业务载荷 =====
	// 使用 RawMessage 实现零拷贝转发，后端不解析直接透传
	Data json.RawMessage `json:"data"`
}

// SubscribeCommand 前端订阅指令
type SubscribeCommand struct {
	Action     string   `json:"action"`      // subscribe 或 unsubscribe
	ProjectIDs []string `json:"project_ids"` // 项目ID列表
}

// JWTClaims JWT 载荷结构（匹配后端项目）
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	RoleName string `json:"role_name"` // 角色名（匹配后端）
	Exp      int64  `json:"exp"`       // 过期时间戳
}

// Valid 实现 jwt.Claims 接口
func (c *JWTClaims) Valid() error {
	if time.Now().Unix() > c.Exp {
		return ErrTokenExpired
	}
	return nil
}

// 错误定义
var (
	ErrTokenExpired = &DomainError{Code: "TOKEN_EXPIRED", Message: "token has expired"}
	ErrInvalidToken = &DomainError{Code: "INVALID_TOKEN", Message: "invalid token"}
)

// DomainError 领域错误
type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}
