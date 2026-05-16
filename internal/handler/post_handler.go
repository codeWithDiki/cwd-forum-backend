package handler

import (
	"gin-quickstart/internal/service"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	s *service.PostService
}

func NewPostHandler(s *service.PostService) *PostHandler {
	return &PostHandler{
		s: s,
	}
}

type CreatePostRequest struct {
	ThreadID uint   `json:"thread_id" binding:"required" form:"thread_id" `
	Content  string `json:"content" binding:"required" form:"content" `
	AuthorID uint   `json:"author_id" binding:"required" form:"author_id" `
	ParentID *uint  `json:"parent_id,omitempty" form:"parent_id,omitempty" `
}

type UpdatePostRequest struct {
	Content *string `json:"content,omitempty" binding:"required"`
}

// GETTER
func (h PostHandler) GetAllPosts(c *gin.Context) {
	posts, err := h.s.GetAllPosts(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    posts,
	})
}

func (h PostHandler) GetPostByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	post, err := h.s.GetPostByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"post":    post,
	})
}

func (h PostHandler) GetPostsByThreadID(c *gin.Context) {
	threadIDParam := c.Param("thread_id")

	threadID, err := strconv.ParseUint(threadIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid thread ID",
		})
		return
	}

	posts, err := h.s.GetPostsByThreadID(c, threadID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"posts":   posts,
	})
}

func (h PostHandler) GetPostsByAuthorID(c *gin.Context) {
	authorIDParam := c.Param("author_id")

	authorID, err := strconv.ParseUint(authorIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid author ID",
		})
		return
	}

	posts, err := h.s.GetPostsByAuthorID(c, authorID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"posts":   posts,
	})
}

func (h PostHandler) GetPostsByParentID(c *gin.Context) {
	parentIDParam := c.Param("parent_id")

	parentID, err := strconv.ParseUint(parentIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid parent ID",
		})
		return
	}

	posts, err := h.s.GetPostsByParentID(c, parentID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"posts":   posts,
	})
}

func (h PostHandler) GetPostVotes(c *gin.Context) {
	idParam := c.Param("id")

	postID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	votes, err := h.s.GetPostVotes(c, postID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"votes":   votes,
	})
}

// SETTER
func (h *PostHandler) Create(c *gin.Context) {
	var req CreatePostRequest
	var Attachments []*multipart.FileHeader
	var UserId uint

	userID, iErr := c.Get("user_id")

	if iErr {
		UserId = userID.(uint)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

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

	req.Content = c.PostForm("content")

	threadID, err := strconv.ParseUint(c.PostForm("thread_id"), 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid thread ID : " + err.Error(),
		})
		return
	}

	req.ThreadID = uint(threadID)

	parentID, err := strconv.ParseUint(c.PostForm("parent_id"), 10, 64)

	if err == nil {
		req.ParentID = new(uint)
		*req.ParentID = uint(parentID)
	}

	req.AuthorID = UserId

	post, err := h.s.Create(
		c,
		req.ThreadID,
		req.Content,
		req.AuthorID,
		req.ParentID,
		Attachments,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"post":    post,
	})
}

func (h *PostHandler) Update(c *gin.Context) {
	var req UpdatePostRequest

	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid post ID",
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

	post, err := h.s.Update(
		c,
		ID,
		req.Content,
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
		"post":    post,
	})
}

func (h *PostHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid post ID",
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
		"message": "Post deleted successfully",
	})
}

func (h *PostHandler) VotePost(c *gin.Context) {

	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	var req struct {
		Value int `json:"value" binding:"required"` // +1 for upvote, -1 for downvote
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	err = h.s.Vote(c, ID, uint64(userID.(uint)), req.Value)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vote recorded successfully",
	})

}

func (h *PostHandler) ReactPost(c *gin.Context) {
	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	var req struct {
		Emoji int `json:"emoji" binding:"required"` // Emoji represented as an integer code
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	err = h.s.React(c, ID, uint64(userID.(uint)), req.Emoji)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reaction recorded successfully",
	})
}

func (h *PostHandler) MarkAsSolution(c *gin.Context) {
	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)
	userId := c.GetUint("user_id")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	post, err := h.s.GetPostByID(c, ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err = h.s.MarkAsSolution(c, uint64(post.ID), uint64(userId))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Post marked as solution successfully",
	})
}
