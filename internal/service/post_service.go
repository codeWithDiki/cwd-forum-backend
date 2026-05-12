package service

import (
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
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

	return s.r.Vote(post, userID, isUpvote)
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

	return s.r.Reaction(post, userID, int(emojiValue))
}

func (s *PostService) CreateAttachment(post *model.Post, attachment *model.Attachment) (*model.Attachment, error) {
	return s.r.CreateAttachment(uint64(post.ID), attachment)
}
