package utils

import (
	"fmt"
	"gin-quickstart/internal/model"
	"math/rand/v2"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gosimple/slug"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func String(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

func Slugify(s string) string {
	slug := slug.Make(s)

	if slug == "" {
		slug = String(8)
	}

	return slug
}

func PasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func UploadPostAttachmentFilesToS3(post *model.Post, s3client interface{}, files []*multipart.FileHeader) *model.Attachment {
	err_env := godotenv.Load()

	if err_env != nil {
		return nil
	}

	var Tasks []func() *model.Attachment

	for _, file := range files {
		fmt.Printf("Uploading file: %s\n", file.Filename)
		fileBinary, err := file.Open()

		if err != nil {
			return nil
		}

		defer fileBinary.Close()
		task := func(f *multipart.FileHeader) func() *model.Attachment {
			return func() *model.Attachment {
				_, uErr := s3client.(*s3.S3).PutObject(&s3.PutObjectInput{
					Bucket: aws.String(os.Getenv("S3_BUCKET")),
					Key:    aws.String(f.Filename), // You can customize the key as needed
					Body:   fileBinary,             // You should provide the actual file content here
					ACL:    aws.String("public-read"),
				})

				attachment := model.Attachment{
					PostID:     post.ID,
					UploaderId: post.AuthorID,
					Url:        fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), f.Filename),
					Filename:   f.Filename,
					MimeType:   f.Header.Get("Content-Type"),
					FileSize:   f.Size,
				}

				post.Attachments = append(post.Attachments, attachment)

				if uErr != nil {
					return nil
				}

				return &attachment
			}
		}(file)

		Tasks = append(Tasks, task)
	}

	attachment := runGoRoutineParallel(Tasks)

	return attachment
}

func runGoRoutineParallel(tasks []func() *model.Attachment) *model.Attachment {
	attachment := make(chan *model.Attachment, len(tasks))

	for _, task := range tasks {
		go func(t func() *model.Attachment) {
			attachment <- t()
		}(task)
	}

	return <-attachment
}
