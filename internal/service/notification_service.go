package service

import (
	"errors"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/email"
	"gin-quickstart/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

type NotificationService struct {
	log         *logger.Logger
	Repo        *repository.NotificationRepository
	emailClient *email.EmailClient
}

func NewNotificationService(log *logger.Logger, repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{
		log:         log,
		Repo:        repo,
		emailClient: email.NewEmailClient(),
	}
}

func (s NotificationService) GetNotificationsByUserID(ctx *gin.Context, userID uint64) ([]model.Notification, error) {
	return s.Repo.GetNotificationsByUserID(ctx, userID)
}

func (s NotificationService) GetNotificationByID(ctx *gin.Context, id uint64, userID uint64) (*model.Notification, error) {
	notification, err := s.Repo.GetNotificationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if notification.UserId != uint(userID) {
		return nil, errors.New("notification not found")
	}

	return notification, nil
}

func (s NotificationService) CreateNotification(ctx *gin.Context, notification *model.Notification) (*model.Notification, error) {
	if notification.UserId == 0 {
		return nil, errors.New("user_id is required")
	}

	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	err := s.Repo.Create(ctx, notification)
	if err != nil {
		return nil, err
	}

	// Send email notification if user has an email
	var user model.User
	if err := s.Repo.GormDB.Select("email").First(&user, notification.UserId).Error; err == nil && user.Email != "" {
		go s.emailClient.SendNotificationEmail(user.Email, notification.Type, notification.Payload)
	}

	return notification, nil
}

func (s NotificationService) MarkNotificationAsRead(ctx *gin.Context, id uint64, userID uint64) (*model.Notification, error) {
	notification, err := s.GetNotificationByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if notification.IsRead {
		return notification, nil
	}

	now := time.Now()
	notification.IsRead = true
	notification.ReadAt = &now
	notification.UpdatedAt = now

	err = s.Repo.Update(ctx, notification)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

func (s NotificationService) UpdateNotificationReadState(ctx *gin.Context, id uint64, userID uint64, isRead bool) (*model.Notification, error) {
	notification, err := s.GetNotificationByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	notification.IsRead = isRead
	if isRead {
		now := time.Now()
		notification.ReadAt = &now
	} else {
		notification.ReadAt = nil
	}
	notification.UpdatedAt = time.Now()

	err = s.Repo.Update(ctx, notification)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

func (s NotificationService) DeleteNotification(ctx *gin.Context, id uint64, userID uint64) error {
	notification, err := s.GetNotificationByID(ctx, id, userID)
	if err != nil {
		return err
	}

	return s.Repo.Delete(ctx, notification)
}
