package repository

import (
	"gin-quickstart/internal/model"
	"gin-quickstart/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ModerationLogRepository struct {
	log    *logger.Logger
	GormDB *gorm.DB
}

func NewModerationLogRepository(log *logger.Logger, db *gorm.DB) *ModerationLogRepository {
	return &ModerationLogRepository{
		log:    log,
		GormDB: db,
	}
}

func (r *ModerationLogRepository) Create(ctx *gin.Context, log *model.ModerationLog) error {
	return r.GormDB.Create(log).Error
}

func (r *ModerationLogRepository) GetByModeratorID(ctx *gin.Context, moderatorID uint64) ([]model.ModerationLog, error) {
	var logs []model.ModerationLog
	err := r.GormDB.Where("moderator_id = ?", moderatorID).Find(&logs).Error
	return logs, err
}

func (r *ModerationLogRepository) GetByTargetUserID(ctx *gin.Context, targetUserID uint64) ([]model.ModerationLog, error) {
	var logs []model.ModerationLog
	err := r.GormDB.Where("target_user_id = ?", targetUserID).Find(&logs).Error
	return logs, err
}

func (r *ModerationLogRepository) GetAll(ctx *gin.Context) ([]model.ModerationLog, error) {
	var logs []model.ModerationLog
	err := r.GormDB.Preload("Moderator").Preload("TargetUser").Preload("TargetPost").Find(&logs).Error
	return logs, err
}

func (r *ModerationLogRepository) DeleteExpired(ctx *gin.Context) error {
	return r.GormDB.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&model.ModerationLog{}).Error
}
