package repository

import (
	"errors"
	"fmt"
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

type StorageRepository struct {
	log        *logger.Logger
	s3Client   *s3.S3
	WorkerPool *workerpool.WorkerPool
}

func NewStorageRepository(log *logger.Logger, s3Client *s3.S3, workerPool *workerpool.WorkerPool) *StorageRepository {
	return &StorageRepository{
		log:        log,
		s3Client:   s3Client,
		WorkerPool: workerPool,
	}
}

func (r *StorageRepository) Driver(ctx *gin.Context) (string, error) {
	driver := os.Getenv("FILE_STORAGE_TYPE")
	r.log.Debug(ctx, "Storage Driver", r.log.Field("Driver", driver))
	if driver == "" {
		return "", errors.New("Need to set FILE_STORAGE_TYPE")
	}

	return driver, nil
}

func (r *StorageRepository) GenerateFileUrl(ctx *gin.Context, file *multipart.FileHeader) string {
	driver, err := r.Driver(ctx)
	if err != nil {
		r.log.Error(ctx, "Storage Driver Error", err)
		return ""
	}

	if driver == "s3" {
		return r.GenerateS3FileUrl(ctx, file)
	}

	if driver == "local" {
		return r.GenerateLocalFileUrl(ctx, file)
	}

	return ""
}

func (r *StorageRepository) GenerateFileName(ctx *gin.Context, file *multipart.FileHeader) string {
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	return newFileName
}

func (r *StorageRepository) GenerateS3FileUrl(ctx *gin.Context, file *multipart.FileHeader) string {
	newFileName := r.GenerateFileName(ctx, file)

	url := fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName)

	return url
}

func (r *StorageRepository) GenerateLocalFileUrl(ctx *gin.Context, file *multipart.FileHeader) string {
	fileName := r.GenerateFileName(ctx, file)

	// Implement local file URL generation logic here
	return "http://localhost:8080/uploads/" + fileName
}

func (r *StorageRepository) GetFileSize(ctx *gin.Context, file *multipart.FileHeader) (int64, error) {
	f, err := file.Open()
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return file.Size, nil
}

func (r *StorageRepository) GetFileContentType(ctx *gin.Context, file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil

}

func (r *StorageRepository) UploadFile(ctx *gin.Context, fileHeader *multipart.FileHeader, destinationPath string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	url := r.GenerateFileUrl(ctx, fileHeader)
	fileName := r.GenerateFileName(ctx, fileHeader)
	driver, err := r.Driver(ctx)

	if err != nil {
		r.log.Error(ctx, "Storage Driver Error", err)
		return "", err
	}

	switch driver {
	case "s3":
		r.WorkerPool.Submit(func() {

			r.log.Info(ctx, "Uploading file to S3 in background worker")
			fileBinary, err := fileHeader.Open()

			if err != nil {
				r.log.Error(ctx, "Failed to open file for S3 upload", err)
				ctx.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Failed to open file: " + err.Error(),
				})
				return
			}

			defer fileBinary.Close()

			_, uErr := r.s3Client.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(fileName), // You can customize the key as needed
				Body:   fileBinary,           // You should provide the actual file content here
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

	case "local":
		savePath := filepath.Join(destinationPath, fileName)
		if err := os.MkdirAll(destinationPath, os.ModePerm); err != nil {
			return "", err
		}

		out, err := os.Create(savePath)
		if err != nil {
			return "", err
		}
		defer out.Close()

		if _, err = file.Seek(0, 0); err != nil {
			return "", err
		}
	default:
		return "", errors.New("Unsupported storage driver")
	}

	return url, nil
}

func (r *StorageRepository) DeleteFile(ctx *gin.Context, fileUrl string) error {
	driver, err := r.Driver(ctx)
	if err != nil {
		r.log.Error(ctx, "Storage Driver Error", err)
		return err
	}

	switch driver {
	case "s3":
		s3Key := fileUrl[strings.LastIndex(fileUrl, "/")+1:]
		_, err := r.s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(os.Getenv("S3_BUCKET")),
			Key:    aws.String(s3Key),
		})

		if err != nil {
			r.log.Error(ctx, "Failed to delete file from S3", err)
			return err
		}
	case "local":
		localPath := filepath.Join("uploads", fileUrl[strings.LastIndex(fileUrl, "/")+1:])
		if err := os.Remove(localPath); err != nil {
			r.log.Error(ctx, "Failed to delete local file", err)
			return err
		}
	default:
		return errors.New("Unsupported storage driver")
	}

	return nil
}
