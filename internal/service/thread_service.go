package service

import (
	"errors"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/utils"
)

type ThreadService struct {
	r *repository.ThreadRepository
}

func NewThreadService(r *repository.ThreadRepository) *ThreadService {
	return &ThreadService{
		r: r,
	}
}

// GETTER
func (s ThreadService) GetAllThreads() ([]model.Thread, error) {
	return s.r.GetAllThreads()
}

func (s ThreadService) GetThreadByID(id uint64) (*model.Thread, error) {
	return s.r.GetThreadByID(id)
}

func (s ThreadService) GetThreadBySlug(slug string) (*model.Thread, error) {
	return s.r.GetThreadBySlug(slug)
}

func (s ThreadService) GetThreadsByCategoryID(categoryID uint) ([]model.Thread, error) {
	return s.r.GetThreadsByCategoryID(categoryID)
}

func (s ThreadService) GetThreadsByAuthorID(authorID uint) ([]model.Thread, error) {
	return s.r.GetThreadsByAuthorID(authorID)
}

func (s ThreadService) GetThreadsByTagID(tagID uint) ([]model.Thread, error) {
	return s.r.GetThreadsByTagID(tagID)
}

// SETTER
func (s *ThreadService) Create(
	CategoryID uint,
	Title string,
	Slug string,
	Content string,
	AuthorID uint,
	TagIDs []uint,
) (*model.Thread, *model.Post, error) {
	thread := &model.Thread{
		CategoryID: CategoryID,
		Title:      Title,
		Slug:       Slug,
		AuthorID:   AuthorID,
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

	slugExists, _ := s.r.GetThreadBySlug(Slug)

	if slugExists != nil {
		thread.Slug = Slug + "-" + utils.String(5)
	}

	err := s.r.Create(thread)

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

	return thread, post, nil
}

func (s *ThreadService) Update(
	ID uint64,
	CategoryID *uint,
	Title *string,
	Slug *string,
	IsPinned *bool,
	IsLocked *bool,
	IsSolved *bool,
) (*model.Thread, error) {
	thread, err := s.r.GetThreadByID(ID)

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
		slugExists, _ := s.r.GetThreadBySlug(*Slug)

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

	err = s.r.Update(thread)

	if err != nil {
		return nil, err
	}

	return thread, nil
}

func (s *ThreadService) Delete(ID uint64) error {
	thread, err := s.r.GetThreadByID(ID)

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

	return s.r.Delete(thread)
}

func (s *ThreadService) CreatePostAttachment(post *model.Post, attachment *model.Attachment) error {
	return s.r.CreatePostAttachment(post, attachment)
}

func (s *ThreadService) CanMarkAsSolution(threadID uint64, userID uint64) (bool, error) {
	thread, err := s.r.GetThreadByID(threadID)

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
