package handler

import (
	"fmt"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/service"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gammazero/workerpool"
	"github.com/google/uuid"

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
	ThreadID uint   `json:"thread_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	AuthorID uint   `json:"author_id" binding:"required"`
	ParentID *uint  `json:"parent_id,omitempty"`
}

type UpdatePostRequest struct {
	Content *string `json:"content,omitempty"`
}

// GETTER
func (h PostHandler) GetAllPosts(c *gin.Context) {
	posts, err := h.s.GetAllPosts(c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    posts,
	})
}

func (h PostHandler) GetPostByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	post, err := h.s.GetPostByID(id, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"post":    post,
	})
}

func (h PostHandler) GetPostsByThreadID(c *gin.Context) {
	threadIDParam := c.Param("thread_id")

	threadID, err := strconv.ParseUint(threadIDParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid thread ID",
		})
		return
	}

	posts, err := h.s.GetPostsByThreadID(threadID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"posts":   posts,
	})
}

func (h PostHandler) GetPostsByAuthorID(c *gin.Context) {
	authorIDParam := c.Param("author_id")

	authorID, err := strconv.ParseUint(authorIDParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid author ID",
		})
		return
	}

	posts, err := h.s.GetPostsByAuthorID(authorID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"posts":   posts,
	})
}

func (h PostHandler) GetPostsByParentID(c *gin.Context) {
	parentIDParam := c.Param("parent_id")

	parentID, err := strconv.ParseUint(parentIDParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid parent ID",
		})
		return
	}

	posts, err := h.s.GetPostsByParentID(parentID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"posts":   posts,
	})
}

func (h PostHandler) GetPostVotes(c *gin.Context) {
	idParam := c.Param("id")

	postID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	votes, err := h.s.GetPostVotes(postID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"votes":   votes,
	})
}

// SETTER
func (h *PostHandler) Create(c *gin.Context) {
	var req CreatePostRequest
	var Attachments []*multipart.FileHeader
	var UserId uint
	wp, wpExists := c.Get("fileUploadWorkerPool") // Create a worker pool with 10 workers

	if wpExists == false {
		c.JSON(500, gin.H{
			"success": false,
			"error":   "Failed to get worker pool",
		})
		return
	}

	userID, iErr := c.Get("user_id")

	if iErr {
		UserId = userID.(uint)
	} else {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	form, err := c.MultipartForm()

	if err != nil {
		c.JSON(400, gin.H{
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
		c.JSON(400, gin.H{
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
		req.ThreadID,
		req.Content,
		req.AuthorID,
		req.ParentID,
		c,
	)

	for _, file := range Attachments {

		wp.(*workerpool.WorkerPool).Submit(func() {
			fmt.Println("Uploading from Post")
			ext := filepath.Ext(file.Filename)
			newFileName := fmt.Sprintf("%d_%s%s", post.ID, uuid.New().String(), ext)

			s3client := c.MustGet("s3Client")
			fileBinary, err := file.Open()

			if err != nil {
				return
			}

			_, uErr := s3client.(*s3.S3).PutObject(&s3.PutObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(newFileName), // You can customize the key as needed
				Body:   fileBinary,              // You should provide the actual file content here
				ACL:    aws.String("public-read"),
			})

			attachment := model.Attachment{
				PostID:     post.ID,
				UploaderId: post.AuthorID,
				Url:        fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName),
				Filename:   newFileName,
				MimeType:   file.Header.Get("Content-Type"),
				FileSize:   file.Size,
			}

			post.Attachments = append(post.Attachments, attachment)

			h.s.CreateAttachment(post, &attachment, c)

			if uErr != nil {
				return
			}

		})
	}

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"post":    post,
	})
}

func (h *PostHandler) Update(c *gin.Context) {
	var req UpdatePostRequest

	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	post, err := h.s.Update(
		ID,
		req.Content,
		c,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"post":    post,
	})
}

func (h *PostHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	err = h.s.Delete(ID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Post deleted successfully",
	})
}

func (h *PostHandler) VotePost(c *gin.Context) {

	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	var req struct {
		Value int `json:"value" binding:"required"` // +1 for upvote, -1 for downvote
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	err = h.s.Vote(ID, uint64(userID.(uint)), req.Value, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Vote recorded successfully",
	})

}

func (h *PostHandler) ReactPost(c *gin.Context) {
	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	var req struct {
		Emoji int `json:"emoji" binding:"required"` // Emoji represented as an integer code
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	err = h.s.React(ID, uint64(userID.(uint)), req.Emoji, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Reaction recorded successfully",
	})
}

func (h *PostHandler) MarkAsSolution(c *gin.Context) {
	idParam := c.Param("id")
	ID, err := strconv.ParseUint(idParam, 10, 64)
	userId := c.GetUint("user_id")

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "invalid post ID",
		})
		return
	}

	post, err := h.s.GetPostByID(ID, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err = h.s.MarkAsSolution(uint64(post.ID), uint64(userId), c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Post marked as solution successfully",
	})
}
