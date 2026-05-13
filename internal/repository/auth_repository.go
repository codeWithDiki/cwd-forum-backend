package repository

import (
	"gin-quickstart/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthRepository struct {
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewAuthRepository(db *gorm.DB, redis *redis.Client) *AuthRepository {
	return &AuthRepository{
		GormDB:      db,
		RedisClient: redis,
	}
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GETTER
func (r *AuthRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.GormDB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r AuthRepository) GetUserById(id uint64) (*model.User, error) {
	var user model.User
	err := r.GormDB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r AuthRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.GormDB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SETTER
func (r *AuthRepository) Register(user *model.User) error {
	return r.GormDB.Create(user).Error
}

func (r *AuthRepository) ChangePassword(userID uint64, newPassword string) error {
	var user model.User
	err := r.GormDB.First(&user, userID).Error
	if err != nil {
		return err
	}

	user.Password = newPassword

	return r.GormDB.Save(&user).Error
}

func (r *AuthRepository) UpdateProfile(
	user *model.User,
) error {

	return r.GormDB.Save(&user).Error
}

func (r *AuthRepository) Logout(userID uint64) error {
	var user model.User
	err := r.GormDB.First(&user, userID).Error
	if err != nil {
		return err
	}

	now := time.Now()

	user.LastSeenAt = &now

	return r.GormDB.Save(&user).Error
}
