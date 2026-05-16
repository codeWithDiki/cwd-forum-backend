package service

import (
	"errors"
	"fmt"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"
	"gin-quickstart/pkg/utils"
	"gin-quickstart/pkg/worker"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ThreadService struct {
	log *logger.Logger
	r   *repository.ThreadRepository
}

func NewThreadService(log *logger.Logger, r *repository.ThreadRepository) *ThreadService {
	return &ThreadService{
		log: log,
		r:   r,
	}
}

// GETTER
func (s ThreadService) GetAllThreads(ctx *gin.Context) ([]model.Thread, error) {

	threads, err := s.r.GetAllThreads(ctx)
	s.log.Debug(ctx, "GetAllThreads Service", s.log.Field("Count", len(threads)))

	if err != nil {
		s.log.Error(ctx, "GetAllThreads Service Error", err)
		return nil, err
	}

	return threads, nil
}

func (s ThreadService) GetThreadByID(ctx *gin.Context, id uint64) (*model.Thread, error) {

	thread, err := s.r.GetThreadByID(ctx, id)
	s.log.Debug(ctx, "GetThreadByID Service", s.log.Field("ID", id))

	if err != nil {
		s.log.Error(ctx, "GetThreadByID Service Error", err, s.log.Field("ID", id))
		return nil, err
	}

	return thread, nil
}

func (s ThreadService) GetThreadBySlug(ctx *gin.Context, slug string) (*model.Thread, error) {

	thread, err := s.r.GetThreadBySlug(ctx, slug)
	s.log.Debug(ctx, "GetThreadBySlug Service", s.log.Field("Slug", slug))

	if err != nil {
		s.log.Error(ctx, "GetThreadBySlug Service Error", err, s.log.Field("Slug", slug))
		return nil, err
	}

	return thread, nil
}

func (s ThreadService) GetThreadsByCategoryID(ctx *gin.Context, categoryID uint) ([]model.Thread, error) {

	threads, err := s.r.GetThreadsByCategoryID(ctx, categoryID)
	s.log.Debug(ctx, "GetThreadsByCategoryID Service", s.log.Field("CategoryID", categoryID), s.log.Field("Count", len(threads)))

	if err != nil {
		s.log.Error(ctx, "GetThreadsByCategoryID Service Error", err, s.log.Field("CategoryID", categoryID))
		return nil, err
	}

	return threads, nil
}

func (s ThreadService) GetThreadsByAuthorID(ctx *gin.Context, authorID uint) ([]model.Thread, error) {

	threads, err := s.r.GetThreadsByAuthorID(ctx, authorID)
	s.log.Debug(ctx, "GetThreadsByAuthorID Service", s.log.Field("AuthorID", authorID), s.log.Field("Count", len(threads)))

	if err != nil {
		s.log.Error(ctx, "GetThreadsByAuthorID Service Error", err, s.log.Field("AuthorID", authorID))
		return nil, err
	}

	return threads, nil
}

func (s ThreadService) GetThreadsByTagID(ctx *gin.Context, tagID uint) ([]model.Thread, error) {

	threads, err := s.r.GetThreadsByTagID(ctx, tagID)
	s.log.Debug(ctx, "GetThreadsByTagID Service", s.log.Field("TagID", tagID), s.log.Field("Count", len(threads)))

	if err != nil {
		s.log.Error(ctx, "GetThreadsByTagID Service Error", err, s.log.Field("TagID", tagID))
		return nil, err
	}

	return threads, nil
}

// SETTER
func (s *ThreadService) Create(
	ctx *gin.Context,
	CategoryID uint,
	Title string,
	Slug string,
	Content string,
	AuthorID uint,
	TagIDs []uint,
	Attachments []*multipart.FileHeader,
) (*model.Thread, *model.Post, error) {
	thread := &model.Thread{
		CategoryID: CategoryID,
		Title:      Title,
		Slug:       Slug,
		AuthorID:   AuthorID,
	}

	wp, wpExists := ctx.Get("workerPool")

	if !wpExists {
		return nil, nil, errors.New("Worker pool not found in context")
	}

	var userExists bool

	uErr := s.r.GormDB.
		Model(&model.User{}).
		Where("id = ?", AuthorID).
		Select("count(*) > 0").
		Row().
		Scan(&userExists)

	if uErr != nil {
		return nil, nil, uErr
	}

	if !userExists {
		return nil, nil, errors.New("Author is not found!")
	}

	slugExists, _ := s.r.GetThreadBySlug(ctx, Slug)

	if slugExists != nil {
		thread.Slug = Slug + "-" + utils.String(5)
	}

	err := s.r.Create(ctx, thread)

	if err != nil {
		return nil, nil, err
	}

	var post *model.Post

	if Content != "" {
		post = &model.Post{
			ThreadID: thread.ID,
			Content:  Content,
			AuthorID: AuthorID,
		}

		pErr := s.r.GormDB.Create(post).Error

		if pErr != nil {
			return thread, nil, pErr
		}

		delPostCacheStatus := s.r.RedisClient.Del(ctx, "posts", "post:id:"+strconv.FormatUint(uint64(post.ID), 10))

		if delPostCacheStatus.Err() != nil {
			return thread, post, delPostCacheStatus.Err()
		}

	}

	if len(TagIDs) > 0 {
		var tags []model.Tag

		for _, tagID := range TagIDs {
			var tag model.Tag

			tErr := s.r.GormDB.First(&tag, tagID).Error

			if tErr != nil {
				return thread, post, nil
			}

			tags = append(tags, tag)

		}

		err = s.r.GormDB.Model(thread).Association("Tags").Append(&tags)

		if err != nil {
			return thread, post, err
		}
	}

	for _, file := range Attachments {

		wp.(*worker.WorkerPool).Worker.Submit(func() {
			fmt.Println("Uploading from Thread")
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

			s.CreatePostAttachment(ctx, post, &attachment)

			if uErr != nil {
				return
			}
		})
	}

	return thread, post, nil
}

func (s *ThreadService) Update(
	ctx *gin.Context,
	ID uint64,
	CategoryID *uint,
	Title *string,
	Slug *string,
	IsPinned *bool,
	IsLocked *bool,
	IsSolved *bool,
) (*model.Thread, error) {
	thread, err := s.GetThreadByID(ctx, ID)

	if err != nil {
		return nil, err
	}

	if thread == nil {
		return nil, errors.New("Thread not found")
	}

	if CategoryID != nil {
		thread.CategoryID = *CategoryID
	}

	if Title != nil {
		thread.Title = *Title
	}

	if Slug != nil {
		slugExists, _ := s.GetThreadBySlug(ctx, *Slug)

		if slugExists != nil && uint64(slugExists.ID) != ID {
			var newSlug string

			newSlug = *Slug + "-" + utils.String(5)

			Slug = &newSlug
		}

		thread.Slug = *Slug
	}

	if IsPinned != nil {
		thread.IsPinned = *IsPinned
	}

	if IsLocked != nil {
		thread.IsLocked = *IsLocked
	}

	if IsSolved != nil {
		thread.IsSolved = *IsSolved
	}

	err = s.r.Update(ctx, thread)

	if err != nil {
		return nil, err
	}

	return thread, nil
}

func (s *ThreadService) Delete(ctx *gin.Context, ID uint64) error {
	thread, err := s.r.GetThreadByID(ctx, ID)

	if err != nil {
		return err
	}

	if thread == nil {
		return errors.New("Thread not found")
	}

	posts := thread.Posts

	if posts != nil && len(thread.Posts) > 0 {
		for _, post := range posts {
			err = s.r.GormDB.Delete(&post).Error

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ThreadService) CreatePostAttachment(ctx *gin.Context, post *model.Post, attachment *model.Attachment) error {
	return s.r.CreatePostAttachment(ctx, post, attachment)
}

func (s *ThreadService) CanMarkAsSolution(ctx *gin.Context, threadID uint64, userID uint64) (bool, error) {
	thread, err := s.r.GetThreadByID(ctx, threadID)

	if err != nil {
		return false, err
	}

	if thread == nil {
		return false, errors.New("Thread not found")
	}

	if thread.AuthorID != uint(userID) {
		return false, errors.New("Unauthorized")
	}

	return true, nil
}
