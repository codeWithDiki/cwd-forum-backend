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
	"gorm.io/gorm"
)

type UserRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewUserRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *UserRepository {
	return &UserRepository{
		log:         log,
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r UserRepository) GetAllUsers(ctx context.Context) ([]model.User, error) {
	r.log.Debug(ctx, "GetAllUser Repo Called")
	var users []model.User

	getResult, err := r.GetCache(ctx, "user:all")

	if err == nil {
		r.log.Debug(ctx, "GetAllUser Repo Cache Hit")
		users := getResult

		return users, nil
	}

	fErr := r.GormDB.Find(&users).Error

	if fErr != nil {
		r.log.Error(ctx, "GetAllUser Repo Error", fErr)
		return nil, fErr
	}

	json, err := json.Marshal(users)

	if err != nil {
		r.log.Error(ctx, "GetAllUser Cache Marshal Error", err)
		return nil, err
	}

	r.SetCache(ctx, "user:all", json, time.Hour*24)

	return users, nil
}

func (r UserRepository) GetUserByID(ctx *gin.Context, id uint64) (*model.User, error) {
	var user model.User
	key := "user:" + strconv.FormatUint(id, 10)

	getResult, err := r.GetCache(ctx, key)

	if err == nil {
		user = getResult[0]

		r.log.Debug(ctx, "GetUserByID Repo Cache Hit", r.log.Field("UserID", id))
		return &user, nil
	}

	fErr := r.GormDB.First(&user, id).Error

	if fErr != nil {
		return nil, fErr
	}

	json, err := json.Marshal(user)

	if err != nil {
		r.log.Error(ctx, "GetUserByID Cache Marshal Error", err, r.log.Field("UserID", id))
		return nil, err
	}

	sErr := r.SetCache(ctx, key, json, time.Hour*24)

	if sErr != nil {
		r.log.Error(ctx, "GetUserByID Cache Set Error", sErr, r.log.Field("UserID", id))
		return nil, sErr
	}

	return &user, nil
}

func (r UserRepository) GetUserByUsername(ctx *gin.Context, username string) (*model.User, error) {
	var user model.User
	key := "user:username:" + username
	r.log.Debug(ctx, "Repo GetUserByUsername Called", r.log.Field("Username", username))

	getResult, err := r.GetCache(ctx, key)

	if err == nil {
		user := getResult[0]

		r.log.Debug(ctx, "GetUserByUsername Repo Cache Hit", r.log.Field("Username", username))
		return &user, nil
	}

	fErr := r.GormDB.Where("username = ?", username).First(&user).Error

	if fErr != nil {
		r.log.Error(ctx, "Repo GetUserByUsername Error", fErr, r.log.Field("Username", username))
		return nil, fErr
	}

	json, err := json.Marshal(user)

	if err != nil {
		r.log.Error(ctx, "GetUserByUsername Cache Marshal Error", err, r.log.Field("Username", username))
		return nil, err
	}

	sErr := r.SetCache(ctx, key, json, time.Hour*24)

	if sErr != nil {
		r.log.Error(ctx, "GetUserByUsername Cache Set Error", sErr, r.log.Field("Username", username))
		return nil, sErr
	}

	return &user, nil
}

func (r UserRepository) GetUserByEmail(ctx *gin.Context, email string) (*model.User, error) {
	var user model.User
	getResult, err := r.GetCache(ctx, "user:email:"+email)
	r.log.Debug(ctx, "Repo GetUserByEmail Called", r.log.Field("Email", email))

	if err == nil {
		user = getResult[0]

		r.log.Debug(ctx, "Repo GetUserByEmail Cache Hit", r.log.Field("Email", email))
		return &user, nil
	}

	fErr := r.GormDB.Where("email = ?", email).First(&user).Error

	if fErr != nil {
		r.log.Error(ctx, "Repo GetUserByEmail Error", fErr, r.log.Field("Email", email))
		return nil, err
	}

	json, err := json.Marshal(user)

	if err != nil {
		r.log.Error(ctx, "Repo GetUserByEmail Cache Marshal Error", err, r.log.Field("Email", email))
		return nil, err
	}

	sErr := r.SetCache(ctx, "user:email:"+email, json, time.Hour*24)

	if sErr != nil {
		r.log.Error(ctx, "Repo GetUserByEmail Cache Set Error", sErr, r.log.Field("Email", email))
		return nil, sErr
	}

	return &user, nil
}

func (r UserRepository) GetFollowers(ctx *gin.Context, userID uint64) ([]model.User, error) {
	var followers []model.User
	r.log.Debug(ctx, "Repo GetFollowers Called", r.log.Field("UserID", userID))

	getResult, err := r.GetCache(ctx, "user:"+strconv.FormatUint(userID, 10)+":followers")

	if err == nil {
		followers = getResult

		r.log.Debug(ctx, "Repo GetFollowers Cache Hit", r.log.Field("UserID", userID))
		return followers, nil
	}

	fErr := r.GormDB.Joins("JOIN user_users ON user_users.follower_id = users.id").
		Where("user_users.followed_id = ?", userID).
		Find(&followers).Error

	if fErr != nil {
		r.log.Error(ctx, "Repo GetFollowers Error", fErr, r.log.Field("UserID", userID))

		return nil, err
	}

	json, err := json.Marshal(followers)

	if err != nil {
		r.log.Error(ctx, "Repo GetFollowers Cache Marshal Error", err, r.log.Field("UserID", userID))
		return nil, err
	}

	sErr := r.SetCache(ctx, "user:"+strconv.FormatUint(userID, 10)+":followers", json, time.Hour*24)

	if sErr != nil {
		r.log.Error(ctx, "Repo GetFollowers Cache Set Error", sErr, r.log.Field("UserID", userID))
		return nil, sErr
	}

	return followers, nil
}

func (r UserRepository) GetFollowing(ctx *gin.Context, userID uint64) ([]model.User, error) {
	var following []model.User
	r.log.Debug(ctx, "Repo GetFollowing Called", r.log.Field("UserID", userID))

	getResult := r.RedisClient.Get(ctx, "user:"+strconv.FormatUint(userID, 10)+":following")

	if getResult.Err() == nil {
		err := json.Unmarshal([]byte(getResult.Val()), &following)

		if err != nil {
			r.log.Error(ctx, "Repo GetFollowing Cache Unmarshal Error", err, r.log.Field("UserID", userID))
			return nil, err
		}

		r.log.Debug(ctx, "Repo GetFollowing Cache Hit", r.log.Field("UserID", userID))
		return following, nil
	}

	err := r.GormDB.Joins("JOIN user_users ON user_users.followed_id = users.id").
		Where("user_users.follower_id = ?", userID).
		Find(&following).Error

	if err != nil {
		r.log.Error(ctx, "Repo GetFollowing Error", err, r.log.Field("UserID", userID))
		return nil, err
	}

	json, err := json.Marshal(following)

	if err != nil {
		r.log.Error(ctx, "Repo GetFollowing Cache Marshal Error", err, r.log.Field("UserID", userID))
		return nil, err
	}

	cmdStatus := r.RedisClient.Set(ctx, "user:"+strconv.FormatUint(userID, 10)+":following", json, time.Hour*24)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo GetFollowing Cache Set Error", cmdStatus.Err(), r.log.Field("UserID", userID))
		return nil, cmdStatus.Err()
	}

	return following, nil
}

func (r UserRepository) GetCache(ctx context.Context, key string) ([]model.User, error) {
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
func (r *UserRepository) Create(ctx *gin.Context, user *model.User) error {
	r.log.Debug(ctx, "Repo CreateUser Called", r.log.Field("Username", user.Username), r.log.Field("Email", user.Email))

	err := r.DeleteCache(ctx, "users")

	if err != nil {
		r.log.Error(ctx, "Repo CreateUser Cache Delete Error", err, r.log.Field("Username", user.Username), r.log.Field("Email", user.Email))
	}

	return r.GormDB.Create(user).Error
}

func (r *UserRepository) Update(ctx *gin.Context, user *model.User) error {
	r.log.Debug(ctx, "Repo UpdateUser Called", r.log.Field("UserID", user.ID))

	err := r.DeleteCache(ctx, "user:"+strconv.FormatUint(uint64(user.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo UpdateUser Cache Delete Error", err, r.log.Field("UserID", user.ID))
	}

	uErr := r.DeleteCache(ctx, "user:username:"+user.Username)

	if uErr != nil {
		r.log.Error(ctx, "Repo UpdateUser Cache Delete Error", uErr, r.log.Field("Username", user.Username))
	}

	eErr := r.DeleteCache(ctx, "user:email:"+user.Email)

	if eErr != nil {
		r.log.Error(ctx, "Repo UpdateUser Cache Delete Error", eErr, r.log.Field("Email", user.Email))
	}

	return r.GormDB.Save(user).Error
}

func (r *UserRepository) Delete(ctx *gin.Context, user *model.User) error {
	r.log.Debug(ctx, "Repo DeleteUser Called", r.log.Field("UserID", user.ID))
	user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

	err := r.DeleteCache(ctx, "user:"+strconv.FormatUint(uint64(user.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", err, r.log.Field("UserID", user.ID))
	}

	uErr := r.DeleteCache(ctx, "user:username:"+user.Username)

	if uErr != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", uErr, r.log.Field("Username", user.Username))
	}

	eErr := r.DeleteCache(ctx, "user:email:"+user.Email)

	if eErr != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", eErr, r.log.Field("Email", user.Email))
	}

	return r.GormDB.Save(user).Error
}

func (r *UserRepository) HardDelete(ctx *gin.Context, user *model.User) error {
	r.log.Debug(ctx, "Repo HardDeleteUser Called", r.log.Field("UserID", user.ID))

	err := r.DeleteCache(ctx, "user:"+strconv.FormatUint(uint64(user.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", err, r.log.Field("UserID", user.ID))
	}

	uErr := r.DeleteCache(ctx, "user:username:"+user.Username)

	if uErr != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", uErr, r.log.Field("Username", user.Username))
	}

	eErr := r.DeleteCache(ctx, "user:email:"+user.Email)

	if eErr != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", eErr, r.log.Field("Email", user.Email))
	}

	return r.GormDB.Unscoped().Delete(user).Error
}

func (r *UserRepository) Restore(ctx *gin.Context, user *model.User) error {
	r.log.Debug(ctx, "Repo RestoreUser Called", r.log.Field("UserID", user.ID))
	user.DeletedAt = gorm.DeletedAt{Time: time.Time{}, Valid: false}

	err := r.DeleteCache(ctx, "user:"+strconv.FormatUint(uint64(user.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", err, r.log.Field("UserID", user.ID))
	}

	uErr := r.DeleteCache(ctx, "user:username:"+user.Username)

	if uErr != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", uErr, r.log.Field("Username", user.Username))
	}

	eErr := r.DeleteCache(ctx, "user:email:"+user.Email)

	if eErr != nil {
		r.log.Error(ctx, "Repo DeleteUser Cache Delete Error", eErr, r.log.Field("Email", user.Email))
	}

	return r.GormDB.Save(user).Error
}

func (r *UserRepository) Follow(ctx *gin.Context, userID uint64, followerID uint64) error {
	r.log.Debug(ctx, "Repo Follow Called", r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))

	err := r.DeleteCache(ctx, "user:"+strconv.FormatUint(userID, 10)+":followers")

	if err != nil {
		r.log.Error(ctx, "Repo Follow Cache Delete Error", err, r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))
	}

	err = r.DeleteCache(ctx, "user:"+strconv.FormatUint(followerID, 10)+":following")

	if err != nil {
		r.log.Error(ctx, "Repo Follow Cache Delete Error", err, r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))
	}

	return r.GormDB.Exec("INSERT INTO user_users (follower_id, followed_id) VALUES (?, ?)", userID, followerID).Error
}

func (r *UserRepository) Unfollow(ctx *gin.Context, userID uint64, followerID uint64) error {
	r.log.Debug(ctx, "Repo Unfollow Called", r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))

	err := r.DeleteCache(ctx, "user:"+strconv.FormatUint(userID, 10)+":followers")

	if err != nil {
		r.log.Error(ctx, "Repo Follow Cache Delete Error", err, r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))
	}

	err = r.DeleteCache(ctx, "user:"+strconv.FormatUint(followerID, 10)+":following")

	if err != nil {
		r.log.Error(ctx, "Repo Follow Cache Delete Error", err, r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))
	}

	return r.GormDB.Exec("DELETE FROM user_users WHERE follower_id = ? AND followed_id = ?", userID, followerID).Error
}

func (r *UserRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *UserRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r UserRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func (r *UserRepository) IsFollowing(ctx *gin.Context, userID uint64, followerID uint64) (bool, error) {
	var count int64
	r.log.Debug(ctx, "Repo IsFollowing Called", r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))

	getResult, err := r.GetCache(ctx, "user:"+strconv.FormatUint(userID, 10)+":followers")

	if err == nil {
		followers := getResult

		for _, follower := range followers {
			if follower.ID == uint(followerID) {
				r.log.Debug(ctx, "Repo IsFollowing Cache Hit", r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))
				return true, nil
			}
		}

		r.log.Debug(ctx, "Repo IsFollowing Cache Hit - Not Following", r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))
		return false, nil
	}

	fErr := r.GormDB.Table("user_users").Where("follower_id = ? AND followed_id = ?", userID, followerID).Count(&count).Error

	if fErr != nil {
		r.log.Error(ctx, "Repo IsFollowing Error", fErr, r.log.Field("UserID", userID), r.log.Field("FollowerID", followerID))
		return false, fErr
	}

	return count > 0, nil
}
