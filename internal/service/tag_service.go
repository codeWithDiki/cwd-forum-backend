package service

import (
	"errors"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
)

type TagService struct {
	r *repository.TagRepository
}

func NewTagService(r *repository.TagRepository) *TagService {
	return &TagService{
		r: r,
	}
}

// GETTER
func (s TagService) GetAllTags() ([]model.Tag, error) {
	return s.r.GetAllTags()
}

func (s TagService) GetTagByID(id uint64) (*model.Tag, error) {
	return s.r.GetTagByID(id)
}

func (s TagService) GetTagBySlug(slug string) (*model.Tag, error) {
	return s.r.GetTagBySlug(slug)
}

// SETTER
func (s *TagService) Create(
	Name string,
	Slug string,
	Color string,
) (*model.Tag, error) {
	tag := &model.Tag{
		Name:  Name,
		Slug:  Slug,
		Color: Color,
	}

	slugExists, _ := s.r.GetTagBySlug(Slug)
	if slugExists != nil {
		return nil, errors.New("tag with the same slug already exists")
	}

	err := s.r.Create(tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *TagService) Update(
	ID uint64,
	Name *string,
	Slug *string,
	Color *string,
) (*model.Tag, error) {
	tag, err := s.r.GetTagByID(ID)
	if err != nil {
		return nil, err
	}

	if tag == nil {
		return nil, errors.New("tag not found")
	}

	if Name != nil {
		tag.Name = *Name
	}
	if Slug != nil {
		slugExists, _ := s.r.GetTagBySlug(*Slug)
		if slugExists != nil && slugExists.ID != uint(ID) {
			return nil, errors.New("tag with the same slug already exists")
		}
		tag.Slug = *Slug
	}
	if Color != nil {
		tag.Color = *Color
	}

	err = s.r.Update(tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *TagService) Delete(id uint64) error {
	tag, err := s.r.GetTagByID(id)
	if err != nil {
		return err
	}

	if tag == nil {
		return errors.New("tag not found")
	}

	s.r.GormDB.Model(tag).Association("Threads").Clear()

	return s.r.Delete(id)
}
