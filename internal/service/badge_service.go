package service

import (
	"errors"
	"fmt"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"
	"log"
	"mime/multipart"
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

type BadgeService struct {
	log *logger.Logger
	r   *repository.BadgeRepository
}

func NewBadgeService(log *logger.Logger, r *repository.BadgeRepository) *BadgeService {
	return &BadgeService{
		log: log,
		r:   r,
	}
}

// GETTER
func (s BadgeService) GetAllBadges(ctx *gin.Context) ([]model.Badge, error) {
	badges, err := s.r.GetAllBadges(ctx)
	s.log.Debug(ctx, "Service GetAllBadges Called", s.log.Field("Count", len(badges)))

	if err != nil {
		s.log.Error(ctx, "Service GetAllBadges Error", err)
		return nil, err
	}

	return badges, nil
}

func (s BadgeService) GetBadgeByID(ctx *gin.Context, id uint64) (*model.Badge, error) {
	badge, err := s.r.GetBadgeByID(ctx, id)
	s.log.Debug(ctx, "Service GetBadgeByID Called", s.log.Field("BadgeID", id))

	if err != nil {
		s.log.Error(ctx, "Service GetBadgeByID Error", err, s.log.Field("BadgeID", id))
		return nil, err
	}

	return badge, nil
}

// SETTER
func (s *BadgeService) Create(
	ctx *gin.Context,
	Name string,
	Description string,
	CriteriaType string,
	CriteriaValue int,
	FontColor string,
	BackgroundColor string,
	File *multipart.FileHeader,
) (*model.Badge, error) {

	wp, wpExists := ctx.Get("fileUploadWorkerPool")

	if !wpExists {
		return nil, errors.New("Worker Pool is not available")
	}

	fileBinary, fErr := File.Open()

	if fErr != nil {
		return nil, fErr
	}

	defer fileBinary.Close()

	ext := filepath.Ext(File.Filename)
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	iconUrlStr := fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName)

	criteriaType, err := enum.BadgeCriteriaTypeFromString(CriteriaType)

	if err == false {
		return nil, errors.New("Criteria is not registered")
	}

	badge := &model.Badge{
		Name:          Name,
		Description:   Description,
		CriteriaType:  criteriaType.String(),
		IconUrl:       iconUrlStr,
		CriteriaValue: CriteriaValue,
	}

	cErr := s.r.Create(ctx, badge)

	wp.(*workerpool.WorkerPool).Submit(func() {
		fmt.Println("Uploading from Post")

		s3client := ctx.MustGet("s3Client")

		defer fileBinary.Close()

		_, uErr := s3client.(*s3.S3).PutObject(&s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET")),
			Key:    aws.String(newFileName), // You can customize the key as needed
			Body:   fileBinary,              // You should provide the actual file content here
			ACL:    aws.String("public-read"),
		})

		if uErr != nil {
			return
		}

		s.r.GormDB.Model(&model.Badge{}).Where("id = ?", badge.ID).Update("icon_url", iconUrlStr)

	})

	if cErr != nil {
		return nil, cErr
	}

	return badge, nil
}

func (s *BadgeService) Update(
	ctx *gin.Context,
	ID uint64,
	Name string,
	Description string,
	CriteriaType string,
	CriteriaValue int,
	FontColor string,
	BackgroundColor string,
	File *multipart.FileHeader,
) (*model.Badge, error) {
	badge, err := s.r.GetBadgeByID(ctx, ID)

	if err != nil {
		return nil, err
	}

	if badge == nil {
		return nil, errors.New("Badge not found")
	}

	if Name != "" {
		badge.Name = Name
	}

	if Description != "" {
		badge.Description = Description
	}

	if CriteriaType != "" {
		criteriaType, err := enum.BadgeCriteriaTypeFromString(CriteriaType)

		if err == false {
			return nil, errors.New("Criteria is not registered")
		}

		badge.CriteriaType = criteriaType.String()
	}

	if CriteriaValue != 0 {
		badge.CriteriaValue = CriteriaValue
	}

	if FontColor != "" {
		badge.FontColor = FontColor
	}

	if BackgroundColor != "" {
		badge.BackgroundColor = BackgroundColor
	}

	if File != nil {
		wp, wpExists := ctx.Get("fileUploadWorkerPool")

		if !wpExists {
			return nil, errors.New("Worker Pool is not available")
		}

		fileBinary, fErr := File.Open()

		if fErr != nil {
			return nil, fErr
		}

		ext := filepath.Ext(File.Filename)
		newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		var iconUrl *string

		iconUrlStr := fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName)
		iconUrl = &iconUrlStr

		badge.IconUrl = *iconUrl

		defer fileBinary.Close()

		wp.(*workerpool.WorkerPool).Submit(func() {
			fmt.Println("Uploading from Post")

			s3client := ctx.MustGet("s3Client")
			fileBinary, err := File.Open()

			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Failed to open file: " + err.Error(),
				})
				return
			}

			defer fileBinary.Close()

			oldIconUrl := badge.IconUrl
			s3Key := oldIconUrl[strings.LastIndex(oldIconUrl, "/")+1:]
			_, dErr := s3client.(*s3.S3).DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(s3Key),
			})

			if dErr != nil {
				log.Printf("Failed to delete file from S3: %v", dErr)
				return
			}

			_, uErr := s3client.(*s3.S3).PutObject(&s3.PutObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(newFileName), // You can customize the key as needed
				Body:   fileBinary,              // You should provide the actual file content here
				ACL:    aws.String("public-read"),
			})

			if uErr != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to upload file to S3: " + uErr.Error(),
				})
				return
			}
		})
	}

	err = s.r.Update(ctx, badge)

	if err != nil {
		return nil, err
	}

	return badge, nil
}

func (s *BadgeService) Delete(ctx *gin.Context, badge *model.Badge) error {
	return s.r.Delete(ctx, badge)
}
