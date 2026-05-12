package repository

import (
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"

	"gorm.io/gorm"
)

type PostRepository struct {
	GormDB *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{
		GormDB: db,
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

func (r *PostRepository) Vote(post *model.Post, userID uint64, isUpvote bool) error {
	var vote model.Vote

	err := r.GormDB.Where("post_id = ? AND user_id = ?", post.ID, userID).First(&vote).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
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
		r.GormDB.Save(post)

		vote.Value = int(enum.VoteUp)
		return r.GormDB.Save(&vote).Error
	}

	post.VoteScore = post.VoteScore - 1
	r.GormDB.Save(post)

	vote.Value = int(enum.VoteDown)
	return r.GormDB.Save(&vote).Error

}

func (r *PostRepository) Reaction(post *model.Post, userID uint64, emoji int) error {
	emojiValue, eErr := enum.EmojiFromInt(emoji)

	if eErr != true {
		return errors.New("Emoji is not registered")
	}

	reaction := model.Reaction{
		PostId: post.ID,
		UserId: uint(userID),
		Emoji:  emojiValue.String(),
	}

	var existsReaction model.Reaction

	err := r.GormDB.
		Where("post_id = ? AND user_id = ?", post.ID, userID).
		First(&existsReaction).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if existsReaction.ID != 0 {
		return r.GormDB.Delete(&existsReaction).Error
	}

	if existsReaction.Emoji == reaction.Emoji {
		return r.GormDB.Delete(&existsReaction).Error
	}

	if existsReaction.Emoji != reaction.Emoji {
		err = r.GormDB.Delete(&existsReaction).Error
		if err != nil {
			return err
		}
	}

	return r.GormDB.Create(&reaction).Error
}

func (r *PostRepository) CreateAttachment(postID uint64,
	attachment *model.Attachment) (*model.Attachment, error) {

	r.GormDB.Model(&model.Post{ID: uint(postID)}).Association("Attachments").Append(attachment)

	return attachment, nil
}
