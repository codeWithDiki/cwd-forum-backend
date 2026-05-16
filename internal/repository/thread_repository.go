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

type ThreadRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewThreadRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *ThreadRepository {
	return &ThreadRepository{
		log:         log,
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r ThreadRepository) GetAllThreads(ctx *gin.Context) ([]model.Thread, error) {
	var threads []model.Thread

	getResult, err := r.GetCache(ctx, "thread:all")

	if err == nil {
		r.log.Debug(ctx, "GetAllThreads Repo Cache Hit")
		threads = getResult
		return threads, nil
	}

	r.log.Debug(ctx, "GetAllThreads Repo Cache Miss")

	err = r.GormDB.
		Preload("Category").
		Preload("Posts").
		Preload("Tags").
		Preload("Author").
		Preload("PinnedByUser").
		Find(&threads).Error

	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetAllThreads Repo Cache Set", r.log.Field("Count", len(threads)))

	threadsJSON, mErr := json.Marshal(threads)

	if mErr != nil {
		r.log.Error(ctx, "GetAllThreads Repo Cache Marshal Error", mErr)
		return threads, nil
	}

	err = r.SetCache(ctx, "thread:all", threadsJSON, 10*time.Minute)

	if err != nil {
		r.log.Error(ctx, "GetAllThreads Repo Cache Set Error", err)
	}

	return threads, nil
}

func (r ThreadRepository) GetThreadByID(ctx *gin.Context, id uint64) (*model.Thread, error) {
	var thread model.Thread

	getResult, err := r.GetCache(ctx, "thread:"+strconv.FormatUint(id, 10))

	if err == nil {
		r.log.Debug(ctx, "GetThreadByID Repo Cache Hit", r.log.Field("ID", id))
		if len(getResult) > 0 {
			return &getResult[0], nil
		}
	}

	r.log.Debug(ctx, "GetThreadByID Repo Cache Miss", r.log.Field("ID", id))

	err = r.GormDB.
		Preload("Category").
		Preload("Posts").
		Preload("Tags").
		Preload("Author").
		Preload("PinnedByUser").
		First(&thread, id).
		Error

	if err != nil {
		return nil, err
	}

	threadJSON, mErr := json.Marshal(thread)

	if mErr != nil {
		r.log.Error(ctx, "GetThreadByID Repo Cache Marshal Error", mErr, r.log.Field("ID", id))
		return &thread, nil
	}

	cErr := r.SetCache(ctx, "thread:"+strconv.FormatUint(id, 10), threadJSON, 10*time.Minute)

	if cErr != nil {
		r.log.Error(ctx, "GetThreadByID Repo Cache Set Error", cErr, r.log.Field("ID", id))
	}

	return &thread, nil
}

func (r ThreadRepository) GetThreadBySlug(ctx *gin.Context, slug string) (*model.Thread, error) {
	var thread model.Thread

	getResult, err := r.GetCache(ctx, "thread:slug:"+slug)

	if err == nil {
		r.log.Debug(ctx, "GetThreadBySlug Repo Cache Hit", r.log.Field("Slug", slug))
		if len(getResult) > 0 {
			return &getResult[0], nil
		}
	}

	r.log.Debug(ctx, "GetThreadBySlug Repo Cache Miss", r.log.Field("Slug", slug))

	err = r.GormDB.Where("slug = ?", slug).First(&thread).Error
	if err != nil {
		return nil, err
	}

	threadJSON, mErr := json.Marshal(thread)

	if mErr != nil {
		r.log.Error(ctx, "GetThreadBySlug Repo Cache Marshal Error", mErr, r.log.Field("Slug", slug))
		return &thread, nil
	}

	cErr := r.SetCache(ctx, "thread:slug:"+slug, threadJSON, 10*time.Minute)

	if cErr != nil {
		r.log.Error(ctx, "GetThreadBySlug Repo Cache Set Error", cErr, r.log.Field("Slug", slug))
	}

	return &thread, nil
}

func (r ThreadRepository) GetThreadsByCategoryID(ctx *gin.Context, categoryID uint) ([]model.Thread, error) {
	var threads []model.Thread

	getResult, err := r.GetCache(ctx, "thread:category:"+strconv.FormatUint(uint64(categoryID), 10))

	if err == nil {
		r.log.Debug(ctx, "GetThreadsByCategoryID Repo Cache Hit", r.log.Field("CategoryID", categoryID))
		threads = getResult
		return threads, nil
	}

	r.log.Debug(ctx, "GetThreadsByCategoryID Repo Cache Miss", r.log.Field("CategoryID", categoryID))

	err = r.GormDB.Where("category_id = ?", categoryID).Find(&threads).Error
	if err != nil {
		return nil, err
	}

	cmdStatus := r.SetCache(ctx, "thread:category:"+strconv.FormatUint(uint64(categoryID), 10), []byte{}, 10*time.Minute)

	if cmdStatus != nil {
		r.log.Error(ctx, "GetThreadsByCategoryID Repo Cache Set Error", cmdStatus, r.log.Field("CategoryID", categoryID))
	}

	return threads, nil
}

func (r ThreadRepository) GetThreadsByAuthorID(ctx *gin.Context, authorID uint) ([]model.Thread, error) {
	var threads []model.Thread

	getStatus, err := r.GetCache(ctx, "thread:author:"+strconv.FormatUint(uint64(authorID), 10))

	if err == nil {
		r.log.Debug(ctx, "GetThreadsByAuthorID Repo Cache Hit", r.log.Field("AuthorID", authorID))
		threads = getStatus
		return threads, nil
	}

	r.log.Debug(ctx, "GetThreadsByAuthorID Repo Cache Miss", r.log.Field("AuthorID", authorID))

	err = r.GormDB.Where("author_id = ?", authorID).Find(&threads).Error
	if err != nil {
		return nil, err
	}

	cmdStatus := r.SetCache(ctx, "thread:author:"+strconv.FormatUint(uint64(authorID), 10), []byte{}, 10*time.Minute)

	if cmdStatus != nil {
		r.log.Error(ctx, "GetThreadsByAuthorID Repo Cache Set Error", cmdStatus, r.log.Field("AuthorID", authorID))
	}

	return threads, nil
}

func (r ThreadRepository) GetThreadsByTagID(ctx *gin.Context, tagID uint) ([]model.Thread, error) {
	var threads []model.Thread

	getResult, err := r.GetCache(ctx, "thread:tag:"+strconv.FormatUint(uint64(tagID), 10))

	if err == nil {
		r.log.Debug(ctx, "GetThreadsByTagID Repo Cache Hit", r.log.Field("TagID", tagID))
		threads = getResult
		return threads, nil
	}

	r.log.Debug(ctx, "GetThreadsByTagID Repo Cache Miss", r.log.Field("TagID", tagID))

	err = r.GormDB.
		Joins("JOIN thread_tags ON thread_tags.thread_id = threads.id").
		Where("thread_tags.tag_id = ?", tagID).
		Find(&threads).Error

	if err != nil {
		return nil, err
	}

	cmdStatus := r.SetCache(ctx, "thread:tag:"+strconv.FormatUint(uint64(tagID), 10), []byte{}, 10*time.Minute)

	if cmdStatus != nil {
		r.log.Error(ctx, "GetThreadsByTagID Repo Cache Set Error", cmdStatus, r.log.Field("TagID", tagID))
	}

	return threads, nil
}
func (r ThreadRepository) GetCache(ctx context.Context, key string) ([]model.Thread, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.Thread

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
		var users []model.Thread

		err := json.Unmarshal([]byte(getResult.Val()), &users)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return users, nil
	}

	if !r.isSlice(result) {
		var user model.Thread

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
func (r *ThreadRepository) Create(ctx *gin.Context, thread *model.Thread) error {
	if err := r.GormDB.Create(thread).Error; err != nil {
		return err
	}

	err := r.DeleteCache(ctx, "thread:all")

	if err != nil {
		r.log.Error(ctx, "Create Thread Repo Cache Delete Error", err)
	}

	return nil
}

func (r *ThreadRepository) Update(ctx *gin.Context, thread *model.Thread) error {
	err := r.GormDB.Save(thread).Error

	if err != nil {
		return err
	}

	err = r.DeleteCache(ctx, "thread:all")

	if err != nil {
		r.log.Error(ctx, "Update Thread Repo Cache Delete Error", err)
	}

	err = r.DeleteCache(ctx, "thread:"+strconv.FormatUint(uint64(thread.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Update Thread Repo Cache Delete Error", err, r.log.Field("ID", thread.ID))
	}

	err = r.DeleteCache(ctx, "thread:slug:"+thread.Slug)

	if err != nil {
		r.log.Error(ctx, "Update Thread Repo Cache Delete Error", err, r.log.Field("Slug", thread.Slug))
	}

	return nil
}

func (r *ThreadRepository) Delete(ctx *gin.Context, thread *model.Thread) error {
	err := r.GormDB.Delete(thread).Error

	if err != nil {
		return err
	}

	err = r.DeleteCache(ctx, "thread:all")

	if err != nil {
		r.log.Error(ctx, "Delete Thread Repo Cache Delete Error", err)
	}

	err = r.DeleteCache(ctx, "thread:"+strconv.FormatUint(uint64(thread.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Delete Thread Repo Cache Delete Error", err, r.log.Field("ID", thread.ID))
	}

	err = r.DeleteCache(ctx, "thread:slug:"+thread.Slug)

	if err != nil {
		r.log.Error(ctx, "Delete Thread Repo Cache Delete Error", err, r.log.Field("Slug", thread.Slug))
	}

	return nil
}

func (r *ThreadRepository) CreatePostAttachment(
	ctx *gin.Context,
	post *model.Post,
	attachment *model.Attachment,
) error {
	if err := r.GormDB.Create(attachment).Error; err != nil {
		return err
	}

	if err := r.GormDB.Save(post).Error; err != nil {
		return err
	}

	if err := r.GormDB.Save(attachment).Error; err != nil {
		return err
	}

	return nil
}

func (r *ThreadRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *ThreadRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r ThreadRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
