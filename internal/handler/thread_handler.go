package handler

import (
	"fmt"
	"gin-quickstart/internal/service"
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
	CategoryID uint   `json:"category_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Slug       string `json:"slug,omitempty"`
	Content    string `json:"content" binding:"required"`
	AuthorID   uint   `json:"author_id" binding:"required"`
	TagIDs     []uint `json:"tag_ids,omitempty"`
}

type UpdateThreadRequest struct {
	CategoryID uint   `json:"category_id,omitempty"`
	Title      string `json:"title,omitempty"`
	Slug       string `json:"slug,omitempty"`
	IsSolved   bool   `json:"is_solved,omitempty"`
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
	categoryIdParam := c.PostForm("category_id")
	categoryID, err := strconv.ParseUint(categoryIdParam, 10, 64)
	wp, wpExists := c.Get("fileUploadWorkerPool") // Create a worker pool with 10 workers

	fmt.Println(wp)

	if wpExists == false {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get worker pool from context",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid category ID",
		})
		return
	}

	req := CreateThreadRequest{
		CategoryID: uint(categoryID),
		Title:      c.PostForm("title"),
		Slug:       c.PostForm("slug"),
		Content:    c.PostForm("content"),
		AuthorID:   c.GetUint("user_id"),
	}
	var Attachments []*multipart.FileHeader

	form, err := c.MultipartForm()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to parse multipart form: " + err.Error(),
		})
		return
	}

	files := form.File["attachments"]

	for _, file := range files {
		Attachments = append(Attachments, file)
	}

	thread, post, err := h.s.Create(
		c,
		req.CategoryID,
		req.Title,
		req.Slug,
		req.Content,
		req.AuthorID,
		req.TagIDs,
		Attachments,
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

	err = h.s.Delete(c, ID)

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
