package repository

import (
	"context"
	"gin-quickstart/internal/model"
	"gin-quickstart/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserRepository struct {
	log         logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewUserRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *UserRepository {
	return &UserRepository{
		log:         *log,
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r UserRepository) GetAllUsers(ctx context.Context) ([]model.User, error) {
	r.log.Debug(ctx, "GetAllUser Repo Called")
	var users []model.User
	err := r.GormDB.Find(&users).Error
	if err != nil {
		r.log.Error(ctx, "GetAllUser Repo Error", err)
		return nil, err
	}
	return users, nil
}

func (r UserRepository) GetUserByID(id uint64) (*model.User, error) {
	var user model.User
	err := r.GormDB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r UserRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.GormDB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.GormDB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r UserRepository) GetFollowers(userID uint64) ([]model.User, error) {
	var followers []model.User
	err := r.GormDB.Joins("JOIN user_users ON user_users.follower_id = users.id").
		Where("user_users.user_id = ?", userID).
		Find(&followers).Error
	if err != nil {
		return nil, err
	}
	return followers, nil
}

func (r UserRepository) GetFollowing(userID uint64) ([]model.User, error) {
	var following []model.User
	err := r.GormDB.Joins("JOIN user_users ON user_users.user_id = users.id").
		Where("user_users.follower_id = ?", userID).
		Find(&following).Error
	if err != nil {
		return nil, err
	}
	return following, nil
}

// SETTER
func (r *UserRepository) Create(user *model.User) error {
	return r.GormDB.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.GormDB.Save(user).Error
}

func (r *UserRepository) Delete(user *model.User) error {
	user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	return r.GormDB.Save(user).Error
}

func (r *UserRepository) HardDelete(user *model.User) error {
	return r.GormDB.Unscoped().Delete(user).Error
}

func (r *UserRepository) Restore(user *model.User) error {
	user.DeletedAt = gorm.DeletedAt{Time: time.Time{}, Valid: false}
	return r.GormDB.Save(user).Error
}

func (r *UserRepository) Follow(userID uint64, followerID uint64) error {
	return r.GormDB.Exec("INSERT INTO user_users (follower_id, followed_id) VALUES (?, ?)", userID, followerID).Error
}

func (r *UserRepository) Unfollow(userID uint64, followerID uint64) error {
	return r.GormDB.Exec("DELETE FROM user_users WHERE follower_id = ? AND followed_id = ?", userID, followerID).Error
}

func (r *UserRepository) IsFollowing(userID uint64, followerID uint64) (bool, error) {
	var count int64
	err := r.GormDB.Table("user_users").Where("follower_id = ? AND followed_id = ?", userID, followerID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
