package service

import (
	"errors"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"

	"github.com/gin-gonic/gin"
)

type TagService struct {
	log *logger.Logger
	r   *repository.TagRepository
}

func NewTagService(log *logger.Logger, r *repository.TagRepository) *TagService {
	return &TagService{
		log: log,
		r:   r,
	}
}

// GETTER
func (s TagService) GetAllTags(ctx *gin.Context) ([]model.Tag, error) {
	tags, err := s.r.GetAllTags(ctx)
	s.log.Debug(ctx, "GetAllTags Service", s.log.Field("Count", len(tags)))

	if err != nil {
		s.log.Error(ctx, "GetAllTags Service Error", err)
		return nil, err
	}

	return tags, nil
}

func (s TagService) GetTagByID(ctx *gin.Context, id uint64) (*model.Tag, error) {

	tag, err := s.r.GetTagByID(ctx, id)
	s.log.Debug(ctx, "GetTagByID Service", s.log.Field("ID", id))

	if err != nil {
		s.log.Error(ctx, "GetTagByID Service Error", err, s.log.Field("ID", id))
		return nil, err
	}

	return tag, nil
}

func (s TagService) GetTagBySlug(ctx *gin.Context, slug string) (*model.Tag, error) {

	tag, err := s.r.GetTagBySlug(ctx, slug)
	s.log.Debug(ctx, "GetTagBySlug Service", s.log.Field("Slug", slug))

	if err != nil {
		s.log.Error(ctx, "GetTagBySlug Service Error", err, s.log.Field("Slug", slug))
		return nil, err
	}

	return tag, nil
}

// SETTER
func (s *TagService) Create(
	ctx *gin.Context,
	Name string,
	Slug string,
	Color string,
) (*model.Tag, error) {
	tag := &model.Tag{
		Name:  Name,
		Slug:  Slug,
		Color: Color,
	}

	slugExists, _ := s.r.GetTagBySlug(ctx, Slug)
	if slugExists != nil {
		return nil, errors.New("tag with the same slug already exists")
	}

	err := s.r.Create(ctx, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *TagService) Update(
	ctx *gin.Context,
	ID uint64,
	Name string,
	Slug string,
	Color string,
) (*model.Tag, error) {
	tag, err := s.GetTagByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	if tag == nil {
		return nil, errors.New("tag not found")
	}

	if Name != "" {
		tag.Name = Name
	}
	if Slug != "" {
		slugExists, _ := s.r.GetTagBySlug(ctx, Slug)
		if slugExists != nil && slugExists.ID != uint(ID) {
			return nil, errors.New("tag with the same slug already exists")
		}
		tag.Slug = Slug
	}
	if Color != "" {
		tag.Color = Color
	}

	err = s.r.Update(ctx, tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *TagService) Delete(ctx *gin.Context, id uint64) error {
	tag, err := s.r.GetTagByID(ctx, id)
	if err != nil {
		return err
	}

	if tag == nil {
		return errors.New("tag not found")
	}

	pruneErr := s.r.GormDB.Model(tag).Association("Threads").Clear()

	if pruneErr != nil {
		return pruneErr
	}

	err = s.r.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
