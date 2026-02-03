package api

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
	"web_websocket/internal/application"
	"web_websocket/internal/domain"

	"github.com/gin-gonic/gin"
)

// HTTPHandler HTTP 接口处理器
type HTTPHandler struct {
	hub            *application.Hub
	messageService *application.MessageService
}

// NewHTTPHandler 创建 HTTP 处理器
func NewHTTPHandler(hub *application.Hub) *HTTPHandler {
	return &HTTPHandler{
		hub:            hub,
		messageService: application.NewMessageService(hub),
	}
}

// HandlePushMessage 接收外部平台的推送消息 Webhook
func (h *HTTPHandler) HandlePushMessage(c *gin.Context) {
	var req struct {
		ProjectID string          `json:"project_id" binding:"required"`
		Type      string          `json:"type" binding:"required"`
		SourceID  string          `json:"source_id" binding:"required"`
		Data      json.RawMessage `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 构造推送消息
	message := &domain.PushMessage{
		Type:      req.Type,
		ProjectID: req.ProjectID,
		SourceID:  req.SourceID,
		Ts:        time.Now().UnixMilli(),
		Data:      req.Data,
	}

	// 使用 MessageService 发布消息（带验证）
	if err := h.messageService.PublishMessage(message); err != nil {
		log.Printf("[HTTP] Failed to publish message: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[HTTP] Message received: ProjectID=%s, Type=%s, SourceID=%s", req.ProjectID, req.Type, req.SourceID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "push message received",
	})
}

// HandleTestPush 测试推送接口
func (h *HTTPHandler) HandleTestPush(c *gin.Context) {
	projectID := c.DefaultQuery("project_id", "test_project")
	msgType := c.DefaultQuery("type", "test_message")
	sourceID := c.DefaultQuery("source_id", "test_device_01")

	testData := map[string]interface{}{
		"temperature": 25.5 + rand.Float64()*5,
		"humidity":    40 + rand.Float64()*20,
		"location": map[string]float64{
			"lat": 31.2304 + (rand.Float64()-0.5)*0.1,  // 上海附近随机位置
			"lng": 121.4737 + (rand.Float64()-0.5)*0.1,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(testData)

	message := &domain.PushMessage{
		Type:      msgType,
		ProjectID: projectID,
		SourceID:  sourceID,
		Ts:        time.Now().UnixMilli(),
		Data:      jsonData,
	}

	if err := h.messageService.PublishMessage(message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "test message sent",
		"data":    testData,
	})
}

// HandleHealth 健康检查接口
func (h *HTTPHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// HandleStats 获取系统统计信息
func (h *HTTPHandler) HandleStats(c *gin.Context) {
	stats := h.messageService.GetClientStats()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}