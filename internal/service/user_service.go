package service

import (
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	log  *logger.Logger
	Repo *repository.UserRepository
}

func NewUserService(log *logger.Logger, repo *repository.UserRepository) *UserService {
	return &UserService{
		log:  log,
		Repo: repo,
	}
}

// GETTER
func (s UserService) GetAllUsers(ctx *gin.Context) ([]model.User, error) {
	s.log.Debug(ctx, "Service GetAllUsers Called")

	users, err := s.Repo.GetAllUsers(ctx)

	if err != nil {
		s.log.Error(ctx, "Service GetAllUsers Error", err)
		return nil, err
	}

	return users, nil
}

func (s UserService) GetUserByID(ctx *gin.Context, id uint64) (*model.User, error) {
	s.log.Debug(ctx, "Service GetUserByID Called", s.log.Field("UserID", id))

	user, err := s.Repo.GetUserByID(ctx, id)

	if err != nil {
		s.log.Error(ctx, "Service GetUserByID Error", err, s.log.Field("UserID", id))
		return nil, err
	}

	return user, nil
}

func (s UserService) GetUserByUsername(ctx *gin.Context, username string) (*model.User, error) {

	user, err := s.Repo.GetUserByUsername(ctx, username)

	if err != nil {
		s.log.Error(ctx, "Repo GetUserByUsername Error", err, s.log.Field("Username", username))
		return nil, err
	}

	return user, nil
}

func (s UserService) GetUserByEmail(ctx *gin.Context, email string) (*model.User, error) {

	user, err := s.Repo.GetUserByEmail(ctx, email)

	if err != nil {
		s.log.Error(ctx, "Repo GetUserByEmail Error", err, s.log.Field("Email", email))
		return nil, err
	}

	return user, nil
}

func (s UserService) GetFollowers(ctx *gin.Context, userID uint64) ([]model.User, error) {

	followers, err := s.Repo.GetFollowers(ctx, userID)

	if err != nil {
		s.log.Error(ctx, "Repo GetFollowers Error", err, s.log.Field("UserID", userID))
		return nil, err
	}

	return followers, nil
}

func (s UserService) GetFollowing(ctx *gin.Context, userID uint64) ([]model.User, error) {

	following, err := s.Repo.GetFollowing(ctx, userID)

	if err != nil {
		s.log.Error(ctx, "Repo GetFollowing Error", err, s.log.Field("UserID", userID))
		return nil, err
	}

	return following, nil
}

// SETTER
func (s *UserService) CreateUser(
	ctx *gin.Context,
	Name string,
	Username string,
	Email string,
	Password string,
	Avatar string,
	Bio string,
) (*model.User, error) {
	user := &model.User{
		Name:     Name,
		Username: Username,
		Email:    Email,
		Password: Password,
		Avatar:   Avatar,
		Bio:      Bio,
		Role:     enum.RoleUser.String(),
	}

	usernameExists, _ := s.Repo.GetUserByUsername(ctx, user.Username)

	if usernameExists != nil {
		return nil, errors.New("Username already exists")
	}

	emailExists, _ := s.Repo.GetUserByEmail(ctx, user.Email)

	if emailExists != nil {
		return nil, errors.New("Email already exists")
	}

	err := s.Repo.Create(ctx, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(
	ctx *gin.Context,
	ID uint64,
	Name *string,
	Username *string,
	Email *string,
	Password *string,
	Avatar *string,
	Bio *string,
) (*model.User, error) {
	user, err := s.Repo.GetUserByID(ctx, ID)

	var errorBags []error

	if err != nil {
		return nil, err
	}

	if Name != nil {
		user.Name = *Name
	}

	if Username != nil {
		existingUser, _ := s.Repo.GetUserByUsername(ctx, *Username)

		if existingUser != nil && existingUser.ID != user.ID {
			errorBags = append(errorBags, errors.New("Username already exists"))
		}

		user.Username = *Username
	}

	if Email != nil {
		existingUser, _ := s.Repo.GetUserByEmail(ctx, *Email)

		if existingUser != nil && existingUser.ID != user.ID {
			errorBags = append(errorBags, errors.New("Email already exists"))
		}

		user.Email = *Email
	}

	if Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*Password), bcrypt.DefaultCost)

		if err != nil {
			errorBags = append(errorBags, err)
		}

		user.Password = string(hashedPassword)

	}

	if Avatar != nil {
		user.Avatar = *Avatar
	}

	if Bio != nil {
		user.Bio = *Bio
	}

	if errorBags != nil {
		return nil, errors.Join(errorBags...)
	}

	uErr := s.Repo.Update(ctx, user)

	if uErr != nil {
		return nil, uErr
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx *gin.Context, ID uint64) error {
	user, err := s.Repo.GetUserByID(ctx, ID)

	if err != nil {
		return err
	}

	dErr := s.Repo.Delete(ctx, user)

	if dErr != nil {
		return dErr
	}

	return nil
}

func (s *UserService) FollowUser(ctx *gin.Context, userID uint64, targetUserID uint64) error {
	isFollowing, err := s.Repo.IsFollowing(ctx, userID, targetUserID)

	if err != nil {
		return err
	}

	if isFollowing {
		return errors.New("Already following this user")
	}

	fErr := s.Repo.Follow(ctx, userID, targetUserID)

	if fErr != nil {
		return fErr
	}

	return nil
}

func (s *UserService) UnfollowUser(ctx *gin.Context, userID uint64, targetUserID uint64) error {
	isFollowing, err := s.Repo.IsFollowing(ctx, userID, targetUserID)

	if err != nil {
		return err
	}

	if !isFollowing {
		return errors.New("Not following this user")
	}

	uErr := s.Repo.Unfollow(ctx, userID, targetUserID)

	if uErr != nil {
		return uErr
	}

	return nil
}
