package handler

import (
	"gin-quickstart/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TagHandler struct {
	s *service.TagService
}

func NewTagHandler(s *service.TagService) *TagHandler {
	return &TagHandler{
		s: s,
	}
}

type CreateTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Slug  string `json:"slug" binding:"required"`
	Color string `json:"color" binding:"required"`
}

type UpdateTagRequest struct {
	Name  *string `json:"name,omitempty"`
	Slug  *string `json:"slug,omitempty"`
	Color *string `json:"color,omitempty"`
}

func (h TagHandler) GetAllTags(c *gin.Context) {
	tags, err := h.s.GetAllTags(c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    tags,
	})
}

func (h TagHandler) GetTagByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid tag ID",
		})
		return
	}

	tag, err := h.s.GetTagByID(id, c)
	if err != nil {
		c.JSON(404, gin.H{
			"success": false,
			"error":   "tag not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    tag,
	})
}

func (h TagHandler) GetTagBySlug(c *gin.Context) {
	slug := c.Param("slug")

	tag, err := h.s.GetTagBySlug(slug, c)
	if err != nil {
		c.JSON(404, gin.H{
			"success": false,
			"error":   "tag not found",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    tag,
	})
}

func (h *TagHandler) CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	tag, err := h.s.Create(req.Name, req.Slug, req.Color, c)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    tag,
	})
}

func (h *TagHandler) UpdateTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid tag ID",
		})
		return
	}

	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	tag, err := h.s.Update(id, req.Name, req.Slug, req.Color, c)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    tag,
	})
}

func (h *TagHandler) DeleteTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid tag ID",
		})
		return
	}

	err = h.s.Delete(id, c)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "tag deleted successfully",
	})
}
