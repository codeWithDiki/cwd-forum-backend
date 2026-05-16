package service

import (
	"errors"
	"fmt"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gammazero/workerpool"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostService struct {
	log *logger.Logger
	r   *repository.PostRepository
}

func NewPostService(log *logger.Logger, r *repository.PostRepository) *PostService {
	return &PostService{
		log: log,
		r:   r,
	}
}

// GETTER
func (s PostService) GetAllPosts(ctx *gin.Context) ([]model.Post, error) {
	posts, err := s.r.GetAllPosts(ctx)
	s.log.Debug(ctx, "GetAllPosts Service", s.log.Field("Count", len(posts)))

	if err != nil {
		s.log.Error(ctx, "GetAllPosts Service Error", err)
		return nil, err
	}

	return posts, nil
}

func (s PostService) GetPostByID(ctx *gin.Context, id uint64) (*model.Post, error) {

	post, err := s.r.GetPostByID(ctx, id)
	s.log.Debug(ctx, "GetPostByID Service", s.log.Field("PostID", id))

	if err != nil {
		s.log.Error(ctx, "GetPostByID Service Error", err, s.log.Field("PostID", id))
		return nil, err
	}

	return post, nil
}

func (s PostService) GetPostsByThreadID(ctx *gin.Context, threadID uint64) ([]model.Post, error) {

	posts, err := s.r.GetPostsByThreadID(ctx, threadID)
	s.log.Debug(ctx, "GetPostsByThreadID Service", s.log.Field("ThreadID", threadID), s.log.Field("Count", len(posts)))

	if err != nil {
		s.log.Error(ctx, "GetPostsByThreadID Service Error", err, s.log.Field("ThreadID", threadID))
		return nil, err
	}

	return posts, nil
}

func (s PostService) GetPostsByAuthorID(ctx *gin.Context, authorID uint64) ([]model.Post, error) {

	posts, err := s.r.GetPostsByAuthorID(ctx, authorID)
	s.log.Debug(ctx, "GetPostsByAuthorID Service", s.log.Field("AuthorID", authorID), s.log.Field("Count", len(posts)))

	if err != nil {
		s.log.Error(ctx, "GetPostsByAuthorID Service Error", err, s.log.Field("AuthorID", authorID))
		return nil, err
	}

	return posts, nil
}

func (s PostService) GetPostsByParentID(ctx *gin.Context, parentID uint64) ([]model.Post, error) {

	posts, err := s.r.GetPostsByParentID(ctx, parentID)
	s.log.Debug(ctx, "GetPostsByParentID Service", s.log.Field("ParentID", parentID), s.log.Field("Count", len(posts)))

	if err != nil {
		s.log.Error(ctx, "GetPostsByParentID Service Error", err, s.log.Field("ParentID", parentID))
		return nil, err
	}

	return posts, nil
}

func (s PostService) GetPostVotes(ctx *gin.Context, postID uint64) ([]model.Vote, error) {

	votes, err := s.r.GetPostVotes(ctx, postID)
	s.log.Debug(ctx, "GetPostVotes Service", s.log.Field("PostID", postID), s.log.Field("Count", len(votes)))

	if err != nil {
		s.log.Error(ctx, "GetPostVotes Service Error", err, s.log.Field("PostID", postID))
		return nil, err
	}

	return votes, nil
}

// SETTER
func (s *PostService) Create(
	ctx *gin.Context,
	ThreadID uint,
	Content string,
	AuthorID uint,
	ParentId *uint,
	Attachments []*multipart.FileHeader,
) (*model.Post, error) {
	post := &model.Post{
		ThreadID: ThreadID,
		Content:  Content,
		AuthorID: AuthorID,
		ParentID: ParentId,
	}

	wp, wpExists := ctx.Get("workerPool")

	if !wpExists {
		return nil, errors.New("Failed to get worker pool")
	}

	if ParentId != nil {
		parentPost, err := s.r.GetPostByID(ctx, uint64(*ParentId))

		if err != nil {
			return nil, err
		}

		if parentPost == nil {
			return nil, errors.New("Parent is not found!")
		}
	}

	for _, file := range Attachments {

		wp.(*workerpool.WorkerPool).Submit(func() {
			fmt.Println("Uploading from Post")
			ext := filepath.Ext(file.Filename)
			newFileName := fmt.Sprintf("%d_%s%s", post.ID, uuid.New().String(), ext)

			s3client := ctx.MustGet("s3Client")
			fileBinary, err := file.Open()

			if err != nil {
				return
			}

			_, uErr := s3client.(*s3.S3).PutObject(&s3.PutObjectInput{
				Bucket: aws.String(os.Getenv("S3_BUCKET")),
				Key:    aws.String(newFileName), // You can customize the key as needed
				Body:   fileBinary,              // You should provide the actual file content here
				ACL:    aws.String("public-read"),
			})

			attachment := model.Attachment{
				PostID:     post.ID,
				UploaderId: post.AuthorID,
				Url:        fmt.Sprintf("%s/%s/%s", os.Getenv("S3_FILE_URL"), os.Getenv("S3_BUCKET"), newFileName),
				Filename:   newFileName,
				MimeType:   file.Header.Get("Content-Type"),
				FileSize:   file.Size,
			}

			post.Attachments = append(post.Attachments, attachment)

			s.CreateAttachment(ctx, post, &attachment)

			if uErr != nil {
				return
			}

		})
	}

	err := s.r.Create(ctx, post)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) Update(
	ctx *gin.Context,
	ID uint64,
	Content *string,
) (*model.Post, error) {
	post, err := s.r.GetPostByID(ctx, ID)

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

	err = s.r.Update(ctx, post)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) Delete(ctx *gin.Context, ID uint64) error {
	post, err := s.r.GetPostByID(ctx, ID)

	if err != nil {
		return err
	}

	if post == nil {
		return errors.New("Post not found")
	}

	replies := post.Posts

	for _, reply := range replies {
		err = s.Delete(ctx, uint64(reply.ID))

		if err != nil {
			return err
		}
	}

	err = s.r.Delete(ctx, post)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostService) Vote(ctx *gin.Context, postID uint64, userID uint64, value int) error {
	post, err := s.GetPostByID(ctx, postID)

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

	return nil
}

func (s *PostService) React(ctx *gin.Context, postID uint64, userID uint64, emoji int) error {
	emojiValue, eErr := enum.EmojiFromInt(emoji)

	if eErr != true {
		return errors.New("Emoji is not registered")
	}

	post, err := s.r.GetPostByID(ctx, postID)

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

	return nil
}

func (s *PostService) MarkAsSolution(ctx *gin.Context, postID uint64, userID uint64) error {
	post, err := s.r.GetPostByID(ctx, postID)

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

	return nil
}

func (s *PostService) CreateAttachment(ctx *gin.Context, post *model.Post, attachment *model.Attachment) (*model.Attachment, error) {

	createdAttachment, err := s.r.CreateAttachment(ctx, uint64(post.ID), attachment)

	if err != nil {
		return nil, err
	}

	return createdAttachment, nil
}
