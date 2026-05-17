package handler

import (
	"gin-quickstart/internal/service"
	"gin-quickstart/pkg/utils"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ThreadHandler struct {
	s *service.ThreadService
}

func NewThreadHandler(s *service.ThreadService) *ThreadHandler {
	return &ThreadHandler{
		s: s,
	}
}

type CreateThreadRequest struct {
	CategoryID  uint                    `json:"category_id" binding:"required" form:"category_id"`
	Title       string                  `json:"title" binding:"required" form:"title"`
	Slug        string                  `json:"slug,omitempty" binding:"omitempty,slug,no_spaces" form:"slug,omitempty"`
	Content     string                  `json:"content" binding:"required" form:"content"`
	TagIDs      []uint                  `json:"tag_ids,omitempty" form:"tag_ids,omitempty"`
	Attachments []*multipart.FileHeader `json:"attachments,omitempty" form:"attachments,omitempty"`
}

type UpdateThreadRequest struct {
	CategoryID uint   `json:"category_id,omitempty" form:"category_id,omitempty"`
	Title      string `json:"title,omitempty" form:"title,omitempty"`
	Slug       string `json:"slug,omitempty" form:"slug,omitempty" binding:"omitempty,slug,no_spaces"`
	IsSolved   bool   `json:"is_solved,omitempty" form:"is_solved,omitempty"`
}

// GETTER
func (h ThreadHandler) GetAllThreads(c *gin.Context) {
	threads, err := h.s.GetAllThreads(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
	})
}

func (h ThreadHandler) GetThreadByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid thread ID",
		})
		return
	}

	thread, err := h.s.GetThreadByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    thread,
	})
}

func (h ThreadHandler) GetThreadBySlug(c *gin.Context) {
	slug := c.Param("slug")

	thread, err := h.s.GetThreadBySlug(c, slug)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    thread,
	})
}

func (h ThreadHandler) GetThreadsByCategoryID(c *gin.Context) {
	categoryIDParam := c.Param("category_id")
	categoryID, err := strconv.ParseUint(categoryIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	threads, err := h.s.GetThreadsByCategoryID(c, uint(categoryID))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
	})
}

func (h ThreadHandler) GetThreadsByAuthorID(c *gin.Context) {
	authorIDParam := c.Param("author_id")
	authorID, err := strconv.ParseUint(authorIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid author ID",
		})
		return
	}

	threads, err := h.s.GetThreadsByAuthorID(c, uint(authorID))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
	})
}

func (h ThreadHandler) GetThreadsByTagID(c *gin.Context) {
	tagIDParam := c.Param("tag_id")
	tagID, err := strconv.ParseUint(tagIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tag ID",
		})
		return
	}

	threads, err := h.s.GetThreadsByTagID(c, uint(tagID))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    threads,
	})
}

// SETTER
func (h *ThreadHandler) Create(c *gin.Context) {
	var req CreateThreadRequest

	userIdParam, userIdExists := c.Get("user_id")

	if !userIdExists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    utils.BuildValidationErrors(err, &req),
			"error":   "Validation Errors",
		})
		return
	}

	thread, post, err := h.s.Create(
		c,
		req.CategoryID,
		req.Title,
		req.Slug,
		req.Content,
		uint(userIdParam.(uint)),
		req.TagIDs,
		req.Attachments,
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
		"message": "Thread created successfully",
		"data": gin.H{
			"thread": thread,
			"post":   post,
		},
	})
}

func (h *ThreadHandler) Update(c *gin.Context) {
	var req UpdateThreadRequest

	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid thread ID",
		})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	thread, err := h.s.Update(
		c,
		ID,
		req.CategoryID,
		req.Title,
		req.Slug,
		req.IsSolved,
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
		"message": "Thread updated successfully",
		"data":    thread,
	})

}

func (h *ThreadHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid thread ID",
		})
		return
	}

	var req struct {
		Reason *string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	var moderatorID *uint
	if userID, exists := c.Get("user_id"); exists {
		uid := uint(userID.(uint))
		moderatorID = &uid
	}

	err = h.s.Delete(c, ID, moderatorID, req.Reason)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread deleted successfully",
	})
}
