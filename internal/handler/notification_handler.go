package handler

import (
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type NotificationHandler struct {
	Service *service.NotificationService
	Hub     *service.WSHub
}

func NewNotificationHandler(service *service.NotificationService, hub *service.WSHub) *NotificationHandler {
	return &NotificationHandler{Service: service, Hub: hub}
}

type CreateNotificationRequest struct {
	ThreadID *uint  `json:"thread_id,omitempty" form:"thread_id,omitempty"`
	PostID   *uint  `json:"post_id,omitempty" form:"post_id,omitempty"`
	UserID   uint   `json:"user_id" binding:"required" form:"user_id" `
	Type     string `json:"type" binding:"required" form:"type" `
	Payload  string `json:"payload" form:"payload" `
}

type UpdateNotificationRequest struct {
	IsRead *bool `json:"is_read,omitempty" binding:"required" form:"is_read,omitempty"`
}

func (h NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")

	notifications, err := h.Service.GetNotificationsByUserID(c, uint64(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": notifications})
}

func (h NotificationHandler) GetNotificationByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification ID"})
		return
	}

	userID := c.GetUint("user_id")
	notification, err := h.Service.GetNotificationByID(c, id, uint64(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "notification": notification})
}

func (h NotificationHandler) CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "notification": notification})
}

func (h NotificationHandler) MarkAsRead(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification ID"})
		return
	}

	userID := c.GetUint("user_id")
	notification, err := h.Service.MarkNotificationAsRead(c, id, uint64(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "notification": notification})
}

func (h NotificationHandler) UpdateNotification(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification ID"})
		return
	}

	var req UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.IsRead == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "is_read is required"})
		return
	}

	userID := c.GetUint("user_id")
	notification, err := h.Service.UpdateNotificationReadState(c, id, uint64(userID), *req.IsRead)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "notification": notification})
}

func (h NotificationHandler) DeleteNotification(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid notification ID"})
		return
	}

	userID := c.GetUint("user_id")
	if err := h.Service.DeleteNotification(c, id, uint64(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "notification deleted"})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

func (h NotificationHandler) HandleWebSocket(c *gin.Context) {
	userID := c.GetUint("user_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "failed to upgrade connection"})
		return
	}

	h.Hub.Register(userID, conn)

	// Keep connection alive and handle disconnect
	go func() {
		defer h.Hub.Unregister(userID)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
