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

type NotificationRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewNotificationRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *NotificationRepository {
	return &NotificationRepository{
		log:         log,
		GormDB:      db,
		RedisClient: redis,
	}
}

func (r NotificationRepository) GetNotificationsByUserID(ctx *gin.Context, userID uint64) ([]model.Notification, error) {
	var notifications []model.Notification

	getResult, err := r.GetCache(ctx, "notifications:user:"+strconv.FormatUint(userID, 10))

	if err == nil {
		r.log.Debug(ctx, "GetNotificationsByUserID Repo Cache Hit", r.log.Field("UserID", userID))
		notifications = getResult
		return notifications, nil
	}

	r.log.Debug(ctx, "GetNotificationsByUserID Repo Cache Miss", r.log.Field("UserID", userID))

	fErr := r.GormDB.Where("user_id = ?", userID).
		Order("created_at desc").
		Preload("User").
		Preload("Thread").
		Preload("Post").
		Find(&notifications).Error
	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetNotificationsByUserID Repo Cache Set", r.log.Field("UserID", userID), r.log.Field("Count", len(notifications)))

	notificationsJSON, mErr := json.Marshal(notifications)

	if mErr != nil {
		r.log.Error(ctx, "GetNotificationsByUserID Repo Cache Marshal Error", mErr, r.log.Field("UserID", userID))
		return notifications, nil
	}

	err = r.SetCache(ctx, "notifications:user:"+strconv.FormatUint(userID, 10), notificationsJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetNotificationsByUserID Repo Cache Set Error", err, r.log.Field("UserID", userID))
		return notifications, nil
	}

	return notifications, nil
}

func (r NotificationRepository) GetNotificationByID(ctx *gin.Context, id uint64) (*model.Notification, error) {
	var notification model.Notification

	getResult, err := r.GetCache(ctx, "notification:"+strconv.FormatUint(id, 10))

	if err == nil {
		r.log.Debug(ctx, "GetNotificationByID Repo Cache Hit", r.log.Field("ID", id))

		if len(getResult) == 0 {
			return nil, gorm.ErrRecordNotFound
		}

		notification = getResult[0]

		return &notification, nil
	}

	r.log.Debug(ctx, "GetNotificationByID Repo Cache Miss", r.log.Field("ID", id))

	fErr := r.GormDB.Preload("User").Preload("Thread").Preload("Post").First(&notification, id).Error
	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetNotificationByID Repo Cache Set", r.log.Field("ID", id))

	notificationJSON, mErr := json.Marshal(notification)

	if mErr != nil {
		r.log.Error(ctx, "GetNotificationByID Repo Cache Marshal Error", mErr, r.log.Field("ID", id))
		return &notification, nil
	}

	err = r.SetCache(ctx, "notification:"+strconv.FormatUint(id, 10), notificationJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetNotificationByID Repo Cache Set Error", err, r.log.Field("ID", id))
	}

	return &notification, nil
}

func (r NotificationRepository) GetCache(ctx context.Context, key string) ([]model.Notification, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.Notification

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
		var users []model.Notification

		err := json.Unmarshal([]byte(getResult.Val()), &users)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return users, nil
	}

	if !r.isSlice(result) {
		var user model.Notification

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

func (r *NotificationRepository) Create(ctx *gin.Context, notification *model.Notification) error {
	err := r.GormDB.Create(notification).Error

	if err != nil {
		return err
	}

	dErr := r.DeleteCache(ctx, "notifications:user:"+strconv.FormatUint(uint64(notification.UserId), 10))

	if dErr != nil {
		r.log.Error(ctx, "Create Notification Repo Cache Delete Error", dErr, r.log.Field("UserID", notification.UserId))
	}

	return nil
}

func (r *NotificationRepository) Update(ctx *gin.Context, notification *model.Notification) error {
	err := r.GormDB.Save(notification).Error

	if err != nil {
		return err
	}

	dErr := r.DeleteCache(ctx, "notification:"+strconv.FormatUint(uint64(notification.ID), 10))

	if dErr != nil {
		r.log.Error(ctx, "Update Notification Repo Cache Delete Error", dErr, r.log.Field("ID", notification.ID))
	}

	dErr = r.DeleteCache(ctx, "notifications:user:"+strconv.FormatUint(uint64(notification.UserId), 10))

	if dErr != nil {
		r.log.Error(ctx, "Update Notification Repo Cache Delete Error", dErr, r.log.Field("UserID", notification.UserId))
	}

	return nil
}

func (r *NotificationRepository) Delete(ctx *gin.Context, notification *model.Notification) error {
	err := r.GormDB.Delete(notification).Error

	if err != nil {
		return err
	}

	dErr := r.DeleteCache(ctx, "notification:"+strconv.FormatUint(uint64(notification.ID), 10))

	if dErr != nil {
		r.log.Error(ctx, "Delete Notification Repo Cache Delete Error", dErr, r.log.Field("ID", notification.ID))
	}

	dErr = r.DeleteCache(ctx, "notifications:user:"+strconv.FormatUint(uint64(notification.UserId), 10))

	if dErr != nil {
		r.log.Error(ctx, "Delete Notification Repo Cache Delete Error", dErr, r.log.Field("UserID", notification.UserId))
	}

	return nil
}

func (r *NotificationRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *NotificationRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r NotificationRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
