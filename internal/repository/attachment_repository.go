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

var ctx = context.Background()

type AttachmentRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewAttachmentRepository(logger *logger.Logger, db *gorm.DB, redis *redis.Client) *AttachmentRepository {
	return &AttachmentRepository{
		log:         logger,
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER

func (r AttachmentRepository) GetAllAttachments(ctx *gin.Context) ([]model.Attachment, error) {
	var attachments []model.Attachment

	getResult, err := r.GetCache(ctx, "attachment:all")

	if err == nil {
		r.log.Debug(ctx, "GetAllAttachments Repo Cache Hit")
		attachments = getResult
		return attachments, nil
	}

	r.log.Debug(ctx, "GetAllAttachments Repo Cache Miss")

	fErr := r.GormDB.Find(&attachments).Error

	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetAllAttachments Repo Cache Set", r.log.Field("Count", len(attachments)))

	attachmentsJSON, mErr := json.Marshal(attachments)

	if mErr != nil {
		r.log.Error(ctx, "GetAllAttachments Repo Cache Marshal Error", mErr)
		return attachments, nil
	}

	cErr := r.SetCache(ctx, "attachment:all", attachmentsJSON, 10*time.Minute)

	if cErr != nil {
		r.log.Error(ctx, "GetAllAttachments Repo Cache Set Error", cErr)
	}

	return attachments, nil
}

func (r AttachmentRepository) GetAttachmentByID(ctx *gin.Context, id uint64) (*model.Attachment, error) {
	var attachment model.Attachment

	getResult, err := r.GetCache(ctx, "attachment:"+strconv.FormatUint(id, 10))

	if err == nil {
		r.log.Debug(ctx, "GetAttachmentByID Repo Cache Hit", r.log.Field("AttachmentID", id))
		if len(getResult) > 0 {
			return &getResult[0], nil
		}
		return nil, gorm.ErrRecordNotFound
	}

	r.log.Debug(ctx, "GetAttachmentByID Repo Cache Miss", r.log.Field("AttachmentID", id))

	fErr := r.GormDB.First(&attachment, id).Error

	if fErr != nil {
		r.log.Error(ctx, "GetAttachmentByID Repo DB Error", fErr, r.log.Field("AttachmentID", id))
		return nil, fErr
	}

	r.log.Debug(ctx, "GetAttachmentByID Repo Cache Set", r.log.Field("AttachmentID", id))

	attachmentJSON, mErr := json.Marshal(attachment)

	if mErr != nil {
		r.log.Error(ctx, "GetAttachmentByID Repo Cache Marshal Error", mErr, r.log.Field("AttachmentID", id))
		return &attachment, nil
	}

	cErr := r.SetCache(ctx, "attachment:"+strconv.FormatUint(id, 10), attachmentJSON, 10*time.Minute)

	if cErr != nil {
		r.log.Error(ctx, "GetAttachmentByID Repo Cache Set Error", cErr, r.log.Field("AttachmentID", id))
	}

	return &attachment, nil
}

func (r AttachmentRepository) GetAttachmentsByPostID(ctx *gin.Context, postID uint64) ([]model.Attachment, error) {
	var attachments []model.Attachment

	getResult, err := r.GetCache(ctx, "attachment:post:"+strconv.FormatUint(postID, 10))

	if err == nil {
		r.log.Debug(ctx, "GetAttachmentsByPostID Repo Cache Hit", r.log.Field("PostID", postID))
		return getResult, nil
	}

	r.log.Debug(ctx, "GetAttachmentsByPostID Repo Cache Miss", r.log.Field("PostID", postID))

	fErr := r.GormDB.Where("post_id = ?", postID).Find(&attachments).Error
	if fErr != nil {
		return nil, fErr
	}

	r.log.Debug(ctx, "GetAttachmentsByPostID Repo Cache Set", r.log.Field("PostID", postID), r.log.Field("Count", len(attachments)))

	attachmentsJSON, mErr := json.Marshal(attachments)

	if mErr != nil {
		r.log.Error(ctx, "GetAttachmentsByPostID Repo Cache Marshal Error", mErr, r.log.Field("PostID", postID))
		return attachments, nil
	}

	cErr := r.SetCache(ctx, "attachment:post:"+strconv.FormatUint(postID, 10), attachmentsJSON, 10*time.Minute)

	if cErr != nil {
		r.log.Error(ctx, "GetAttachmentsByPostID Repo Cache Set Error", cErr, r.log.Field("PostID", postID))
	}

	return attachments, nil
}

func (r AttachmentRepository) GetCache(ctx context.Context, key string) ([]model.Attachment, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.Attachment

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
		var attachments []model.Attachment

		err := json.Unmarshal([]byte(getResult.Val()), &attachments)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return attachments, nil
	}

	if !r.isSlice(result) {
		var attachment model.Attachment

		jsonR, err := json.Marshal(result)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Marshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		err = json.Unmarshal(jsonR, &attachment)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		returns = append(returns, attachment)
	}

	return returns, nil
}

// SETTER
func (r *AttachmentRepository) Delete(ctx *gin.Context, attachment *model.Attachment) error {

	err := r.GormDB.Delete(attachment).Error

	if err != nil {
		r.log.Error(ctx, "Repo Delete Error", err, r.log.Field("AttachmentID", attachment.ID))
		return err
	}

	// Invalidate related caches
	r.DeleteCache(ctx, "attachment:"+strconv.FormatUint(uint64(attachment.ID), 10))
	r.DeleteCache(ctx, "attachment:post:"+strconv.FormatUint(uint64(attachment.PostID), 10))
	r.DeleteCache(ctx, "attachment:all")

	return nil
}

func (r *AttachmentRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *AttachmentRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r AttachmentRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
