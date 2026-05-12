package repository

import (
	"gin-quickstart/internal/model"

	"gorm.io/gorm"
)

type AttachmentRepository struct {
	GormDB *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) *AttachmentRepository {
	return &AttachmentRepository{
		GormDB: db,
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
