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

type BadgeRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewBadgeRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *BadgeRepository {
	return &BadgeRepository{
		log:         log,
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r BadgeRepository) GetAllBadges(ctx *gin.Context) ([]model.Badge, error) {
	var badges []model.Badge

	getResult, err := r.GetCache(ctx, "badge:all")

	if err == nil {
		r.log.Debug(ctx, "GetAllBadges Repo Cache Hit")
		badges = getResult
		return badges, nil
	}

	fErr := r.GormDB.Find(&badges).Error

	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetAllBadges Repo Cache Set", r.log.Field("Count", len(badges)))

	badgesJSON, mErr := json.Marshal(badges)

	if mErr != nil {
		r.log.Error(ctx, "GetAllBadges Repo Cache Marshal Error", mErr)
		return badges, nil
	}

	err = r.SetCache(ctx, "badge:all", badgesJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetAllBadges Repo Cache Set Error", err)
	}

	return badges, nil
}

func (r BadgeRepository) GetBadgeByID(ctx *gin.Context, id uint64) (*model.Badge, error) {
	var badge model.Badge

	getResult, err := r.GetCache(ctx, "badge:id:"+strconv.FormatUint(id, 10))

	if err == nil {
		r.log.Debug(ctx, "GetBadgeByID Repo Cache Hit", r.log.Field("ID", id))

		if len(getResult) == 0 {
			return nil, gorm.ErrRecordNotFound
		}

		badge = getResult[0]

		return &badge, nil
	}

	r.log.Debug(ctx, "GetBadgeByID Repo Cache Miss", r.log.Field("ID", id))

	fErr := r.GormDB.First(&badge, id).Error
	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetBadgeByID Repo Cache Set", r.log.Field("ID", id))

	badgeJSON, mErr := json.Marshal(badge)

	if mErr != nil {
		r.log.Error(ctx, "GetBadgeByID Repo Cache Marshal Error", mErr, r.log.Field("ID", id))
		return &badge, nil
	}

	err = r.SetCache(ctx, "badge:id:"+strconv.FormatUint(id, 10), badgeJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetBadgeByID Repo Cache Set Error", err, r.log.Field("ID", id))
	}

	return &badge, nil
}

func (r BadgeRepository) GetCache(ctx context.Context, key string) ([]model.Badge, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.Badge

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
		var users []model.Badge

		err := json.Unmarshal([]byte(getResult.Val()), &users)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return users, nil
	}

	if !r.isSlice(result) {
		var user model.Badge

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
func (r *BadgeRepository) Create(ctx *gin.Context, badge *model.Badge) error {

	cErr := r.GormDB.Create(badge).Error

	if cErr != nil {
		return cErr
	}

	// Invalidate Cache
	err := r.DeleteCache(ctx, "badge:all")

	if err != nil {
		r.log.Error(ctx, "Create Badge Repo Cache Delete Error", err)
	}

	return nil
}

func (r *BadgeRepository) Update(ctx *gin.Context, badge *model.Badge) error {
	uErr := r.GormDB.Save(badge).Error

	if uErr != nil {
		return uErr
	}

	// Invalidate Cache
	err := r.DeleteCache(ctx, "badge:all")

	if err != nil {
		r.log.Error(ctx, "Update Badge Repo Cache Delete Error", err)
	}

	err = r.DeleteCache(ctx, "badge:id:"+strconv.FormatUint(uint64(badge.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Update Badge Repo Cache Delete Error", err, r.log.Field("ID", badge.ID))
	}

	return nil
}

func (r *BadgeRepository) Delete(ctx *gin.Context, badge *model.Badge) error {
	dErr := r.GormDB.Delete(badge).Error

	if dErr != nil {
		return dErr
	}

	// Invalidate Cache
	err := r.DeleteCache(ctx, "badge:all")

	if err != nil {
		r.log.Error(ctx, "Delete Badge Repo Cache Delete Error", err)
	}

	err = r.DeleteCache(ctx, "badge:id:"+strconv.FormatUint(uint64(badge.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Delete Badge Repo Cache Delete Error", err, r.log.Field("ID", badge.ID))
	}

	return nil
}

func (r *BadgeRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *BadgeRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r BadgeRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
