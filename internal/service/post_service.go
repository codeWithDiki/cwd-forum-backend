package service

import (
	"encoding/json"
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PostService struct {
	r *repository.PostRepository
}

func NewPostService(r *repository.PostRepository) *PostService {
	return &PostService{
		r: r,
	}
}

// GETTER
func (s PostService) GetAllPosts(ctx *gin.Context) ([]model.Post, error) {
	getStatus := s.r.RedisClient.Get(ctx, "posts")

	if getStatus.Err() == nil {
		var posts []model.Post
		err := json.Unmarshal([]byte(getStatus.Val()), &posts)

		if err != nil {
			return nil, err
		}

		return posts, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	posts, err := s.r.GetAllPosts()

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(posts)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "posts", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return posts, nil
}

func (s PostService) GetPostByID(id uint64, ctx *gin.Context) (*model.Post, error) {
	getStatus := s.r.RedisClient.Get(ctx, "post:id:"+strconv.FormatUint(id, 10))

	if getStatus.Err() == nil {
		var post model.Post
		err := json.Unmarshal([]byte(getStatus.Val()), &post)

		if err != nil {
			return nil, err
		}

		return &post, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	post, err := s.r.GetPostByID(id)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(post)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "post:id:"+strconv.FormatUint(id, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return post, nil
}

func (s PostService) GetPostsByThreadID(threadID uint64, ctx *gin.Context) ([]model.Post, error) {
	getStatus := s.r.RedisClient.Get(ctx, "posts:thread_id:"+strconv.FormatUint(threadID, 10))

	if getStatus.Err() == nil {
		var posts []model.Post
		err := json.Unmarshal([]byte(getStatus.Val()), &posts)

		if err != nil {
			return nil, err
		}

		return posts, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	posts, err := s.r.GetPostsByThreadID(threadID)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(posts)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "posts:thread_id:"+strconv.FormatUint(threadID, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return posts, nil
}

func (s PostService) GetPostsByAuthorID(authorID uint64, ctx *gin.Context) ([]model.Post, error) {
	getStatus := s.r.RedisClient.Get(ctx, "posts:author_id:"+strconv.FormatUint(authorID, 10))

	if getStatus.Err() == nil {
		var posts []model.Post
		err := json.Unmarshal([]byte(getStatus.Val()), &posts)

		if err != nil {
			return nil, err
		}

		return posts, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	posts, err := s.r.GetPostsByAuthorID(authorID)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(posts)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "posts:author_id:"+strconv.FormatUint(authorID, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return posts, nil
}

func (s PostService) GetPostsByParentID(parentID uint64, ctx *gin.Context) ([]model.Post, error) {
	getStatus := s.r.RedisClient.Get(ctx, "post:parent:"+strconv.FormatUint(parentID, 10))

	if getStatus.Err() == nil {
		var posts []model.Post
		err := json.Unmarshal([]byte(getStatus.Val()), &posts)

		if err != nil {
			return nil, err
		}

		return posts, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	posts, err := s.r.GetPostsByParentID(parentID)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(posts)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "post:parent:"+strconv.FormatUint(parentID, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return posts, nil
}

func (s PostService) GetPostVotes(postID uint64, ctx *gin.Context) ([]model.Vote, error) {
	getStatus := s.r.RedisClient.Get(ctx, "post:votes:"+strconv.FormatUint(postID, 10))

	if getStatus.Err() == nil {
		var votes []model.Vote
		err := json.Unmarshal([]byte(getStatus.Val()), &votes)

		if err != nil {
			return nil, err
		}

		return votes, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	votes, err := s.r.GetPostVotes(postID)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(votes)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.r.RedisClient.Set(ctx, "post:votes:"+strconv.FormatUint(postID, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return votes, nil
}

// SETTER
func (s *PostService) Create(
	ThreadID uint,
	Content string,
	AuthorID uint,
	ParentId *uint,
	ctx *gin.Context,
) (*model.Post, error) {
	post := &model.Post{
		ThreadID: ThreadID,
		Content:  Content,
		AuthorID: AuthorID,
		ParentID: ParentId,
	}

	if ParentId != nil {
		parentPost, err := s.r.GetPostByID(uint64(*ParentId))

		if err != nil {
			return nil, err
		}

		if parentPost == nil {
			return nil, errors.New("Parent is not found!")
		}
	}

	err := s.r.Create(post)

	if err != nil {
		return nil, err
	}

	delParentStatus := s.r.RedisClient.Del(ctx, "post:id:"+strconv.FormatUint(uint64(post.ID), 10))

	if delParentStatus.Err() != nil {
		return nil, delParentStatus.Err()
	}

	delPostsStatus := s.r.RedisClient.Del(ctx, "posts")

	if delPostsStatus.Err() != nil {
		return nil, delPostsStatus.Err()
	}

	if ParentId != nil {
		delParentPostsStatus := s.r.RedisClient.Del(ctx, "post:parent:"+strconv.FormatUint(uint64(*ParentId), 10))

		if delParentPostsStatus.Err() != nil {
			return nil, delParentPostsStatus.Err()
		}
	}

	delAuthorStatus := s.r.RedisClient.Del(ctx, "posts:author_id:"+strconv.FormatUint(uint64(AuthorID), 10))

	if delAuthorStatus.Err() != nil {
		return nil, delAuthorStatus.Err()
	}

	delThreadStatus := s.r.RedisClient.Del(ctx, "posts:thread_id:"+strconv.FormatUint(uint64(ThreadID), 10))

	if delThreadStatus.Err() != nil {
		return nil, delThreadStatus.Err()
	}

	return post, nil
}

func (s *PostService) Update(
	ID uint64,
	Content *string,
	ctx *gin.Context,
) (*model.Post, error) {
	post, err := s.r.GetPostByID(ID)

	if err != nil {
		return nil, err
	}

	if post == nil {
		return nil, errors.New("Post not found")
	}

	if Content != nil {
		post.Content = *Content
	}

	post.IsEdited = true

	err = s.r.Update(post)

	if err != nil {
		return nil, err
	}

	delParentStatus := s.r.RedisClient.Del(ctx, "post:id:"+strconv.FormatUint(uint64(post.ID), 10))

	if delParentStatus.Err() != nil {
		return nil, delParentStatus.Err()
	}

	delPostsStatus := s.r.RedisClient.Del(ctx, "posts")

	if delPostsStatus.Err() != nil {
		return nil, delPostsStatus.Err()
	}

	delParentPostsStatus := s.r.RedisClient.Del(ctx, "post:parent:"+strconv.FormatUint(uint64(*post.ParentID), 10))

	if delParentPostsStatus.Err() != nil {
		return nil, delParentPostsStatus.Err()
	}

	delAuthorStatus := s.r.RedisClient.Del(ctx, "posts:author_id:"+strconv.FormatUint(uint64(post.AuthorID), 10))

	if delAuthorStatus.Err() != nil {
		return nil, delAuthorStatus.Err()
	}

	delThreadStatus := s.r.RedisClient.Del(ctx, "posts:thread_id:"+strconv.FormatUint(uint64(post.ThreadID), 10))

	if delThreadStatus.Err() != nil {
		return nil, delThreadStatus.Err()
	}

	return post, nil
}

func (s *PostService) Delete(ID uint64, ctx *gin.Context) error {
	post, err := s.r.GetPostByID(ID)

	if err != nil {
		return err
	}

	if post == nil {
		return errors.New("Post not found")
	}

	replies := post.Posts

	for _, reply := range replies {
		err = s.Delete(uint64(reply.ID), ctx)

		if err != nil {
			return err
		}
	}

	delErr := s.r.Delete(post)

	if delErr != nil {
		return delErr
	}

	delParentStatus := s.r.RedisClient.Del(ctx, "post:id:"+strconv.FormatUint(uint64(post.ID), 10))

	if delParentStatus.Err() != nil {
		return delParentStatus.Err()
	}

	delPostsStatus := s.r.RedisClient.Del(ctx, "posts")

	if delPostsStatus.Err() != nil {
		return delPostsStatus.Err()
	}

	delParentPostsStatus := s.r.RedisClient.Del(ctx, "post:parent:"+strconv.FormatUint(uint64(*post.ParentID), 10))

	if delParentPostsStatus.Err() != nil {
		return delParentPostsStatus.Err()
	}

	delAuthorStatus := s.r.RedisClient.Del(ctx, "posts:author_id:"+strconv.FormatUint(uint64(post.AuthorID), 10))

	if delAuthorStatus.Err() != nil {
		return delAuthorStatus.Err()
	}

	delThreadStatus := s.r.RedisClient.Del(ctx, "posts:thread_id:"+strconv.FormatUint(uint64(post.ThreadID), 10))

	if delThreadStatus.Err() != nil {
		return delThreadStatus.Err()
	}

	return nil
}

func (s *PostService) Vote(postID uint64, userID uint64, value int, ctx *gin.Context) error {
	post, err := s.r.GetPostByID(postID)

	if err != nil {
		return err
	}

	if post == nil {
		return errors.New("Post not found")
	}

	voteValue, vErr := enum.GetVoteFromValue(value)

	if vErr != nil {
		return vErr
	}

	isUpvote := voteValue == enum.VoteUp

	var vote model.Vote

	fErr := s.r.GormDB.Where("post_id = ? AND user_id = ?", post.ID, userID).First(&vote).Error

	if fErr != nil && err != gorm.ErrRecordNotFound {
		return fErr
	}

	if err == gorm.ErrRecordNotFound {
		vote = model.Vote{
			PostID: post.ID,
			UserID: uint(userID),
			Value:  0,
		}
	}

	if isUpvote {
		post.VoteScore = post.VoteScore + 1
		s.r.GormDB.Save(post)

		vote.Value = int(enum.VoteUp)
		return s.r.GormDB.Save(&vote).Error
	}

	post.VoteScore = post.VoteScore - 1
	s.r.GormDB.Save(post)

	vote.Value = int(enum.VoteDown)

	uErr := s.r.GormDB.Save(&vote).Error

	if uErr != nil {
		return uErr
	}

	delStatus := s.r.RedisClient.Del(ctx, "post:votes:"+strconv.FormatUint(postID, 10))

	if delStatus.Err() != nil {
		return delStatus.Err()
	}

	return nil
}

func (s *PostService) React(postID uint64, userID uint64, emoji int, ctx *gin.Context) error {
	emojiValue, eErr := enum.EmojiFromInt(emoji)

	if eErr != true {
		return errors.New("Emoji is not registered")
	}

	post, err := s.r.GetPostByID(postID)

	if err != nil {
		return err
	}

	if post == nil {
		return errors.New("Post not found")
	}

	reaction := model.Reaction{
		PostId: post.ID,
		UserId: uint(userID),
		Emoji:  emojiValue.String(),
	}

	var existsReaction model.Reaction

	fErr := s.r.GormDB.
		Where("post_id = ? AND user_id = ?", post.ID, userID).
		First(&existsReaction).Error

	if fErr != nil && err != gorm.ErrRecordNotFound {
		return fErr
	}

	if existsReaction.ID != 0 {
		return s.r.GormDB.Delete(&existsReaction).Error
	}

	if existsReaction.Emoji == reaction.Emoji {
		return s.r.GormDB.Delete(&existsReaction).Error
	}

	if existsReaction.Emoji != reaction.Emoji {
		err = s.r.GormDB.Delete(&existsReaction).Error
		if err != nil {
			return err
		}
	}

	err = s.r.GormDB.Create(&reaction).Error

	if err != nil {
		return err
	}

	delStatus := s.r.RedisClient.Del(ctx, "post:reactions:"+strconv.FormatUint(postID, 10))

	if delStatus.Err() != nil {
		return delStatus.Err()
	}

	return nil
}

func (s *PostService) MarkAsSolution(postID uint64, userID uint64, ctx *gin.Context) error {
	post, err := s.r.GetPostByID(postID)

	if err != nil {
		return err
	}

	if post == nil {
		return errors.New("Post not found")
	}

	var thread model.Thread
	err = s.r.GormDB.Where("id = ?", post.ThreadID).First(&thread).Error

	if err != nil {
		return err
	}

	if thread.ID == 0 {
		return errors.New("Thread not found")
	}

	if thread.AuthorID != uint(userID) {
		return errors.New("Unauthorized")
	}

	if post.AuthorID != uint(userID) {
		return errors.New("Unauthorized")
	}

	posts := thread.Posts

	var hasSolution bool

	for _, p := range posts {
		if p.ID == post.ID {
			continue
		}

		if p.IsSolution {
			hasSolution = true
		}
	}

	if hasSolution {
		return errors.New("Thread already has a solution")
	}

	uErr := s.r.GormDB.Model(&model.Post{}).
		Where("id = ?", postID).
		Update("is_solution", true).Error

	if uErr != nil {
		return uErr
	}

	delStatus := s.r.RedisClient.Del(ctx, "post:parent:"+strconv.FormatUint(uint64(post.ID), 10))

	if delStatus.Err() != nil {
		return delStatus.Err()
	}

	delAuthorStatus := s.r.RedisClient.Del(ctx, "posts:author_id:"+strconv.FormatUint(uint64(post.AuthorID), 10))

	if delAuthorStatus.Err() != nil {
		return delAuthorStatus.Err()
	}

	delThreadStatus := s.r.RedisClient.Del(ctx, "posts:thread_id:"+strconv.FormatUint(uint64(post.ThreadID), 10))

	if delThreadStatus.Err() != nil {
		return delThreadStatus.Err()
	}

	return nil
}

func (s *PostService) CreateAttachment(post *model.Post, attachment *model.Attachment, ctx *gin.Context) (*model.Attachment, error) {

	createdAttachment, err := s.r.CreateAttachment(uint64(post.ID), attachment)

	if err != nil {
		return nil, err
	}

	delStatus := s.r.RedisClient.Del(ctx, "attachments:post:"+strconv.FormatUint(uint64(post.ID), 10))

	if delStatus.Err() != nil {
		return nil, delStatus.Err()
	}

	return createdAttachment, nil
}
