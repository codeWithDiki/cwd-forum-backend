package service

import (
	"encoding/json"
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
func (s BadgeService) GetAllBadges(ctx *gin.Context) ([]*model.Badge, error) {
	getStatus := s.r.RedisClient.Get(ctx, "badges")

	if getStatus.Err() == nil {
		var badges []*model.Badge
		err := json.Unmarshal([]byte(getStatus.Val()), &badges)

		if err != nil {
			return nil, err
		}

		return badges, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	badges, err := s.r.GetAllBadges()

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(badges)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "badges", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return badges, nil
}

func (s BadgeService) GetBadgeByID(id uint64, ctx *gin.Context) (*model.Badge, error) {
	getStatus := s.r.RedisClient.Get(ctx, "badge:id:"+strconv.FormatUint(id, 10))

	if getStatus.Err() == nil {
		var badge model.Badge
		err := json.Unmarshal([]byte(getStatus.Val()), &badge)

		if err != nil {
			return nil, err
		}

		return &badge, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	badge, err := s.r.GetBadgeByID(id)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(badge)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "badge:id:"+strconv.FormatUint(id, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return badge, nil
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
	ctx *gin.Context,
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

	delCmdStatus := s.r.RedisClient.Del(ctx, "badges")

	if delCmdStatus.Err() != nil {
		return nil, delCmdStatus.Err()
	}

	return badge, nil
}

func (s *BadgeService) Update(
	ID uint64,
	Name string,
	Description string,
	IconUrl string,
	CriteriaType string,
	CriteriaValue int,
	FontColor string,
	BackgroundColor string,
	ctx *gin.Context,
) (*model.Badge, error) {
	badge, err := s.r.GetBadgeByID(ID)

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

	if IconUrl != "" {
		badge.IconUrl = IconUrl
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

	err = s.r.Update(badge)

	if err != nil {
		return nil, err
	}

	delCmdStatus := s.r.RedisClient.Del(ctx, "badges:"+strconv.FormatUint(ID, 10))

	if delCmdStatus.Err() != nil {
		return nil, delCmdStatus.Err()
	}

	return badge, nil
}

func (s *BadgeService) Delete(badge *model.Badge, ctx *gin.Context) error {
	delCmdStatus := s.r.RedisClient.Del(ctx, "badges:"+strconv.FormatUint(uint64(badge.ID), 10))

	if delCmdStatus.Err() != nil {
		return delCmdStatus.Err()
	}

	delListCmdStatus := s.r.RedisClient.Del(ctx, "badges")

	if delListCmdStatus.Err() != nil {
		return delListCmdStatus.Err()
	}

	return s.r.Delete(badge)
}
