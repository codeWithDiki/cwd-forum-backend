package handler

import (
	"gin-quickstart/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AttachmentHandler struct {
	s *service.AttachmentService
}

func NewAttachmentHandler(s *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{
		s: s,
	}
}

// GETTER
func (h AttachmentHandler) GetAllAttachments(c *gin.Context) {
	attachments, err := h.s.GetAllAttachments(c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    attachments,
	})
}

func (h AttachmentHandler) GetAttachmentByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid attachment ID",
		})
		return
	}

	attachment, err := h.s.GetAttachmentByID(id, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    attachment,
	})
}

func (h AttachmentHandler) GetAttachmentsByPostID(c *gin.Context) {
	postIDParam := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid post ID",
		})
		return
	}

	attachments, err := h.s.GetAttachmentsByPostID(postID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    attachments,
	})
}

// SETTER
func (h AttachmentHandler) DeleteAttachment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid attachment ID",
		})
		return
	}

	attachment, err := h.s.GetAttachmentByID(id, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if attachment == nil {
		c.JSON(404, gin.H{
			"success": false,
			"error":   "Attachment not found",
		})
		return
	}

	err = h.s.Delete(attachment, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Attachment deleted successfully",
	})
}
