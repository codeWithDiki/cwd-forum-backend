package service

import (
	"encoding/json"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AttachmentService struct {
	r *repository.AttachmentRepository
}

func NewAttachmentService(r *repository.AttachmentRepository) *AttachmentService {
	return &AttachmentService{
		r: r,
	}
}

// GETTER
func (s AttachmentService) GetAllAttachments(ctx *gin.Context) ([]*model.Attachment, error) {
	getStatus := s.r.RedisClient.Get(ctx, "attachments")

	if getStatus.Err() == nil {
		var attachments []*model.Attachment
		err := json.Unmarshal([]byte(getStatus.Val()), &attachments)

		if err != nil {
			return nil, err
		}

		return attachments, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	attachments, err := s.r.GetAllAttachments()

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(attachments)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "attachments", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return attachments, nil
}

func (s AttachmentService) GetAttachmentByID(id uint64, ctx *gin.Context) (*model.Attachment, error) {
	idStr := strconv.FormatUint(id, 10)

	getStatus := s.r.RedisClient.Get(ctx, "attachment:"+idStr)

	if getStatus.Err() == nil {
		var attachment model.Attachment
		err := json.Unmarshal([]byte(getStatus.Val()), &attachment)

		if err != nil {
			return nil, err
		}

		return &attachment, nil
	}

	attachmentById, err := s.r.GetAttachmentByID(id)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(attachmentById)

	if err != nil {
		return nil, err
	}

	setStatus := s.r.RedisClient.Set(ctx, "attachment:"+idStr, json, time.Hour)

	if setStatus.Err() != nil {
		return nil, setStatus.Err()
	}

	return s.r.GetAttachmentByID(id)
}

func (s AttachmentService) GetAttachmentsByPostID(postID uint64, ctx *gin.Context) ([]*model.Attachment, error) {
	getStatus := s.r.RedisClient.Get(ctx, "attachments:post:"+strconv.FormatUint(postID, 10))

	if getStatus.Err() == nil {
		var attachments []*model.Attachment
		err := json.Unmarshal([]byte(getStatus.Val()), &attachments)

		if err != nil {
			return nil, err
		}

		return attachments, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	attachments, err := s.r.GetAttachmentsByPostID(postID)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(attachments)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "attachments:post:"+strconv.FormatUint(postID, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return s.r.GetAttachmentsByPostID(postID)
}

// SETTER
func (s *AttachmentService) Delete(attachment *model.Attachment, ctx *gin.Context) error {
	s.r.RedisClient.Del(ctx, "attachments")
	s.r.RedisClient.Del(ctx, "attachment:"+strconv.FormatUint(uint64(attachment.ID), 10))
	s.r.RedisClient.Del(ctx, "attachments:post:"+strconv.FormatUint(uint64(attachment.PostID), 10))
	return s.r.Delete(attachment)
}
