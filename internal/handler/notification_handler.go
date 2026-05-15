package handler

import (
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	Service *service.NotificationService
}

func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{Service: service}
}

type CreateNotificationRequest struct {
	ThreadID *uint  `json:"thread_id,omitempty"`
	PostID   *uint  `json:"post_id,omitempty"`
	UserID   uint   `json:"user_id" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Payload  string `json:"payload" binding:"required"`
}

type UpdateNotificationRequest struct {
	IsRead *bool `json:"is_read,omitempty"`
}

func (h NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")

	notifications, err := h.Service.GetNotificationsByUserID(c, uint64(userID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": notifications})
}

func (h NotificationHandler) GetNotificationByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "invalid notification ID"})
		return
	}

	userID := c.GetUint("user_id")
	notification, err := h.Service.GetNotificationByID(c, id, uint64(userID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "notification": notification})
}

func (h NotificationHandler) CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	notification := &model.Notification{
		ThreadId: req.ThreadID,
		PostId:   req.PostID,
		UserId:   req.UserID,
		Type:     req.Type,
		Payload:  req.Payload,
		IsRead:   false,
	}

	notification, err := h.Service.CreateNotification(c, notification)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"success": true, "notification": notification})
}

func (h NotificationHandler) MarkAsRead(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "invalid notification ID"})
		return
	}

	userID := c.GetUint("user_id")
	notification, err := h.Service.MarkNotificationAsRead(c, id, uint64(userID))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "notification": notification})
}

func (h NotificationHandler) UpdateNotification(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "invalid notification ID"})
		return
	}

	var req UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.IsRead == nil {
		c.JSON(400, gin.H{"success": false, "error": "is_read is required"})
		return
	}

	userID := c.GetUint("user_id")
	notification, err := h.Service.UpdateNotificationReadState(c, id, uint64(userID), *req.IsRead)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "notification": notification})
}

func (h NotificationHandler) DeleteNotification(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "invalid notification ID"})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.Service.DeleteNotification(c, id, uint64(userID)); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "notification deleted"})
}
