package service

import (
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"

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
func (s PostService) GetAllPosts() ([]model.Post, error) {
	return s.r.GetAllPosts()
}

func (s PostService) GetPostByID(id uint64) (*model.Post, error) {
	return s.r.GetPostByID(id)
}

func (s PostService) GetPostsByThreadID(threadID uint64) ([]model.Post, error) {
	return s.r.GetPostsByThreadID(threadID)
}

func (s PostService) GetPostsByAuthorID(authorID uint64) ([]model.Post, error) {
	return s.r.GetPostsByAuthorID(authorID)
}

func (s PostService) GetPostsByParentID(parentID uint64) ([]model.Post, error) {
	return s.r.GetPostsByParentID(parentID)
}

func (s PostService) GetPostVotes(postID uint64) ([]model.Vote, error) {
	return s.r.GetPostVotes(postID)
}

// SETTER
func (s *PostService) Create(
	ThreadID uint,
	Content string,
	AuthorID uint,
	ParentId *uint,
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
	return post, nil
}

func (s *PostService) Update(
	ID uint64,
	Content *string,
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
	return post, nil
}

func (s *PostService) Delete(ID uint64) error {
	post, err := s.r.GetPostByID(ID)

	if err != nil {
		return err
	}

	if post == nil {
		return errors.New("Post not found")
	}

	replies := post.Posts

	for _, reply := range replies {
		err = s.Delete(uint64(reply.ID))

		if err != nil {
			return err
		}
	}

	return s.r.Delete(post)
}

func (s *PostService) Vote(postID uint64, userID uint64, value int) error {
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
	return s.r.GormDB.Save(&vote).Error
}

func (s *PostService) React(postID uint64, userID uint64, emoji int) error {
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

	return s.r.GormDB.Create(&reaction).Error
}

func (s *PostService) MarkAsSolution(postID uint64, userID uint64) error {
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

	return s.r.GormDB.Model(&model.Post{}).
		Where("id = ?", postID).
		Update("is_solution", true).Error
}

func (s *PostService) CreateAttachment(post *model.Post, attachment *model.Attachment) (*model.Attachment, error) {
	return s.r.CreateAttachment(uint64(post.ID), attachment)
}
