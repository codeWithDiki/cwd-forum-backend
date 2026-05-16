package repository

import (
	"context"
	"encoding/json"
	"gin-quickstart/internal/model"
	"gin-quickstart/pkg/logger"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PostRepository struct {
	log         *logger.Logger
	GormDB      *gorm.DB
	RedisClient *redis.Client
}

func NewPostRepository(log *logger.Logger, db *gorm.DB, redis *redis.Client) *PostRepository {
	return &PostRepository{
		log:         log,
		GormDB:      db,
		RedisClient: redis,
	}
}

// GETTER
func (r PostRepository) GetAllPosts(ctx *gin.Context) ([]model.Post, error) {
	var posts []model.Post

	getResult, err := r.GetCache(ctx, "post:all")

	if err == nil {
		r.log.Debug(ctx, "GetAllPosts Repo Cache Hit")
		posts = getResult
		return posts, nil
	}

	r.log.Debug(ctx, "GetAllPosts Repo Cache Miss")

	err = r.GormDB.
		Preload("Thread").
		Preload("Author").
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetAllPosts Repo Cache Set", r.log.Field("Count", len(posts)))

	postsJSON, mErr := json.Marshal(posts)

	if mErr != nil {
		r.log.Error(ctx, "GetAllPosts Repo Cache Marshal Error", mErr)
		return posts, nil
	}

	err = r.SetCache(ctx, "post:all", postsJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetAllPosts Repo Cache Set Error", err)
		return posts, nil
	}

	return posts, nil
}

func (r PostRepository) GetPostByID(ctx *gin.Context, id uint64) (*model.Post, error) {
	var post model.Post

	getResult, err := r.GetCache(ctx, "post:id:"+strconv.FormatUint(id, 10))

	if err == nil {
		r.log.Debug(ctx, "GetPostByID Repo Cache Hit", r.log.Field("ID", id))
		post = getResult[0]
		return &post, nil
	}

	r.log.Debug(ctx, "GetPostByID Repo Cache Miss", r.log.Field("ID", id))

	err = r.GormDB.
		Preload("Thread").
		Preload("Author").
		Preload("Posts").
		First(&post, id).Error

	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetPostByID Repo Cache Set", r.log.Field("ID", id))

	postJSON, mErr := json.Marshal(post)

	if mErr != nil {
		r.log.Error(ctx, "GetPostByID Repo Cache Marshal Error", mErr, r.log.Field("ID", id))
		return &post, nil
	}

	err = r.SetCache(ctx, "post:id:"+strconv.FormatUint(id, 10), postJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetPostByID Repo Cache Set Error", err, r.log.Field("ID", id))
		return &post, nil
	}

	return &post, nil
}

func (r PostRepository) GetPostsByThreadID(ctx *gin.Context, threadID uint64) ([]model.Post, error) {
	var posts []model.Post

	getResult, err := r.GetCache(ctx, "posts:thread:"+strconv.FormatUint(threadID, 10))

	if err == nil {
		r.log.Debug(ctx, "GetPostsByThreadID Repo Cache Hit", r.log.Field("ThreadID", threadID))
		posts = getResult
		return posts, nil
	}

	r.log.Debug(ctx, "GetPostsByThreadID Repo Cache Miss", r.log.Field("ThreadID", threadID))

	err = r.GormDB.
		Preload("Thread").
		Preload("Author").
		Where("thread_id = ?", threadID).Find(&posts).Error

	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetPostsByThreadID Repo Cache Set", r.log.Field("ThreadID", threadID), r.log.Field("Count", len(posts)))

	postsJSON, mErr := json.Marshal(posts)

	if mErr != nil {
		r.log.Error(ctx, "GetPostsByThreadID Repo Cache Marshal Error", mErr, r.log.Field("ThreadID", threadID))
		return posts, nil
	}

	err = r.SetCache(ctx, "posts:thread:"+strconv.FormatUint(threadID, 10), postsJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetPostsByThreadID Repo Cache Set Error", err, r.log.Field("ThreadID", threadID))
		return posts, nil
	}

	return posts, nil
}

func (r PostRepository) GetPostsByAuthorID(ctx *gin.Context, authorID uint64) ([]model.Post, error) {
	var posts []model.Post

	getResult, err := r.GetCache(ctx, "posts:author:"+strconv.FormatUint(authorID, 10))

	if err == nil {
		r.log.Debug(ctx, "GetPostsByAuthorID Repo Cache Hit", r.log.Field("AuthorID", authorID))
		posts = getResult
		return posts, nil
	}

	r.log.Debug(ctx, "GetPostsByAuthorID Repo Cache Miss", r.log.Field("AuthorID", authorID))

	err = r.GormDB.
		Preload("Thread").
		Preload("Author").
		Where("author_id = ?", authorID).Find(&posts).Error

	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetPostsByAuthorID Repo Cache Set", r.log.Field("AuthorID", authorID), r.log.Field("Count", len(posts)))

	postsJSON, mErr := json.Marshal(posts)

	if mErr != nil {
		r.log.Error(ctx, "GetPostsByAuthorID Repo Cache Marshal Error", mErr, r.log.Field("AuthorID", authorID))
		return posts, nil
	}

	err = r.SetCache(ctx, "posts:author:"+strconv.FormatUint(authorID, 10), postsJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetPostsByAuthorID Repo Cache Set Error", err, r.log.Field("AuthorID", authorID))
		return posts, nil
	}

	return posts, nil
}

func (r PostRepository) GetPostsByParentID(ctx *gin.Context, parentID uint64) ([]model.Post, error) {
	var posts []model.Post

	getResult, err := r.GetCache(ctx, "posts:parent:"+strconv.FormatUint(parentID, 10))

	if err == nil {
		r.log.Debug(ctx, "GetPostsByParentID Repo Cache Hit", r.log.Field("ParentID", parentID))
		posts = getResult
		return posts, nil
	}

	r.log.Debug(ctx, "GetPostsByParentID Repo Cache Miss", r.log.Field("ParentID", parentID))

	err = r.GormDB.
		Preload("Thread").
		Preload("Author").
		Where("parent_id = ?", parentID).Find(&posts).Error

	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetPostsByParentID Repo Cache Set", r.log.Field("ParentID", parentID), r.log.Field("Count", len(posts)))

	postsJSON, mErr := json.Marshal(posts)

	if mErr != nil {
		r.log.Error(ctx, "GetPostsByParentID Repo Cache Marshal Error", mErr, r.log.Field("ParentID", parentID))
		return posts, nil
	}

	err = r.SetCache(ctx, "posts:parent:"+strconv.FormatUint(parentID, 10), postsJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetPostsByParentID Repo Cache Set Error", err, r.log.Field("ParentID", parentID))
		return posts, nil
	}

	return posts, nil
}

func (r PostRepository) GetPostVotes(ctx *gin.Context, postID uint64) ([]model.Vote, error) {
	var votes []model.Vote

	getResult, err := r.GetVotesCache(ctx, "post:votes:"+strconv.FormatUint(postID, 10))

	if err == nil {
		r.log.Debug(ctx, "GetPostVotes Repo Cache Hit", r.log.Field("PostID", postID))
		votes = getResult
		return votes, nil
	}

	r.log.Debug(ctx, "GetPostVotes Repo Cache Miss", r.log.Field("PostID", postID))

	err = r.GormDB.Where("post_id = ?", postID).Find(&votes).Error

	if err != nil {
		return nil, err
	}

	r.log.Debug(ctx, "GetPostVotes Repo Cache Set", r.log.Field("PostID", postID), r.log.Field("Count", len(votes)))

	votesJSON, mErr := json.Marshal(votes)

	if mErr != nil {
		r.log.Error(ctx, "GetPostVotes Repo Cache Marshal Error", mErr, r.log.Field("PostID", postID))
		return votes, nil
	}

	err = r.SetCache(ctx, "post:votes:"+strconv.FormatUint(postID, 10), votesJSON, time.Hour)

	if err != nil {
		r.log.Error(ctx, "GetPostVotes Repo Cache Set Error", err, r.log.Field("PostID", postID))
		return votes, nil
	}

	return votes, nil
}

func (r PostRepository) GetCache(ctx context.Context, key string) ([]model.Post, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.Post

	if getResult.Err() != nil {
		r.log.Error(ctx, "Repo GetCache Error", getResult.Err(), r.log.Field("Key", key))
		return nil, getResult.Err()
	}

	err := json.Unmarshal([]byte(getResult.Val()), &result)

	if err != nil {
		r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
		return nil, err
	}

	if r.isSlice(result) {
		var users []model.Post

		err := json.Unmarshal([]byte(getResult.Val()), &users)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return users, nil
	}

	if !r.isSlice(result) {
		var user model.Post

		jsonR, err := json.Marshal(result)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Marshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		err = json.Unmarshal(jsonR, &user)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		returns = append(returns, user)
	}

	return returns, nil
}

func (r PostRepository) GetVotesCache(ctx context.Context, key string) ([]model.Vote, error) {
	r.log.Debug(ctx, "Repo GetCache Called", r.log.Field("Key", key))
	getResult := r.RedisClient.Get(ctx, key)
	var result interface{}
	var returns []model.Vote

	if getResult.Err() != nil {
		r.log.Error(ctx, "Repo GetCache Error", getResult.Err(), r.log.Field("Key", key))
		return nil, getResult.Err()
	}

	err := json.Unmarshal([]byte(getResult.Val()), &result)

	if err != nil {
		r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
		return nil, err
	}

	if r.isSlice(result) {
		var users []model.Vote

		err := json.Unmarshal([]byte(getResult.Val()), &users)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		return users, nil
	}

	if !r.isSlice(result) {
		var user model.Vote

		jsonR, err := json.Marshal(result)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Marshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		err = json.Unmarshal(jsonR, &user)

		if err != nil {
			r.log.Error(ctx, "Repo GetCache Unmarshal Error", err, r.log.Field("Key", key))
			return nil, err
		}

		returns = append(returns, user)
	}

	return returns, nil
}

// SETTER

func (r *PostRepository) Create(ctx *gin.Context, post *model.Post) error {
	err := r.GormDB.Create(post).Error

	if err != nil {
		return err
	}

	err = r.DeleteCache(ctx, "post:all")

	if err != nil {
		r.log.Error(ctx, "Repo Create Post DeleteCache Error", err)
	}

	return nil
}

func (r *PostRepository) Update(ctx *gin.Context, post *model.Post) error {
	post.IsEdited = true

	err := r.GormDB.Save(post).Error

	if err != nil {
		return err
	}

	err = r.DeleteCache(ctx, "post:id:"+strconv.FormatUint(uint64(post.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo Update Post DeleteCache Error", err, r.log.Field("ID", post.ID))
	}

	err = r.DeleteCache(ctx, "posts:thread:"+strconv.FormatUint(uint64(post.ThreadID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo Update Post DeleteCache Error", err, r.log.Field("ThreadID", post.ThreadID))
	}

	err = r.DeleteCache(ctx, "posts:author:"+strconv.FormatUint(uint64(post.AuthorID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo Update Post DeleteCache Error", err, r.log.Field("AuthorID", post.AuthorID))
	}
	return nil
}

func (r *PostRepository) Delete(ctx *gin.Context, post *model.Post) error {
	err := r.GormDB.Delete(post).Error

	if err != nil {
		return err
	}

	err = r.DeleteCache(ctx, "post:id:"+strconv.FormatUint(uint64(post.ID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo Delete Post DeleteCache Error", err, r.log.Field("ID", post.ID))
	}

	err = r.DeleteCache(ctx, "posts:thread:"+strconv.FormatUint(uint64(post.ThreadID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo Delete Post DeleteCache Error", err, r.log.Field("ThreadID", post.ThreadID))
	}

	err = r.DeleteCache(ctx, "posts:author:"+strconv.FormatUint(uint64(post.AuthorID), 10))

	if err != nil {
		r.log.Error(ctx, "Repo Delete Post DeleteCache Error", err, r.log.Field("AuthorID", post.AuthorID))
	}

	return nil
}

func (r *PostRepository) CreateAttachment(ctx *gin.Context, postID uint64,
	attachment *model.Attachment) (*model.Attachment, error) {

	err := r.GormDB.Model(&model.Post{ID: uint(postID)}).Association("Attachments").Append(attachment)

	if err != nil {
		return nil, err
	}

	err = r.DeleteCache(ctx, "post:id:"+strconv.FormatUint(postID, 10))

	if err != nil {
		r.log.Error(ctx, "Repo Create Attachment DeleteCache Error", err, r.log.Field("PostID", postID))
	}

	err = r.DeleteCache(ctx, "posts:thread:"+strconv.FormatUint(postID, 10))

	if err != nil {
		r.log.Error(ctx, "Repo Create Attachment DeleteCache Error", err, r.log.Field("PostID", postID))
	}

	return attachment, nil
}

func (r *PostRepository) SetCache(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	r.log.Debug(ctx, "Repo SetCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Set(ctx, key, value, expiration)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo SetCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

func (r *PostRepository) DeleteCache(ctx context.Context, key string) error {
	r.log.Debug(ctx, "Repo DeleteCache Called", r.log.Field("Key", key))
	cmdStatus := r.RedisClient.Del(ctx, key)

	if cmdStatus.Err() != nil {
		r.log.Error(ctx, "Repo DeleteCache Error", cmdStatus.Err(), r.log.Field("Key", key))
		return cmdStatus.Err()
	}

	return nil
}

// CHECKER
func (r PostRepository) isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}
