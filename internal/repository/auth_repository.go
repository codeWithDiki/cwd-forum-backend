package repository

import (
	"context"
	"gin-quickstart/internal/model"
	"gin-quickstart/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewAuthRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *AuthRepository {
	return &AuthRepository{
		log:         log,
		GormDB:      db,
		RedisClient: redis,
	}
}

func CheckPasswordHash(ctx *gin.Context, password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GETTER
func (r *AuthRepository) GetUserByUsername(ctx *gin.Context, username string) (*model.User, error) {
	var user model.User
	err := r.GormDB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r AuthRepository) GetUserById(ctx *gin.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.GormDB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r AuthRepository) GetUserByEmail(ctx *gin.Context, email string) (*model.User, error) {
	var user model.User
	err := r.GormDB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SETTER
func (r *AuthRepository) Register(ctx *gin.Context, user *model.User) error {
	return r.GormDB.Create(user).Error
}

func (r *AuthRepository) ChangePassword(ctx *gin.Context, userID uint64, newPassword string) error {
	var user model.User
	err := r.GormDB.First(&user, userID).Error
	if err != nil {
		return err
	}

	user.Password = newPassword

	return r.GormDB.Save(&user).Error
}

func (r *AuthRepository) UpdateProfile(
	ctx *gin.Context,
	user *model.User,
) error {

	return r.GormDB.Save(&user).Error
}

func (r *AuthRepository) StoreResetToken(ctx context.Context, email string, token string) error {
	return r.RedisClient.Set(ctx, "reset:"+token, email, time.Minute*30).Err()
}

func (r *AuthRepository) GetEmailByResetToken(ctx context.Context, token string) (string, error) {
	email, err := r.RedisClient.Get(ctx, "reset:"+token).Result()
	if err != nil {
		return "", err
	}
	r.RedisClient.Del(ctx, "reset:"+token)
	return email, nil
}

func (r *AuthRepository) Logout(ctx context.Context, userID uint64) error {
	var user model.User
	err := r.GormDB.First(&user, userID).Error
	if err != nil {
		return err
	}

	now := time.Now()

	user.LastSeenAt = &now

	return r.GormDB.Save(&user).Error
}
