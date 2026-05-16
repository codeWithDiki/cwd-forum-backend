package service

import (
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"

	"github.com/gin-gonic/gin"
)

type AttachmentService struct {
	log *logger.Logger
	r   *repository.AttachmentRepository
}

func NewAttachmentService(log *logger.Logger, r *repository.AttachmentRepository) *AttachmentService {
	return &AttachmentService{
		log: log,
		r:   r,
	}
}

// GETTER
func (s AttachmentService) GetAllAttachments(ctx *gin.Context) ([]model.Attachment, error) {
	attachments, err := s.r.GetAllAttachments(ctx)

	s.log.Debug(ctx, "Service GetAllAttachments Called", s.log.Field("Count", len(attachments)))

	if err != nil {
		s.log.Error(ctx, "Service GetAllAttachments Error", err)
		return nil, err
	}

	return attachments, nil
}

func (s AttachmentService) GetAttachmentByID(ctx *gin.Context, id uint64) (*model.Attachment, error) {

	attachmentById, err := s.r.GetAttachmentByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return attachmentById, nil
}

func (s AttachmentService) GetAttachmentsByPostID(ctx *gin.Context, postID uint64) ([]model.Attachment, error) {

	attachments, err := s.r.GetAttachmentsByPostID(ctx, postID)

	return attachments, err
}

// SETTER
func (s *AttachmentService) Delete(ctx *gin.Context, attachment *model.Attachment) error {
	return s.r.Delete(ctx, attachment)
}
