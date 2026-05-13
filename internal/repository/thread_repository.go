package repository

import (
	"gin-quickstart/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ThreadRepository struct {
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewThreadRepository(db *gorm.DB, redis *redis.Client) *ThreadRepository {
	return &ThreadRepository{
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r ThreadRepository) GetAllThreads() ([]model.Thread, error) {
	var threads []model.Thread
	err := r.GormDB.
		Preload("Category").
		Preload("Posts").
		Preload("Tags").
		Preload("Author").
		Preload("PinnedByUser").
		Find(&threads).Error

	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (r ThreadRepository) GetThreadByID(id uint64) (*model.Thread, error) {
	var thread model.Thread
	err := r.GormDB.
		Preload("Category").
		Preload("Posts").
		Preload("Tags").
		Preload("Author").
		Preload("PinnedByUser").
		First(&thread, id).
		Error

	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r ThreadRepository) GetThreadBySlug(slug string) (*model.Thread, error) {
	var thread model.Thread
	err := r.GormDB.Where("slug = ?", slug).First(&thread).Error
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r ThreadRepository) GetThreadsByCategoryID(categoryID uint) ([]model.Thread, error) {
	var threads []model.Thread
	err := r.GormDB.Where("category_id = ?", categoryID).Find(&threads).Error
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (r ThreadRepository) GetThreadsByAuthorID(authorID uint) ([]model.Thread, error) {
	var threads []model.Thread
	err := r.GormDB.Where("author_id = ?", authorID).Find(&threads).Error
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (r ThreadRepository) GetThreadsByTagID(tagID uint) ([]model.Thread, error) {
	var threads []model.Thread
	err := r.GormDB.
		Joins("JOIN thread_tags ON thread_tags.thread_id = threads.id").
		Where("thread_tags.tag_id = ?", tagID).
		Find(&threads).Error

	if err != nil {
		return nil, err
	}
	return threads, nil
}

// SETTER
func (r *ThreadRepository) Create(thread *model.Thread) error {
	return r.GormDB.Create(thread).Error
}

func (r *ThreadRepository) Update(thread *model.Thread) error {
	return r.GormDB.Save(thread).Error
}

func (r *ThreadRepository) Delete(thread *model.Thread) error {
	return r.GormDB.Delete(thread).Error
}

func (r *ThreadRepository) CreatePostAttachment(
	post *model.Post,
	attachment *model.Attachment,
) error {
	if err := r.GormDB.Create(attachment).Error; err != nil {
		return err
	}

	if err := r.GormDB.Save(post).Error; err != nil {
		return err
	}

	if err := r.GormDB.Save(attachment).Error; err != nil {
		return err
	}

	return nil
}
