package service

import (
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
)

type BadgeService struct {
	r *repository.BadgeRepository
}

func NewBadgeService(r *repository.BadgeRepository) *BadgeService {
	return &BadgeService{
		r: r,
	}
}

// GETTER
func (s BadgeService) GetAllBadges() ([]*model.Badge, error) {
	return s.r.GetAllBadges()
}

func (s BadgeService) GetBadgeByID(id uint64) (*model.Badge, error) {
	return s.r.GetBadgeByID(id)
}

// SETTER
func (s *BadgeService) Create(
	Name string,
	Description string,
	IconUrl string,
	CriteriaType string,
	CriteriaValue int,
	FontColor string,
	BackgroundColor string,
) (*model.Badge, error) {
	criteriaType, err := enum.BadgeCriteriaTypeFromString(CriteriaType)

	if err == false {
		return nil, errors.New("Criteria is not registered")
	}

	badge := &model.Badge{
		Name:          Name,
		Description:   Description,
		IconUrl:       IconUrl,
		CriteriaType:  criteriaType.String(),
		CriteriaValue: CriteriaValue,
	}

	cErr := s.r.Create(badge)

	if cErr != nil {
		return nil, cErr
	}

	return badge, nil
}

func (s *BadgeService) Update(
	ID uint64,
	Name *string,
	Description *string,
	IconUrl *string,
	CriteriaType *string,
	CriteriaValue *int,
	FontColor *string,
	BackgroundColor *string,
) (*model.Badge, error) {
	badge, err := s.r.GetBadgeByID(ID)

	if err != nil {
		return nil, err
	}

	if badge == nil {
		return nil, errors.New("Badge not found")
	}

	if Name != nil {
		badge.Name = *Name
	}

	if Description != nil {
		badge.Description = *Description
	}

	if IconUrl != nil {
		badge.IconUrl = *IconUrl
	}

	if CriteriaType != nil {
		criteriaType, err := enum.BadgeCriteriaTypeFromString(*CriteriaType)

		if err == false {
			return nil, errors.New("Criteria is not registered")
		}

		badge.CriteriaType = criteriaType.String()
	}

	if CriteriaValue != nil {
		badge.CriteriaValue = *CriteriaValue
	}

	if FontColor != nil {
		badge.FontColor = *FontColor
	}

	if BackgroundColor != nil {
		badge.BackgroundColor = *BackgroundColor
	}

	err = s.r.Update(badge)

	if err != nil {
		return nil, err
	}

	return badge, nil
}

func (s *BadgeService) Delete(badge *model.Badge) error {
	return s.r.Delete(badge)
}
