package repository

import (
	"gin-quickstart/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewCategoryRepository(db *gorm.DB, redis *redis.Client) *CategoryRepository {
	return &CategoryRepository{
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r CategoryRepository) GetAllCategories() ([]model.Category, error) {
	var categories []model.Category
	err := r.GormDB.
		Preload("Categories").
		Preload("Threads").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r CategoryRepository) GetCategoryByID(id uint64) (*model.Category, error) {
	var category model.Category
	err := r.GormDB.
		Preload("Categories").
		Preload("Threads").
		First(&category, id).Error

	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r CategoryRepository) GetCategoryBySlug(slug string) (*model.Category, error) {
	var category model.Category
	err := r.GormDB.
		Preload("Categories").
		Preload("Threads").
		Where("slug = ?", slug).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// SETTER
func (r *CategoryRepository) Create(category *model.Category) error {
	return r.GormDB.Create(category).Error
}

func (r *CategoryRepository) Update(category *model.Category) error {
	return r.GormDB.Save(category).Error
}

func (r *CategoryRepository) Delete(category *model.Category) error {
	return r.GormDB.Delete(category).Error
}
