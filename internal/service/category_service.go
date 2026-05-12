package service

import (
	"errors"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
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
func (s CategoryService) GetAllCategories() ([]model.Category, error) {
	return s.r.GetAllCategories()
}

func (s CategoryService) GetCategoryByID(id uint64) (*model.Category, error) {
	return s.r.GetCategoryByID(id)
}

func (s CategoryService) GetCategoryBySlug(slug string) (*model.Category, error) {
	return s.r.GetCategoryBySlug(slug)
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

	return category, s.r.Update(category)
}

func (s *CategoryService) Delete(ID uint64) error {
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

	return s.r.Delete(category)
}
