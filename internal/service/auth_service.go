package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/jwt"
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
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	r *repository.AuthRepository
}

func NewAuthService(r *repository.AuthRepository) *AuthService {
	return &AuthService{
		r: r,
	}
}

// GETTER
func (s *AuthService) Login(
	username string,
	password string,
	ctx *gin.Context,
) (string, error) {
	user, err := s.r.GetUserByUsername(username)

	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", err
	}

	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	var now = time.Now()
	user.LastLoginAt = &now

	s.r.RedisClient.Set(ctx, token, user.ID, time.Hour*24)

	err = s.r.GormDB.Save(&user).Error
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s AuthService) GetUserByID(id uint64, ctx *gin.Context) (*model.User, error) {
	getStatus := s.r.RedisClient.Get(ctx, "user:"+strconv.FormatUint(id, 10))

	if getStatus.Err() == nil {
		var user model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &user)

		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	return s.r.GetUserById(id)
}

func (s AuthService) GetUserByUsername(username string, ctx *gin.Context) (*model.User, error) {
	getStatus := s.r.RedisClient.Get(ctx, "user:username:"+username)

	if getStatus.Err() == nil {
		var user model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &user)

		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	user, err := s.r.GetUserByUsername(username)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(user)

	if err != nil {
		return nil, err
	}

	s.r.RedisClient.Set(ctx, "user:username:"+username, json, time.Hour)

	return user, nil
}

func (s AuthService) GetUserByEmail(email string, ctx *gin.Context) (*model.User, error) {
	getStatus := s.r.RedisClient.Get(ctx, "user:email:"+email)

	if getStatus.Err() == nil {
		var user model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &user)

		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	user, err := s.r.GetUserByEmail(email)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(user)

	if err != nil {
		return nil, err
	}

	s.r.RedisClient.Set(ctx, "user:email:"+email, json, time.Hour)

	return user, nil
}

func (s AuthService) GetLoggedUser(ctx *gin.Context) (*model.User, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return nil, errors.New("User not logged in")
	}

	return s.GetUserByID(uint64(userID.(uint)), ctx)
}

// SETTER
func (s *AuthService) Register(
	Name string,
	Username string,
	Email string,
	Password string,
	Role string,
	ctx *gin.Context,
) error {
	user := &model.User{
		Name:     Name,
		Username: Username,
		Email:    Email,
		Password: Password,
		Role:     Role,
	}

	usernameExists, _ := s.GetUserByUsername(Username, ctx)
	if usernameExists != nil {
		return errors.New("Username already Exists!")
	}

	emailExists, _ := s.GetUserByEmail(Email, ctx)
	if emailExists != nil {
		return errors.New("Email already Exists!")
	}

	return s.r.Register(user)
}

func (s *AuthService) ChangePassword(userID uint64, newPassword string, ctx *gin.Context) error {
	user, err := s.GetUserByID(userID, ctx)

	if err != nil {
		return err
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	return s.r.ChangePassword(uint64(user.ID), string(newPasswordHash))
}

func (s *AuthService) Logout(userID uint64, token string) error {
	delTokenStatus := s.r.RedisClient.Del(context.Background(), token)

	if delTokenStatus.Err() != nil {
		return delTokenStatus.Err()
	}

	return s.r.Logout(userID)
}

func (s *AuthService) UpdateProfile(
	ctx *gin.Context,
	userID uint64,
	Name string,
	Email string,
	Bio string,
	File *multipart.FileHeader,
) error {
	user, err := s.GetUserByID(userID, ctx)

	if err != nil {
		return err
	}

	if Name != "" {
		user.Name = Name
	}

	if Email != "" {
		user.Email = Email
	}

	if Bio != "" {
		user.Bio = Bio
	}

	if File != nil {
		wp := ctx.MustGet("fileUploadWorkerPool").(*workerpool.WorkerPool)
		ext := filepath.Ext(File.Filename)
		newFileName := fmt.Sprintf("%d_%s%s", user.ID, uuid.New().String(), ext)

		user.Avatar = fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName)

		wp.Submit(func() {
			s3Client := ctx.MustGet("s3Client").(*s3.S3)

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

			if File != nil {

				fileBinary, err := File.Open()

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

				fmt.Println("Test")
			}
		})
	}

	updateError := s.r.UpdateProfile(user)

	if updateError != nil {
		return updateError
	}

	s.r.RedisClient.Del(ctx, "user:"+strconv.FormatUint(userID, 10))
	s.r.RedisClient.Del(ctx, "user:username:"+user.Username)
	s.r.RedisClient.Del(ctx, "user:email:"+user.Email)

	return nil
}
