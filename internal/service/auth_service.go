package service

import (
	"context"
	"errors"
	"fmt"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/email"
	"gin-quickstart/pkg/jwt"
	"gin-quickstart/pkg/logger"
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
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	log         *logger.Logger
	r           *repository.AuthRepository
	emailClient *email.EmailClient
}

func NewAuthService(log *logger.Logger, r *repository.AuthRepository) *AuthService {
	return &AuthService{
		log:         log,
		r:           r,
		emailClient: email.NewEmailClient(),
	}
}

// GETTER
func (s *AuthService) Login(
	ctx *gin.Context,
	username string,
	password string,
) (string, error) {
	user, err := s.r.GetUserByUsername(ctx, username)
	s.log.Debug(ctx, "Service Login Called", s.log.Field("username", username))
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		s.log.Warn(ctx, "Service Login Failed - Invalid Password", s.log.Field("username", username))
		return "", err
	}

	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		s.log.Error(ctx, "Service Login Failed - Token Generation Error", err, s.log.Field("username", username))
		return "", err
	}

	var now = time.Now()
	user.LastLoginAt = &now

	s.r.RedisClient.Set(ctx, token, user.ID, time.Hour*24)
	s.log.Info(ctx, "Service Login Successful", s.log.Field("username", username), s.log.Field("user_id", user.ID))

	err = s.r.GormDB.Model(&user).Update("last_login_at", now).Error
	s.log.Debug(ctx, "Service Login - Updated LastLoginAt", s.log.Field("username", username), s.log.Field("last_login_at", user.LastLoginAt))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s AuthService) GetUserByID(ctx *gin.Context, id uint64) (*model.User, error) {
	return s.r.GetUserById(ctx, id)
}

func (s AuthService) GetUserByUsername(ctx *gin.Context, username string) (*model.User, error) {
	return s.r.GetUserByUsername(ctx, username)
}

func (s AuthService) GetUserByEmail(ctx *gin.Context, email string) (*model.User, error) {
	return s.r.GetUserByEmail(ctx, email)
}

func (s AuthService) GetLoggedUser(ctx *gin.Context) (*model.User, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return nil, errors.New("User not logged in")
	}

	return s.GetUserByID(ctx, uint64(userID.(uint)))
}

// SETTER
func (s *AuthService) Register(
	ctx *gin.Context,
	Name string,
	Username string,
	Email string,
	Password string,
	Role string,
) error {
	user := &model.User{
		Name:     Name,
		Username: Username,
		Email:    Email,
		Password: Password,
		Role:     Role,
	}

	usernameExists, _ := s.GetUserByUsername(ctx, Username)
	if usernameExists != nil {
		return errors.New("Username already Exists!")
	}

	emailExists, _ := s.GetUserByEmail(ctx, Email)
	if emailExists != nil {
		return errors.New("Email already Exists!")
	}

	err := s.r.Register(ctx, user)
	if err != nil {
		return err
	}

	go func() {
		if err := s.emailClient.SendWelcomeEmail(Email, Name); err != nil {
			log.Printf("Failed to send welcome email to %s: %v", Email, err)
		}
	}()
	return nil
}

func (s *AuthService) ChangePassword(ctx *gin.Context, userID uint64, newPassword string) error {
	user, err := s.GetUserByID(ctx, userID)

	if err != nil {
		return err
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	return s.r.ChangePassword(ctx, uint64(user.ID), string(newPasswordHash))
}

func (s *AuthService) Logout(ctx *gin.Context, userID uint64, token string) error {
	delTokenStatus := s.r.RedisClient.Del(context.Background(), token)

	if delTokenStatus.Err() != nil {
		return delTokenStatus.Err()
	}

	return s.r.Logout(ctx, userID)
}

func (s *AuthService) UpdateProfile(
	ctx *gin.Context,
	userID uint64,
	Name string,
	Email string,
	Bio string,
	File *multipart.FileHeader,
) error {
	user, err := s.GetUserByID(ctx, userID)

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
					s.log.Error(ctx, "Failed to check if old avatar exists in S3", err)
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
					s.log.Error(ctx, "Failed to open new avatar file", err)
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
					s.log.Error(ctx, "Failed to upload new avatar to S3", err)
					return
				}

				user.Avatar = fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName)

				s.log.Info(
					ctx,
					"Successfully uploaded new avatar to S3",
					s.log.Field("username", user.Username),
					s.log.Field("avatar_url", user.Avatar),
				)

				updateErr := s.r.GormDB.Save(&user).Error

				if updateErr != nil {
					s.log.Error(ctx, "Failed to update user avatar in database", updateErr)
					return
				}

				s.log.Info(ctx, "Successfully updated user avatar in database", s.log.Field("username", user.Username))
				s.r.RedisClient.Del(ctx, "user:all")
				s.r.RedisClient.Del(ctx, "user:"+strconv.FormatUint(userID, 10))
				s.r.RedisClient.Del(ctx, "user:username:"+user.Username)
				s.r.RedisClient.Del(ctx, "user:email:"+user.Email)
				s.log.Info(ctx, "Cleared user cache after avatar update", s.log.Field("username", user.Username))

			}
		})
	}

	updateError := s.r.UpdateProfile(ctx, user)

	if updateError != nil {
		return updateError
	}

	s.r.RedisClient.Del(ctx, "user:"+strconv.FormatUint(userID, 10))
	s.r.RedisClient.Del(ctx, "user:username:"+user.Username)
	s.r.RedisClient.Del(ctx, "user:email:"+user.Email)

	return nil
}

func (s *AuthService) ForgotPassword(email string, resetBaseURL string, ctx *gin.Context) error {
	user, err := s.r.GetUserByEmail(ctx, email)
	if err != nil {
		// Return nil to avoid leaking whether email exists
		return nil
	}

	token := uuid.New().String()
	if err := s.r.StoreResetToken(ctx, user.Email, token); err != nil {
		return err
	}

	resetLink := resetBaseURL + "?token=" + token
	go func() {
		if err := s.emailClient.SendForgotPasswordEmail(user.Email, resetLink); err != nil {
			log.Printf("Failed to send forgot password email to %s: %v", user.Email, err)
		}
	}()
	return nil
}

func (s *AuthService) ResetPassword(token string, newPassword string, ctx *gin.Context) error {
	email, err := s.r.GetEmailByResetToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	user, err := s.r.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.r.ChangePassword(ctx, uint64(user.ID), string(hashed))
}
