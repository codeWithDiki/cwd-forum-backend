package handler

import (
	"gin-quickstart/internal/service"
	"gin-quickstart/pkg/utils"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	s *service.CategoryService
}

type CreateCategoryRequest struct {
	ParentID    *uint                 `json:"parent_id" form:"parent_id" binding:"omitempty,gt=0"`
	Name        string                `json:"name" form:"name" binding:"required"`
	Slug        string                `json:"slug" form:"slug" binding:"required,slug,no_spaces"`
	Description string                `json:"description" form:"description" binding:"omitempty"`
	SortOrder   int                   `json:"sort_order" form:"sort_order" binding:"omitempty"`
	IsPrivate   bool                  `json:"is_private" form:"is_private" binding:"omitempty"`
	Icon        *multipart.FileHeader `json:"icon" binding:"omitempty" form:"icon"`
}

type UpdateCategoryRequest struct {
	ParentID    uint                  `json:"parent_id,omitempty" form:"parent_id,omitempty" binding:"omitempty,gt=0"`
	Name        string                `json:"name,omitempty" form:"name,omitempty" binding:"omitempty"`
	Slug        string                `json:"slug,omitempty" form:"slug,omitempty" binding:"omitempty,slug,no_spaces"`
	Description string                `json:"description,omitempty" form:"description,omitempty" binding:"omitempty"`
	SortOrder   int                   `json:"sort_order,omitempty" form:"sort_order,omitempty" binding:"omitempty"`
	IsPrivate   bool                  `json:"is_private,omitempty" form:"is_private,omitempty" binding:"omitempty"`
	Icon        *multipart.FileHeader `json:"icon,omitempty" form:"icon" binding:"omitempty"`
}

func NewCategoryHandler(s *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		s: s,
	}
}

// GETTER
func (h CategoryHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.s.GetAllCategories(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

func (h CategoryHandler) GetCategoryByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	category, err := h.s.GetCategoryByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    category,
	})
}

func (h CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.s.GetCategoryBySlug(c, slug)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    category,
	})
}

// SETTER
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    utils.BuildValidationErrors(err, &req),
			"error":   "Validation Errors",
		})
		return
	}

	category, err := h.s.Create(
		c,
		req.ParentID,
		req.Name,
		req.Slug,
		req.Description,
		req.SortOrder,
		req.IsPrivate,
		req.Icon,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    category,
	})
}

func (h *CategoryHandler) Update(c *gin.Context) {
	var req UpdateCategoryRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    utils.BuildValidationErrors(err, &req),
			"error":   "Validation Errors",
		})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	category, err := h.s.Update(
		c,
		id,
		req.ParentID,
		req.Name,
		req.Slug,
		req.Description,
		req.SortOrder,
		req.IsPrivate,
		req.Icon,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    category,
	})
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	err = h.s.Delete(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Category deleted successfully",
	})
}
