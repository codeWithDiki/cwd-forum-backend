package repository

import (
	"gin-quickstart/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PostRepository struct {
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewPostRepository(db *gorm.DB, redis *redis.Client) *PostRepository {
	return &PostRepository{
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r PostRepository) GetAllPosts() ([]model.Post, error) {
	var posts []model.Post
	err := r.GormDB.
		Preload("Thread").
		Preload("Author").
		Find(&posts).Error

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (r PostRepository) GetPostByID(id uint64) (*model.Post, error) {
	var post model.Post
	err := r.GormDB.
		Preload("Thread").
		Preload("Author").
		Preload("Posts").
		First(&post, id).Error

	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r PostRepository) GetPostsByThreadID(threadID uint64) ([]model.Post, error) {
	var posts []model.Post
	err := r.GormDB.
		Preload("Thread").
		Preload("Author").
		Where("thread_id = ?", threadID).Find(&posts).Error

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (r PostRepository) GetPostsByAuthorID(authorID uint64) ([]model.Post, error) {
	var posts []model.Post
	err := r.GormDB.
		Preload("Thread").
		Preload("Author").
		Where("author_id = ?", authorID).Find(&posts).Error

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (r PostRepository) GetPostsByParentID(parentID uint64) ([]model.Post, error) {
	var posts []model.Post
	err := r.GormDB.
		Preload("Thread").
		Preload("Author").
		Where("parent_id = ?", parentID).Find(&posts).Error

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (r PostRepository) GetPostVotes(postID uint64) ([]model.Vote, error) {
	var votes []model.Vote
	err := r.GormDB.Where("post_id = ?", postID).Find(&votes).Error

	if err != nil {
		return nil, err
	}
	return votes, nil
}

// SETTER

func (r *PostRepository) Create(post *model.Post) error {
	return r.GormDB.Create(post).Error
}

func (r *PostRepository) Update(post *model.Post) error {
	post.IsEdited = true
	return r.GormDB.Save(post).Error
}

func (r *PostRepository) Delete(post *model.Post) error {
	return r.GormDB.Delete(post).Error
}

func (r *PostRepository) CreateAttachment(postID uint64,
	attachment *model.Attachment) (*model.Attachment, error) {

	r.GormDB.Model(&model.Post{ID: uint(postID)}).Association("Attachments").Append(attachment)

	return attachment, nil
}
