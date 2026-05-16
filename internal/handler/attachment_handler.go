package handler

import (
	"gin-quickstart/internal/service"
	"net/http"
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    attachments,
	})
}

func (h AttachmentHandler) GetAttachmentByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid attachment ID",
		})
		return
	}

	attachment, err := h.s.GetAttachmentByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    attachment,
	})
}

func (h AttachmentHandler) GetAttachmentsByPostID(c *gin.Context) {
	postIDParam := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid post ID",
		})
		return
	}

	attachments, err := h.s.GetAttachmentsByPostID(c, postID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    attachments,
	})
}

// SETTER
func (h AttachmentHandler) DeleteAttachment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid attachment ID",
		})
		return
	}

	attachment, err := h.s.GetAttachmentByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if attachment == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Attachment not found",
		})
		return
	}

	err = h.s.Delete(c, attachment)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Attachment deleted successfully",
	})
}
