package repository

import (
	"gin-quickstart/internal/model"

	"gorm.io/gorm"
)

type BadgeRepository struct {
	GormDB *gorm.DB
}

func NewBadgeRepository(db *gorm.DB) *BadgeRepository {
	return &BadgeRepository{
		GormDB: db,
	}
}

// GETTER
func (r BadgeRepository) GetAllBadges() ([]*model.Badge, error) {
	var badges []*model.Badge
	err := r.GormDB.Find(&badges).Error
	if err != nil {
		return nil, err
	}
	return badges, nil
}

func (r BadgeRepository) GetBadgeByID(id uint64) (*model.Badge, error) {
	var badge model.Badge
	err := r.GormDB.First(&badge, id).Error
	if err != nil {
		return nil, err
	}
	return &badge, nil
}

// SETTER
func (r *BadgeRepository) Create(badge *model.Badge) error {
	return r.GormDB.Create(badge).Error
}

func (r *BadgeRepository) Update(badge *model.Badge) error {
	return r.GormDB.Save(badge).Error
}

func (r *BadgeRepository) Delete(badge *model.Badge) error {
	return r.GormDB.Delete(badge).Error
}
