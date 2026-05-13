package repository

import (
	"context"
	"gin-quickstart/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var ctx = context.Background()

type AttachmentRepository struct {
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewAttachmentRepository(db *gorm.DB, redis *redis.Client) *AttachmentRepository {
	return &AttachmentRepository{
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER

func (r AttachmentRepository) GetAllAttachments() ([]*model.Attachment, error) {
	var attachments []*model.Attachment
	err := r.GormDB.Find(&attachments).Error
	if err != nil {
		return nil, err
	}

	return attachments, nil
}

func (r AttachmentRepository) GetAttachmentByID(id uint64) (*model.Attachment, error) {
	var attachment model.Attachment
	err := r.GormDB.First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r AttachmentRepository) GetAttachmentsByPostID(postID uint64) ([]*model.Attachment, error) {
	var attachments []*model.Attachment
	err := r.GormDB.Where("post_id = ?", postID).Find(&attachments).Error
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

// SETTER
func (r *AttachmentRepository) Delete(attachment *model.Attachment) error {
	return r.GormDB.Delete(attachment).Error
}
