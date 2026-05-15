package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type BadgeService struct {
	r *repository.BadgeRepository
}

func NewBadgeService(r *repository.BadgeRepository) *BadgeService {
	return &BadgeService{
		r: r,
	}
}

// GETTER
func (s BadgeService) GetAllBadges(ctx *gin.Context) ([]*model.Badge, error) {
	getStatus := s.r.RedisClient.Get(ctx, "badges")

	if getStatus.Err() == nil {
		var badges []*model.Badge
		err := json.Unmarshal([]byte(getStatus.Val()), &badges)

		if err != nil {
			return nil, err
		}

		return badges, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	badges, err := s.r.GetAllBadges()

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(badges)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "badges", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return badges, nil
}

func (s BadgeService) GetBadgeByID(id uint64, ctx *gin.Context) (*model.Badge, error) {
	getStatus := s.r.RedisClient.Get(ctx, "badge:id:"+strconv.FormatUint(id, 10))

	if getStatus.Err() == nil {
		var badge model.Badge
		err := json.Unmarshal([]byte(getStatus.Val()), &badge)

		if err != nil {
			return nil, err
		}

		return &badge, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	badge, err := s.r.GetBadgeByID(id)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(badge)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "badge:id:"+strconv.FormatUint(id, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
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

	cErr := s.r.Create(badge)

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

	delCmdStatus := s.r.RedisClient.Del(ctx, "badges")

	if delCmdStatus.Err() != nil {
		return nil, delCmdStatus.Err()
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
	badge, err := s.r.GetBadgeByID(ID)

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
				ctx.JSON(400, gin.H{
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
				ctx.JSON(500, gin.H{
					"success": false,
					"error":   "Failed to upload file to S3: " + uErr.Error(),
				})
				return
			}
		})
	}

	err = s.r.Update(badge)

	if err != nil {
		return nil, err
	}

	delCmdStatus := s.r.RedisClient.Del(ctx, "badges:"+strconv.FormatUint(ID, 10))

	if delCmdStatus.Err() != nil {
		return nil, delCmdStatus.Err()
	}

	return badge, nil
}

func (s *BadgeService) Delete(badge *model.Badge, ctx *gin.Context) error {
	delCmdStatus := s.r.RedisClient.Del(ctx, "badges:"+strconv.FormatUint(uint64(badge.ID), 10))

	if delCmdStatus.Err() != nil {
		return delCmdStatus.Err()
	}

	delListCmdStatus := s.r.RedisClient.Del(ctx, "badges")

	if delListCmdStatus.Err() != nil {
		return delListCmdStatus.Err()
	}

	return s.r.Delete(badge)
}
