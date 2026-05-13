package repository

import (
	"gin-quickstart/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type TagRepository struct {
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewTagRepository(db *gorm.DB, redis *redis.Client) *TagRepository {
	return &TagRepository{
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r TagRepository) GetAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.GormDB.
		Preload("Threads").
		Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r TagRepository) GetTagByID(id uint64) (*model.Tag, error) {
	var tag model.Tag
	err := r.GormDB.
		Preload("Threads").
		First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r TagRepository) GetTagBySlug(slug string) (*model.Tag, error) {
	var tag model.Tag
	err := r.GormDB.
		Preload("Threads").
		Where("slug = ?", slug).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// SETTER
func (r *TagRepository) Create(tag *model.Tag) error {
	return r.GormDB.Create(tag).Error
}

func (r *TagRepository) Update(tag *model.Tag) error {
	return r.GormDB.Save(tag).Error
}

func (r *TagRepository) Delete(id uint64) error {
	return r.GormDB.Delete(&model.Tag{}, id).Error
}
