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

type TagService struct {
	r *repository.TagRepository
}

func NewTagService(r *repository.TagRepository) *TagService {
	return &TagService{
		r: r,
	}
}

// GETTER
func (s TagService) GetAllTags(ctx *gin.Context) ([]model.Tag, error) {
	getStatus := s.r.RedisClient.Get(ctx, "tags")

	if getStatus.Err() == nil {
		var tags []model.Tag
		err := json.Unmarshal([]byte(getStatus.Val()), &tags)

		if err != nil {
			return nil, err
		}

		return tags, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	tags, err := s.r.GetAllTags()

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(tags)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "tags", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return tags, nil
}

func (s TagService) GetTagByID(id uint64, ctx *gin.Context) (*model.Tag, error) {
	getStatus := s.r.RedisClient.Get(ctx, "tag:"+strconv.FormatUint(id, 10))

	if getStatus.Err() == nil {
		var tag model.Tag
		err := json.Unmarshal([]byte(getStatus.Val()), &tag)

		if err != nil {
			return nil, err
		}

		return &tag, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	tag, err := s.r.GetTagByID(id)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(tag)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "tag:"+strconv.FormatUint(id, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return tag, nil
}

func (s TagService) GetTagBySlug(slug string, ctx *gin.Context) (*model.Tag, error) {
	getStatus := s.r.RedisClient.Get(ctx, "tag:slug:"+slug)

	if getStatus.Err() == nil {
		var tag model.Tag
		err := json.Unmarshal([]byte(getStatus.Val()), &tag)

		if err != nil {
			return nil, err
		}

		return &tag, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	tag, err := s.r.GetTagBySlug(slug)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(tag)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "tag:slug:"+slug, json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return tag, nil
}

// SETTER
func (s *TagService) Create(
	Name string,
	Slug string,
	Color string,
	ctx *gin.Context,
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

	delStatus := s.r.RedisClient.Del(ctx, "tags")

	if delStatus.Err() != nil {
		return nil, delStatus.Err()
	}

	return tag, nil
}

func (s *TagService) Update(
	ID uint64,
	Name *string,
	Slug *string,
	Color *string,
	ctx *gin.Context,
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

	delStatus := s.r.RedisClient.Del(ctx, "tags", "tag:"+strconv.FormatUint(ID, 10), "tag:slug:"+tag.Slug)

	if delStatus.Err() != nil {
		return nil, delStatus.Err()
	}

	return tag, nil
}

func (s *TagService) Delete(id uint64, ctx *gin.Context) error {
	tag, err := s.r.GetTagByID(id)
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

	err = s.r.Delete(id)
	if err != nil {
		return err
	}

	delStatus := s.r.RedisClient.Del(ctx, "tags", "tag:"+strconv.FormatUint(id, 10), "tag:slug:"+tag.Slug)

	if delStatus.Err() != nil {
		return delStatus.Err()
	}

	return nil
}
