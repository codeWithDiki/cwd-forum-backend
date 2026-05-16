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

type TagRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewTagRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *TagRepository {
	return &TagRepository{
		log:         log,
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r TagRepository) GetAllTags(ctx *gin.Context) ([]model.Tag, error) {
	var tags []model.Tag

	getResult, err := r.GetCache(ctx, "tag:all")

	if err == nil {
		r.log.Debug(ctx, "GetAllTags Repo Cache Hit")
		tags = getResult
		return tags, nil
	}

	r.log.Debug(ctx, "GetAllTags Repo Cache Miss")

	err = r.GormDB.
		Preload("Threads").
		Find(&tags).Error
	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetAllTags Repo Cache Set", r.log.Field("Count", len(tags)))

	tagsJSON, mErr := json.Marshal(tags)

	if mErr != nil {
		r.log.Error(ctx, "GetAllTags Repo Cache Marshal Error", mErr)
		return tags, nil
	}

	err = r.SetCache(ctx, "tag:all", tagsJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetAllTags Repo Cache Set Error", err)
	}

	return tags, nil
}

func (r TagRepository) GetTagByID(ctx *gin.Context, id uint64) (*model.Tag, error) {
	var tag model.Tag

	getResult, err := r.GetCache(ctx, "tag:id:"+strconv.FormatUint(id, 10))

	if err == nil {
		r.log.Debug(ctx, "GetTagByID Repo Cache Hit", r.log.Field("ID", id))
		tag = getResult[0]
		return &tag, nil
	}

	r.log.Debug(ctx, "GetTagByID Repo Cache Miss", r.log.Field("ID", id))

	err = r.GormDB.
		Preload("Threads").
		First(&tag, id).Error
	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetTagByID Repo Cache Set", r.log.Field("ID", id))

	tagJSON, mErr := json.Marshal(tag)

	if mErr != nil {
		r.log.Error(ctx, "GetTagByID Repo Cache Marshal Error", mErr, r.log.Field("ID", id))
		return &tag, nil
	}

	err = r.SetCache(ctx, "tag:id:"+strconv.FormatUint(id, 10), tagJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetTagByID Repo Cache Set Error", err, r.log.Field("ID", id))
		return &tag, nil
	}

	return &tag, nil
}

func (r TagRepository) GetTagBySlug(ctx *gin.Context, slug string) (*model.Tag, error) {
	var tag model.Tag

	getResult, err := r.GetCache(ctx, "tag:slug:"+slug)

	if err == nil {
		r.log.Debug(ctx, "GetTagBySlug Repo Cache Hit", r.log.Field("Slug", slug))
		tag = getResult[0]
		return &tag, nil
	}

	r.log.Debug(ctx, "GetTagBySlug Repo Cache Miss", r.log.Field("Slug", slug))

	err = r.GormDB.
		Preload("Threads").
		Where("slug = ?", slug).First(&tag).Error
	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetTagBySlug Repo Cache Set", r.log.Field("Slug", slug))

	tagJSON, mErr := json.Marshal(tag)

	if mErr != nil {
		r.log.Error(ctx, "GetTagBySlug Repo Cache Marshal Error", mErr, r.log.Field("Slug", slug))
		return &tag, nil
	}

	err = r.SetCache(ctx, "tag:slug:"+slug, tagJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetTagBySlug Repo Cache Set Error", err, r.log.Field("Slug", slug))
		return &tag, nil
	}

	return &tag, nil
}

func (r TagRepository) GetCache(ctx context.Context, key string) ([]model.Tag, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.Tag

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
		var users []model.Tag

		err := json.Unmarshal([]byte(getResult.Val()), &users)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return users, nil
	}

	if !r.isSlice(result) {
		var user model.Tag

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
func (r *TagRepository) Create(ctx *gin.Context, tag *model.Tag) error {
	err := r.GormDB.Create(tag).Error

	if err != nil {
		return err
	}

	r.log.Debug(ctx, "Repo Create Tag Success", r.log.Field("ID", tag.ID))

	err = r.DeleteCache(ctx, "tag:all")

	if err != nil {
		r.log.Error(ctx, "Repo Create Tag DeleteCache Error", err, r.log.Field("ID", tag.ID))
	}

	return nil
}

func (r *TagRepository) Update(ctx *gin.Context, tag *model.Tag) error {
	err := r.GormDB.Save(tag).Error

	if err != nil {
		return err
	}

	r.log.Debug(ctx, "Repo Update Tag Success", r.log.Field("ID", tag.ID))

	err = r.DeleteCache(ctx, "tag:all")

	if err != nil {
		r.log.Error(ctx, "Repo Update Tag DeleteCache Error", err, r.log.Field("ID", tag.ID))
	}

	return nil
}

func (r *TagRepository) Delete(ctx *gin.Context, id uint64) error {
	err := r.GormDB.Delete(&model.Tag{}, id).Error

	if err != nil {
		return err
	}

	r.log.Debug(ctx, "Repo Delete Tag Success", r.log.Field("ID", id))

	err = r.DeleteCache(ctx, "tag:all")

	if err != nil {
		r.log.Error(ctx, "Repo Delete Tag DeleteCache Error", err, r.log.Field("ID", id))
	}

	err = r.DeleteCache(ctx, "tag:id:"+strconv.FormatUint(id, 10))

	if err != nil {
		r.log.Error(ctx, "Repo Delete Tag DeleteCache Error", err, r.log.Field("ID", id))
	}

	return nil
}

func (r *TagRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *TagRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r TagRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
