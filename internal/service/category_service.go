package service

import (
	"errors"
	"fmt"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"
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

type CategoryService struct {
	log *logger.Logger
	r   *repository.CategoryRepository
}

func NewCategoryService(log *logger.Logger, r *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		log: log,
		r:   r,
	}
}

// GETTER
func (s CategoryService) GetAllCategories(ctx *gin.Context) ([]model.Category, error) {

	categories, err := s.r.GetAllCategories(ctx)
	s.log.Debug(ctx, "Service GetAllCategories Called", s.log.Field("Count", len(categories)))

	if err != nil {
		s.log.Error(ctx, "Service GetAllCategories Error", err)
		return nil, err
	}

	s.log.Debug(ctx, "Service GetAllCategories Result", s.log.Field("Count", len(categories)))

	return categories, nil
}

func (s CategoryService) GetCategoryByID(ctx *gin.Context, id uint64) (*model.Category, error) {

	category, err := s.r.GetCategoryByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s CategoryService) GetCategoryBySlug(ctx *gin.Context, slug string) (*model.Category, error) {

	category, err := s.r.GetCategoryBySlug(ctx, slug)

	if err != nil {
		return nil, err
	}

	return category, nil
}

// SETTER
func (s *CategoryService) Create(
	ctx *gin.Context,
	ParentID *uint,
	Name string,
	Slug string,
	Description string,
	SortOrder int,
	IsPrivate bool,
	Icon *multipart.FileHeader,
) (*model.Category, error) {
	category := &model.Category{
		ParentID:    ParentID,
		Name:        Name,
		Slug:        Slug,
		Description: Description,
		SortOrder:   SortOrder,
		IsPrivate:   IsPrivate,
	}

	if ParentID != nil {
		parentCategory, err := s.r.GetCategoryByID(ctx, uint64(*ParentID))

		if err != nil {
			return nil, errors.New("Parent category not found")
		}

		if parentCategory == nil {
			return nil, errors.New("Parent category not found")
		}
	}

	slugExists, _ := s.r.GetCategoryBySlug(ctx, Slug)

	if slugExists != nil {
		return nil, errors.New("Slug already exists")
	}

	if Icon != nil {
		wp, wpExists := ctx.Get("fileUploadWorkerPool")

		if !wpExists {
			return nil, errors.New("Worker pool not found")
		}

		fileBinary, fErr := Icon.Open()

		if fErr != nil {
			return nil, errors.New("Failed to open file")
		}

		ext := filepath.Ext(Icon.Filename)
		newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		var iconUrl *string

		iconUrlStr := fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName)
		iconUrl = &iconUrlStr

		category.IconUrl = *iconUrl

		defer fileBinary.Close()

		wp.(*workerpool.WorkerPool).Submit(func() {
			fmt.Println("Uploading from Post")

			s3client := ctx.MustGet("s3Client")
			fileBinary, err := Icon.Open()

			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Failed to open file: " + err.Error(),
				})
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

	err := s.r.Create(ctx, category)

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Update(
	ctx *gin.Context,
	ID uint64,
	ParentID uint,
	Name string,
	Slug string,
	Description string,
	SortOrder int,
	IsPrivate bool,
	Icon *multipart.FileHeader,
) (*model.Category, error) {
	category, err := s.r.GetCategoryByID(ctx, ID)

	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, errors.New("Category not found")
	}

	if ParentID != 0 {
		parentCategory, err := s.r.GetCategoryByID(ctx, uint64(ParentID))

		if err != nil {
			return nil, errors.New("Parent category not found")
		}

		if parentCategory == nil {
			return nil, errors.New("Parent category not found")
		}

		if category.ID == parentCategory.ID {
			return nil, errors.New("Category cannot be its own parent")
		}
	}

	if Name != "" {
		category.Name = Name
	}

	if Slug != "" {
		slugExists, _ := s.r.GetCategoryBySlug(ctx, Slug)

		if slugExists != nil && slugExists.ID != category.ID {
			return nil, errors.New("Slug already exists")
		}

		category.Slug = Slug
	}

	if Description != "" {
		category.Description = Description
	}

	if Icon != nil {

		wp, wpExists := ctx.Get("fileUploadWorkerPool")

		if !wpExists {
			s.log.Error(ctx, "Worker pool not found", errors.New("Worker pool not found"))
			return nil, errors.New("Worker pool not found")
		}

		fileBinary, fErr := Icon.Open()

		if fErr != nil {
			s.log.Error(ctx, "Failed to open file", fErr)
			return nil, errors.New("Failed to open file")
		}

		ext := filepath.Ext(Icon.Filename)
		newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		var iconUrl *string

		iconUrlStr := fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName)
		iconUrl = &iconUrlStr

		defer fileBinary.Close()

		wp.(*workerpool.WorkerPool).Submit(func() {
			s.log.Info(ctx, "Uploading file to S3 in background worker")

			s3client := ctx.MustGet("s3Client")
			fileBinary, err := Icon.Open()

			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Failed to open file: " + err.Error(),
				})
				return
			}

			defer fileBinary.Close()

			oldIconUrl := category.IconUrl
			s3Key := oldIconUrl[strings.LastIndex(oldIconUrl, "/")+1:]
			_, dErr := s3client.(*s3.S3).DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(s3Key),
			})

			if dErr != nil {
				s.log.Error(ctx, "Failed to delete old file from S3", dErr)
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

			category.IconUrl = *iconUrl

			s.r.Update(ctx, category)
		})
	}

	if SortOrder != 0 {
		category.SortOrder = SortOrder
	}

	if IsPrivate != false {
		category.IsPrivate = IsPrivate
	}

	err = s.r.Update(ctx, category)

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Delete(ctx *gin.Context, ID uint64) error {
	category, err := s.r.GetCategoryByID(ctx, ID)

	if err != nil {
		return err
	}

	if category == nil {
		return errors.New("Category not found")
	}

	threads := category.Threads

	if len(threads) > 0 {
		return errors.New("Cannot delete category with existing threads")
	}

	subcategories := category.Categories

	if len(subcategories) > 0 {
		return errors.New("Cannot delete category with existing subcategories")
	}

	return nil
}
