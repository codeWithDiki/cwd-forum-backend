package service

import (
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

type ModerationLogService struct {
	log  *logger.Logger
	Repo *repository.ModerationLogRepository
}

func NewModerationLogService(log *logger.Logger, repo *repository.ModerationLogRepository) *ModerationLogService {
	return &ModerationLogService{
		log:  log,
		Repo: repo,
	}
}

func (s *ModerationLogService) LogAction(ctx *gin.Context, moderatorID uint, action string, reason string, targetUserID *uint, targetPostID *uint, expiresAt *time.Time) error {
	log := &model.ModerationLog{
		ModeratorId:  moderatorID,
		TargetUserId: targetUserID,
		TargetPostId: targetPostID,
		Action:       action,
		Reason:       reason,
		ExpiresAt:    expiresAt,
	}

	return s.Repo.Create(ctx, log)
}

func (s *ModerationLogService) GetByModeratorID(ctx *gin.Context, moderatorID uint64) ([]model.ModerationLog, error) {
	return s.Repo.GetByModeratorID(ctx, moderatorID)
}

func (s *ModerationLogService) GetByTargetUserID(ctx *gin.Context, targetUserID uint64) ([]model.ModerationLog, error) {
	return s.Repo.GetByTargetUserID(ctx, targetUserID)
}

func (s *ModerationLogService) GetAll(ctx *gin.Context) ([]model.ModerationLog, error) {
	return s.Repo.GetAll(ctx)
}
