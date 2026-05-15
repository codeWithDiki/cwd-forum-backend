package handler

import (
	"gin-quickstart/internal/service"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
)

type BadgeHandler struct {
	s *service.BadgeService
}

func NewBadgeHandler(s *service.BadgeService) *BadgeHandler {
	return &BadgeHandler{
		s: s,
	}
}

type CreateBadgeRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description" binding:"required"`
	IconUrl         string `json:"icon_url" binding:"required"`
	CriteriaType    string `json:"criteria_type" binding:"required"`
	CriteriaValue   int    `json:"criteria_value" binding:"required"`
	FontColor       string `json:"font_color" binding:"required"`
	BackgroundColor string `json:"background_color" binding:"required"`
}

type UpdateBadgeRequest struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	IconUrl         string `json:"icon_url,omitempty"`
	CriteriaType    string `json:"criteria_type,omitempty"`
	CriteriaValue   int    `json:"criteria_value,omitempty"`
	FontColor       string `json:"font_color,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
}

// GETTER
func (h BadgeHandler) GetAllBadges(c *gin.Context) {
	badges, err := h.s.GetAllBadges(c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    badges,
	})
}

func (h BadgeHandler) GetBadgeByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid badge ID",
		})
		return
	}

	badge, err := h.s.GetBadgeByID(id, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    badge,
	})
}

// SETTER
func (h *BadgeHandler) Create(c *gin.Context) {
	var req CreateBadgeRequest
	file, err := c.FormFile("icon")

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Failed to get file from request",
		})
		return
	}

	req.Name = c.PostForm("name")
	req.Description = c.PostForm("description")
	req.CriteriaType = c.PostForm("criteria_type")
	criteriaValueStr := c.PostForm("criteria_value")
	criteriaValue, err := strconv.Atoi(criteriaValueStr)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid criteria_value, must be an integer",
		})
		return
	}
	req.CriteriaValue = criteriaValue
	req.FontColor = c.PostForm("font_color")
	req.BackgroundColor = c.PostForm("background_color")

	badge, err := h.s.Create(
		c,
		req.Name,
		req.Description,
		req.CriteriaType,
		req.CriteriaValue,
		req.FontColor,
		req.BackgroundColor,
		file,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   "Failed to create badge: " + err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    badge,
	})

}

func (h *BadgeHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid badge ID",
		})
		return
	}

	file, _ := c.FormFile("icon")

	var req UpdateBadgeRequest

	nameForm := c.PostForm("name")
	descriptionForm := c.PostForm("description")
	criteriaTypeForm := c.PostForm("criteria_type")
	criteriaValueStr := c.PostForm("criteria_value")
	fontColorForm := c.PostForm("font_color")
	backgroundColorForm := c.PostForm("background_color")

	if nameForm != "" {
		req.Name = nameForm
	}

	if descriptionForm != "" {
		req.Description = descriptionForm
	}

	if criteriaTypeForm != "" {
		req.CriteriaType = criteriaTypeForm
	}

	if criteriaValueStr != "" {
		criteriaValue, err := strconv.Atoi(criteriaValueStr)
		if err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"error":   "Invalid criteria_value, must be an integer",
			})
			return
		}
		req.CriteriaValue = criteriaValue
	}

	if fontColorForm != "" {
		req.FontColor = fontColorForm
	}

	if backgroundColorForm != "" {
		req.BackgroundColor = backgroundColorForm
	}

	badge, err := h.s.Update(
		c,
		id,
		req.Name,
		req.Description,
		req.CriteriaType,
		req.CriteriaValue,
		req.FontColor,
		req.BackgroundColor,
		file,
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
		"data":    badge,
	})
}

func (h *BadgeHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error":   "Invalid badge ID",
		})
		return
	}

	badge, err := h.s.GetBadgeByID(id, c)

	wp := c.MustGet("fileUploadWorkerPool").(*workerpool.WorkerPool)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if badge.IconUrl != "" {
		wp.Submit(func() {
			// Extract the S3 key from the IconUrl
			iconUrl := badge.IconUrl
			s3Key := iconUrl[strings.LastIndex(iconUrl, "/")+1:]

			s3client := c.MustGet("s3Client")
			_, dErr := s3client.(*s3.S3).DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(s3Key),
			})

			if dErr != nil {
				log.Printf("Failed to delete file from S3: %v", dErr)
				return
			}
		})
	}

	err = h.s.Delete(badge, c)

	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}
