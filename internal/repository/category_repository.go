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

type CategoryRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewCategoryRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *CategoryRepository {
	return &CategoryRepository{
		log:         log,
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r CategoryRepository) GetAllCategories(ctx *gin.Context) ([]model.Category, error) {
	var categories []model.Category

	getResult, err := r.GetCache(ctx, "category:all")

	if err == nil {
		r.log.Debug(ctx, "GetAllCategories Repo Cache Hit")
		categories = getResult
		return categories, nil
	}

	r.log.Debug(ctx, "GetAllCategories Repo Cache Miss")

	fErr := r.GormDB.
		Preload("Categories").
		Preload("Threads").
		Find(&categories).Error

	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetAllCategories Repo Cache Set", r.log.Field("Count", len(categories)))

	categoriesJSON, mErr := json.Marshal(categories)

	if mErr != nil {
		r.log.Error(ctx, "GetAllCategories Repo Cache Marshal Error", mErr)
		return categories, nil
	}

	err = r.SetCache(ctx, "category:all", categoriesJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetAllCategories Repo Cache Set Error", err)
		return categories, nil
	}

	return categories, nil
}

func (r CategoryRepository) GetCategoryByID(ctx *gin.Context, id uint64) (*model.Category, error) {
	var category model.Category

	getResult, err := r.GetCache(ctx, "category:id:"+strconv.FormatUint(id, 10))

	if err == nil {
		r.log.Debug(ctx, "GetCategoryByID Repo Cache Hit", r.log.Field("ID", id))

		if len(getResult) == 0 {
			return nil, gorm.ErrRecordNotFound
		}

		category = getResult[0]

		return &category, nil
	}

	r.log.Debug(ctx, "GetCategoryByID Repo Cache Miss", r.log.Field("ID", id))

	fErr := r.GormDB.
		Preload("Categories").
		Preload("Threads").
		First(&category, id).Error

	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetCategoryByID Repo Cache Set", r.log.Field("ID", id))

	categoryJSON, mErr := json.Marshal(category)

	if mErr != nil {
		r.log.Error(ctx, "GetCategoryByID Repo Cache Marshal Error", mErr, r.log.Field("ID", id))
		return &category, nil
	}

	err = r.SetCache(ctx, "category:id:"+strconv.FormatUint(id, 10), categoryJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetCategoryByID Repo Cache Set Error", err, r.log.Field("ID", id))
		return &category, nil
	}

	return &category, nil
}

func (r CategoryRepository) GetCategoryBySlug(ctx *gin.Context, slug string) (*model.Category, error) {
	var category model.Category

	getResult, err := r.GetCache(ctx, "category:slug:"+slug)

	if err == nil {
		r.log.Debug(ctx, "GetCategoryBySlug Repo Cache Hit", r.log.Field("Slug", slug))

		if len(getResult) == 0 {
			return nil, gorm.ErrRecordNotFound
		}

		category = getResult[0]

		return &category, nil
	}

	r.log.Debug(ctx, "GetCategoryBySlug Repo Cache Miss", r.log.Field("Slug", slug))

	fErr := r.GormDB.
		Preload("Categories").
		Preload("Threads").
		Where("slug = ?", slug).First(&category).Error
	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetCategoryBySlug Repo Cache Set", r.log.Field("Slug", slug))

	categoryJSON, mErr := json.Marshal(category)

	if mErr != nil {
		r.log.Error(ctx, "GetCategoryBySlug Repo Cache Marshal Error", mErr, r.log.Field("Slug", slug))
		return &category, nil
	}

	err = r.SetCache(ctx, "category:slug:"+slug, categoryJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetCategoryBySlug Repo Cache Set Error", err, r.log.Field("Slug", slug))
		return &category, nil
	}

	return &category, nil
}

func (r CategoryRepository) GetCache(ctx context.Context, key string) ([]model.Category, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.Category

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
		var users []model.Category

		err := json.Unmarshal([]byte(getResult.Val()), &users)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return users, nil
	}

	if !r.isSlice(result) {
		var user model.Category

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
func (r *CategoryRepository) Create(ctx *gin.Context, category *model.Category) error {
	if err := r.GormDB.Create(category).Error; err != nil {
		r.log.Error(ctx, "Repo Create Category Error", err)
		return err
	}

	r.log.Debug(ctx, "Repo Create Category Success", r.log.Field("ID", category.ID))

	err := r.DeleteCache(ctx, "category:all")

	if err != nil {
		r.log.Error(ctx, "Repo Create Category Cache Invalidate Error", err)
		return err
	}

	return nil
}

func (r *CategoryRepository) Update(ctx *gin.Context, category *model.Category) error {
	err := r.GormDB.Save(category).Error

	if err != nil {
		r.log.Error(ctx, "Repo Update Category Error", err, r.log.Field("ID", category.ID))
		return err
	}

	r.log.Debug(ctx, "Repo Update Category Success", r.log.Field("ID", category.ID))

	err = r.DeleteCache(ctx, "category:all")

	if err != nil {
		r.log.Error(ctx, "Repo Update Category Cache Invalidate Error", err)
		return err
	}

	err = r.DeleteCache(ctx, "category:id:"+strconv.FormatUint(uint64(category.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo Update Category Cache Invalidate Error", err, r.log.Field("ID", category.ID))
		return err
	}

	err = r.DeleteCache(ctx, "category:slug:"+category.Slug)

	if err != nil {
		r.log.Error(ctx, "Repo Update Category Cache Invalidate Error", err, r.log.Field("Slug", category.Slug))
		return err
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx *gin.Context, category *model.Category) error {
	err := r.GormDB.Delete(category).Error

	if err != nil {
		r.log.Error(ctx, "Repo Delete Category Error", err, r.log.Field("ID", category.ID))
		return err
	}

	r.log.Debug(ctx, "Repo Delete Category Success", r.log.Field("ID", category.ID))

	err = r.DeleteCache(ctx, "category:all")

	if err != nil {
		r.log.Error(ctx, "Repo Delete Category Cache Invalidate Error", err)
		return err
	}

	err = r.DeleteCache(ctx, "category:id:"+strconv.FormatUint(uint64(category.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo Delete Category Cache Invalidate Error", err, r.log.Field("ID", category.ID))
		return err
	}

	err = r.DeleteCache(ctx, "category:slug:"+category.Slug)

	if err != nil {
		r.log.Error(ctx, "Repo Delete Category Cache Invalidate Error", err, r.log.Field("Slug", category.Slug))
		return err
	}

	// Invalidate related thread caches
	for _, thread := range category.Threads {
		err = r.DeleteCache(ctx, "thread:id:"+strconv.FormatUint(uint64(thread.ID), 10))

		if err != nil {
			r.log.Error(ctx, "Repo Delete Category Related Thread Cache Invalidate Error", err, r.log.Field("ThreadID", thread.ID))
			return err
		}
	}

	return nil
}

func (r *CategoryRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *CategoryRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r CategoryRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
