package service

import (
	"encoding/json"
	"errors"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type CategoryService struct {
	r *repository.CategoryRepository
}

func NewCategoryService(r *repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		r: r,
	}
}

// GETTER
func (s CategoryService) GetAllCategories(ctx *gin.Context) ([]model.Category, error) {
	getStatus := s.r.RedisClient.Get(ctx, "categories")

	if getStatus.Err() == nil {
		var categories []model.Category
		err := json.Unmarshal([]byte(getStatus.Val()), &categories)

		if err != nil {
			return nil, err
		}

		return categories, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	categories, err := s.r.GetAllCategories()

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(categories)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "categories", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return categories, nil
}

func (s CategoryService) GetCategoryByID(id uint64, ctx *gin.Context) (*model.Category, error) {
	getStatus := s.r.RedisClient.Get(ctx, "category:id:"+strconv.FormatUint(id, 10))

	if getStatus.Err() == nil {
		var category model.Category
		err := json.Unmarshal([]byte(getStatus.Val()), &category)

		if err != nil {
			return nil, err
		}

		return &category, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	category, err := s.r.GetCategoryByID(id)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(category)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "category:id:"+strconv.FormatUint(id, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return category, nil
}

func (s CategoryService) GetCategoryBySlug(slug string, ctx *gin.Context) (*model.Category, error) {
	getStatus := s.r.RedisClient.Get(ctx, "category:slug:"+slug)

	if getStatus.Err() == nil {
		var category model.Category
		err := json.Unmarshal([]byte(getStatus.Val()), &category)

		if err != nil {
			return nil, err
		}

		return &category, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	category, err := s.r.GetCategoryBySlug(slug)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(category)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "category:slug:"+slug, json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return category, nil
}

// SETTER
func (s *CategoryService) Create(
	ParentID *uint,
	Name string,
	Slug string,
	Description string,
	IconUrl string,
	SortOrder int,
	IsPrivate bool,
	ctx *gin.Context,
) (*model.Category, error) {
	category := &model.Category{
		ParentID:    ParentID,
		Name:        Name,
		Slug:        Slug,
		Description: Description,
		IconUrl:     IconUrl,
		SortOrder:   SortOrder,
		IsPrivate:   IsPrivate,
	}

	if ParentID != nil {
		parentCategory, err := s.r.GetCategoryByID(uint64(*ParentID))

		if err != nil {
			return nil, errors.New("Parent category not found")
		}

		if parentCategory == nil {
			return nil, errors.New("Parent category not found")
		}
	}

	slugExists, _ := s.r.GetCategoryBySlug(Slug)

	if slugExists != nil {
		return nil, errors.New("Slug already exists")
	}

	err := s.r.Create(category)

	if err != nil {
		return nil, err
	}

	delStatus := s.r.RedisClient.Del(ctx, "categories")

	if delStatus.Err() != nil {
		return nil, delStatus.Err()
	}

	return category, nil
}

func (s *CategoryService) Update(
	ID uint64,
	ParentID *uint,
	Name *string,
	Slug *string,
	Description *string,
	IconUrl *string,
	SortOrder *int,
	IsPrivate *bool,
	ctx *gin.Context,
) (*model.Category, error) {
	category, err := s.r.GetCategoryByID(ID)

	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, errors.New("Category not found")
	}

	if ParentID != nil {
		parentCategory, err := s.r.GetCategoryByID(uint64(*ParentID))

		if err != nil {
			return nil, errors.New("Parent category not found")
		}

		if parentCategory == nil {
			return nil, errors.New("Parent category not found")
		}

		if category.ID == parentCategory.ID {
			return nil, errors.New("Category cannot be its own parent")
		}
	}

	if Name != nil {
		category.Name = *Name
	}

	if Slug != nil {
		slugExists, _ := s.r.GetCategoryBySlug(*Slug)

		if slugExists != nil && slugExists.ID != category.ID {
			return nil, errors.New("Slug already exists")
		}

		category.Slug = *Slug
	}

	if Description != nil {
		category.Description = *Description
	}

	if IconUrl != nil {
		category.IconUrl = *IconUrl
	}

	if SortOrder != nil {
		category.SortOrder = *SortOrder
	}

	if IsPrivate != nil {
		category.IsPrivate = *IsPrivate
	}

	err = s.r.Update(category)

	if err != nil {
		return nil, err
	}

	delIdStatus := s.r.RedisClient.Del(ctx, "category:id:"+strconv.FormatUint(ID, 10))

	if delIdStatus.Err() != nil {
		return nil, delIdStatus.Err()
	}

	delSlugStatus := s.r.RedisClient.Del(ctx, "category:slug:"+category.Slug)

	if delSlugStatus.Err() != nil {
		return nil, delSlugStatus.Err()
	}

	return category, nil
}

func (s *CategoryService) Delete(ID uint64, ctx *gin.Context) error {
	category, err := s.r.GetCategoryByID(ID)

	if err != nil {
		return err
	}

	if category == nil {
		return errors.New("Category not found")
	}

	threads := category.Threads

	if len(threads) > 0 {
		return errors.New("Cannot delete category with existing threads")
	}

	subcategories := category.Categories

	if len(subcategories) > 0 {
		return errors.New("Cannot delete category with existing subcategories")
	}

	delErr := s.r.Delete(category)

	if delErr != nil {
		return delErr
	}

	delIdStatus := s.r.RedisClient.Del(ctx, "category:id:"+strconv.FormatUint(ID, 10))

	if delIdStatus.Err() != nil {
		return delIdStatus.Err()
	}

	delSlugStatus := s.r.RedisClient.Del(ctx, "category:slug:"+category.Slug)

	if delSlugStatus.Err() != nil {
		return delSlugStatus.Err()
	}

	return nil
}
