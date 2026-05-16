package repository

import (
	"context"
	"encoding/json"
	"gin-quickstart/internal/model"
	"gin-quickstart/pkg/logger"
	"reflect"
	"strconv"
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

	getResult, err := r.GetCache(ctx, "user:username:"+username)

	if err == nil {
		r.log.Debug(ctx, "GetUserByUsername Repo Cache Hit", r.log.Field("Username", username))

		if len(getResult) == 0 {
			return nil, gorm.ErrRecordNotFound
		}

		user = getResult[0]

		return &user, nil
	}

	r.log.Debug(ctx, "GetUserByUsername Repo Cache Miss", r.log.Field("Username", username))

	fErr := r.GormDB.Where("username = ?", username).First(&user).Error
	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetUserByUsername Repo Cache Set", r.log.Field("Username", username))

	userJSON, mErr := json.Marshal(user)

	if mErr != nil {
		r.log.Error(ctx, "GetUserByUsername Repo Cache Marshal Error", mErr, r.log.Field("Username", username))
		return &user, nil
	}

	err = r.SetCache(ctx, "user:username:"+username, userJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetUserByUsername Repo Cache Set Error", err, r.log.Field("Username", username))
	}

	return &user, nil
}

func (r AuthRepository) GetUserById(ctx *gin.Context, id uint64) (*model.User, error) {
	var user model.User

	getResult, err := r.GetCache(ctx, "user:id:"+strconv.FormatUint(id, 10))

	if err == nil {
		r.log.Debug(ctx, "GetUserById Repo Cache Hit", r.log.Field("UserID", id))

		if len(getResult) == 0 {
			return nil, gorm.ErrRecordNotFound
		}

		user = getResult[0]

		return &user, nil
	}

	r.log.Debug(ctx, "GetUserById Repo Cache Miss", r.log.Field("UserID", id))

	fErr := r.GormDB.First(&user, id).Error
	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetUserById Repo Cache Set", r.log.Field("UserID", id))

	userJSON, mErr := json.Marshal(user)

	if mErr != nil {
		r.log.Error(ctx, "GetUserById Repo Cache Marshal Error", mErr, r.log.Field("UserID", id))
		return &user, nil
	}

	err = r.SetCache(ctx, "user:id:"+strconv.FormatUint(id, 10), userJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetUserById Repo Cache Set Error", err, r.log.Field("UserID", id))
	}

	return &user, nil
}

func (r AuthRepository) GetUserByEmail(ctx *gin.Context, email string) (*model.User, error) {
	var user model.User

	getResult, err := r.GetCache(ctx, "user:email:"+email)

	if err == nil {
		r.log.Debug(ctx, "GetUserByEmail Repo Cache Hit", r.log.Field("Email", email))

		if len(getResult) == 0 {
			return nil, gorm.ErrRecordNotFound
		}

		user = getResult[0]

		return &user, nil
	}

	r.log.Debug(ctx, "GetUserByEmail Repo Cache Miss", r.log.Field("Email", email))

	fErr := r.GormDB.Where("email = ?", email).First(&user).Error
	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetUserByEmail Repo Cache Set", r.log.Field("Email", email))

	userJSON, mErr := json.Marshal(user)

	if mErr != nil {
		r.log.Error(ctx, "GetUserByEmail Repo Cache Marshal Error", mErr, r.log.Field("Email", email))
		return &user, nil
	}

	err = r.SetCache(ctx, "user:email:"+email, userJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetUserByEmail Repo Cache Set Error", err, r.log.Field("Email", email))
	}

	return &user, nil
}

func (r AuthRepository) GetCache(ctx context.Context, key string) ([]model.User, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.User

	if getResult.Err() != nil {
		r.log.Error(ctx, "Repo GetCache Error", getResult.Err(), r.log.Field("Key", key))
		return nil, getResult.Err()
	}

	err := json.Unmarshal([]byte(getResult.Val()), &result)

	if err != nil {
		r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
		return nil, err
	}

	if r.isSlice(result) {
		var users []model.User

		err := json.Unmarshal([]byte(getResult.Val()), &users)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return users, nil
	}

	if !r.isSlice(result) {
		var user model.User

		jsonR, err := json.Marshal(result)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Marshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		err = json.Unmarshal(jsonR, &user)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		returns = append(returns, user)
	}

	return returns, nil
}

// SETTER
func (r *AuthRepository) Register(ctx *gin.Context, user *model.User) error {

	cErr := r.GormDB.Create(user).Error

	if cErr != nil {
		return cErr
	}

	err := r.DeleteCache(ctx, "user:all")

	if err != nil {
		r.log.Error(ctx, "Repo Register DeleteCache Error", err, r.log.Field("Key", "user:all"))
	}

	err = r.DeleteCache(ctx, "user:username:"+user.Username)

	if err != nil {
		r.log.Error(ctx, "Repo Register DeleteCache Error", err, r.log.Field("Key", "user:username:"+user.Username))
	}

	err = r.DeleteCache(ctx, "user:email:"+user.Email)

	if err != nil {
		r.log.Error(ctx, "Repo Register DeleteCache Error", err, r.log.Field("Key", "user:email:"+user.Email))
	}

	return nil
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

func (r *AuthRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *AuthRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r AuthRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
