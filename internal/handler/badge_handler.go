package handler

import (
	"gin-quickstart/internal/service"
	"gin-quickstart/pkg/utils"
	"log"
	"mime/multipart"
	"net/http"
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
	Name            string                `form:"name" binding:"required"`
	Description     string                `form:"description" binding:"required"`
	CriteriaType    string                `form:"criteria_type" binding:"required"`
	CriteriaValue   int                   `form:"criteria_value" binding:"required"`
	FontColor       string                `form:"font_color" binding:"required,hexcolor"`
	BackgroundColor string                `form:"background_color" binding:"required, hex_color"`
	Icon            *multipart.FileHeader `form:"icon" binding:"required"`
}

type UpdateBadgeRequest struct {
	Name            string                `form:"name" binding:"omitempty"`
	Description     string                `form:"description" binding:"omitempty"`
	CriteriaType    string                `form:"criteria_type" binding:"omitempty"`
	CriteriaValue   int                   `form:"criteria_value" binding:"omitempty"`
	FontColor       string                `form:"font_color" binding:"omitempty, hex_color"`
	BackgroundColor string                `form:"background_color" binding:"omitempty, hex_color"`
	Icon            *multipart.FileHeader `form:"icon" binding:"omitempty"`
}

// GETTER
func (h BadgeHandler) GetAllBadges(c *gin.Context) {
	badges, err := h.s.GetAllBadges(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    badges,
	})
}

func (h BadgeHandler) GetBadgeByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid badge ID",
		})
		return
	}

	badge, err := h.s.GetBadgeByID(c, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    badge,
	})
}

// SETTER
func (h *BadgeHandler) Create(c *gin.Context) {
	var req CreateBadgeRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    utils.BuildValidationErrors(err, &req),
			"error":   err.Error(),
		})
		return
	}

	badge, err := h.s.Create(
		c,
		req.Name,
		req.Description,
		req.CriteriaType,
		req.CriteriaValue,
		req.FontColor,
		req.BackgroundColor,
		req.Icon,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create badge: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    badge,
	})

}

func (h *BadgeHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid badge ID",
		})
		return
	}

	var req UpdateBadgeRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data":    utils.BuildValidationErrors(err, &req),
			"error":   err.Error(),
		})
		return
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
		"data":    badge,
	})
}

func (h *BadgeHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid badge ID",
		})
		return
	}

	badge, err := h.s.GetBadgeByID(c, id)

	wp := c.MustGet("fileUploadWorkerPool").(*workerpool.WorkerPool)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
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

	err = h.s.Delete(c, badge)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
