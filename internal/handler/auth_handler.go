package handler

import (
	"fmt"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	s *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{
		s: s,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// GETTER
func (h AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	token, err := h.s.Login(req.Username, req.Password, c)

	if err != nil {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Invalid username or password : " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Login successful",
		"token":   token,
	})
}

// SETTER
func (h AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	err := h.s.Register(req.Name, req.Username, req.Email, req.Password, enum.RoleUser.String(), c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	token, tokenExists := c.Get("token")

	if !exists || !tokenExists {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
		return
	}

	err := h.s.Logout(uint64(userID.(uint)), token.(string))

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "User logged out successfully",
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	user, err := h.s.GetUserByID(uint64(userID.(uint)), c)

	file, err := c.FormFile("avatar")

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Failed to get file from request",
		})
		return
	}

	if user == nil {
		c.JSON(404, gin.H{
			"success": false,
			"error":   "User not found",
		})
	}

	var req struct {
		Name   *string `json:"name"`
		Email  *string `json:"email" binding:"omitempty,email"`
		Avatar *string `json:"avatar"`
		Bio    *string `json:"bio"`
	}

	wp := c.MustGet("fileUploadWorkerPool").(*workerpool.WorkerPool)
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d_%s%s", user.ID, uuid.New().String(), ext)

	name := c.PostForm("name")
	email := c.PostForm("email")
	avatar := c.PostForm("avatar")
	bio := c.PostForm("bio")

	req.Name = &name
	req.Email = &email
	req.Avatar = &avatar
	req.Bio = &bio
	iconUrlStr := fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName)
	req.Avatar = &iconUrlStr

	wp.Submit(func() {
		s3Client := c.MustGet("s3Client").(*s3.S3)

		if user.Avatar != "" {
			// Extract the S3 key from the Avatar URL
			avatarUrl := user.Avatar
			s3Key := avatarUrl[strings.LastIndex(avatarUrl, "/")+1:]

			// Check if the file exists in S3
			_, err := s3Client.HeadObject(&s3.HeadObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(s3Key),
			})

			if err != nil {
				fmt.Printf("Error checking S3 for avatar: %v\n", err)
			}

			// If the file exists, delete it
			_, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(s3Key),
			})

			if err != nil {
				fmt.Printf("Error deleting old avatar from S3: %v\n", err)
			}

		}

		if file != nil {

			fileBinary, err := file.Open()

			if err != nil {
				fmt.Printf("Error opening new avatar file: %v\n", err)
				return
			}
			defer fileBinary.Close()

			_, err = s3Client.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(newFileName),
				Body:   fileBinary,
				ACL:    aws.String("public-read"),
			})

			if err != nil {
				fmt.Printf("Error uploading new avatar to S3: %v\n", err)
				return
			}
		}
	})

	uErr := h.s.UpdateProfile(uint64(user.ID), req.Name, req.Email, req.Avatar, req.Bio, c)

	if uErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   uErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
	})
}
