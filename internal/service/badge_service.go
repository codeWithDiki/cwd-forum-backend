package service

import (
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type BadgeService struct {
	log     *logger.Logger
	r       *repository.BadgeRepository
	storage *repository.StorageRepository
}

func NewBadgeService(log *logger.Logger, r *repository.BadgeRepository, storage *repository.StorageRepository) *BadgeService {
	return &BadgeService{
		log:     log,
		r:       r,
		storage: storage,
	}
}

// GETTER
func (s BadgeService) GetAllBadges(ctx *gin.Context) ([]model.Badge, error) {
	badges, err := s.r.GetAllBadges(ctx)
	s.log.Debug(ctx, "Service GetAllBadges Called", s.log.Field("Count", len(badges)))

	if err != nil {
		s.log.Error(ctx, "Service GetAllBadges Error", err)
		return nil, err
	}

	return badges, nil
}

func (s BadgeService) GetBadgeByID(ctx *gin.Context, id uint64) (*model.Badge, error) {
	badge, err := s.r.GetBadgeByID(ctx, id)
	s.log.Debug(ctx, "Service GetBadgeByID Called", s.log.Field("BadgeID", id))

	if err != nil {
		s.log.Error(ctx, "Service GetBadgeByID Error", err, s.log.Field("BadgeID", id))
		return nil, err
	}

	return badge, nil
}

// SETTER
func (s *BadgeService) Create(
	ctx *gin.Context,
	Name string,
	Description string,
	CriteriaType string,
	CriteriaValue int,
	FontColor string,
	BackgroundColor string,
	File *multipart.FileHeader,
) (*model.Badge, error) {

	criteriaType, err := enum.BadgeCriteriaTypeFromString(CriteriaType)

	if err == false {
		return nil, errors.New("Criteria is not registered")
	}

	if File == nil {
		return nil, errors.New("Icon file is required")
	}

	iconUrl, uErr := s.storage.UploadFile(ctx, File, "uploads/badges")

	if uErr != nil {
		s.log.Error(ctx, "Failed to upload badge icon", uErr)
		return nil, errors.New("Failed to upload badge icon: " + uErr.Error())
	}

	badge := &model.Badge{
		Name:          Name,
		Description:   Description,
		CriteriaType:  criteriaType.String(),
		IconUrl:       iconUrl,
		CriteriaValue: CriteriaValue,
	}

	cErr := s.r.Create(ctx, badge)

	if cErr != nil {
		return nil, cErr
	}

	return badge, nil
}

func (s *BadgeService) Update(
	ctx *gin.Context,
	ID uint64,
	Name string,
	Description string,
	CriteriaType string,
	CriteriaValue int,
	FontColor string,
	BackgroundColor string,
	File *multipart.FileHeader,
) (*model.Badge, error) {
	badge, err := s.r.GetBadgeByID(ctx, ID)

	if err != nil {
		return nil, err
	}

	if badge == nil {
		return nil, errors.New("Badge not found")
	}

	if Name != "" {
		badge.Name = Name
	}

	if Description != "" {
		badge.Description = Description
	}

	if CriteriaType != "" {
		criteriaType, err := enum.BadgeCriteriaTypeFromString(CriteriaType)

		if err == false {
			return nil, errors.New("Criteria is not registered")
		}

		badge.CriteriaType = criteriaType.String()
	}

	if CriteriaValue != 0 {
		badge.CriteriaValue = CriteriaValue
	}

	if FontColor != "" {
		badge.FontColor = FontColor
	}

	if BackgroundColor != "" {
		badge.BackgroundColor = BackgroundColor
	}

	if File != nil {
		iconUrl, uErr := s.storage.UploadFile(ctx, File, "uploads/badges")

		if uErr != nil {
			s.log.Error(ctx, "Failed to upload badge icon", uErr)
			return nil, errors.New("Failed to upload badge icon: " + uErr.Error())
		}

		deleteStatus := s.storage.DeleteFile(ctx, badge.IconUrl)

		if deleteStatus != nil {
			s.log.Error(ctx, "Failed to delete old badge icon", deleteStatus)
		}

		badge.IconUrl = iconUrl

	}

	err = s.r.Update(ctx, badge)

	if err != nil {
		return nil, err
	}

	return badge, nil
}

func (s *BadgeService) Delete(ctx *gin.Context, badge *model.Badge) error {
	return s.r.Delete(ctx, badge)
}
