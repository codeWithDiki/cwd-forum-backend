package service

import (
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

// GETTER
func (s UserService) GetAllUsers() ([]model.User, error) {
	return s.Repo.GetAllUsers()
}

func (s UserService) GetUserByID(id uint64) (*model.User, error) {
	return s.Repo.GetUserByID(id)
}

func (s UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.Repo.GetUserByUsername(username)
}

func (s UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.Repo.GetUserByEmail(email)
}

func (s UserService) Login(
	Username string,
	Password string,
) (string, error) {
	user, err := s.Repo.GetUserByUsername(Username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(Password),
	)

	if err != nil {
		return "", errors.New("invalid credentials")
	}

	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// SETTER
func (s *UserService) CreateUser(
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

	usernameExists, _ := s.Repo.GetUserByUsername(user.Username)

	if usernameExists != nil {
		return nil, errors.New("Username already exists")
	}

	emailExists, _ := s.Repo.GetUserByEmail(user.Email)

	if emailExists != nil {
		return nil, errors.New("Email already exists")
	}

	err := s.Repo.Create(user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(
	ID uint64,
	Name *string,
	Username *string,
	Email *string,
	Password *string,
	Avatar *string,
	Bio *string,
) (*model.User, error) {
	user, err := s.Repo.GetUserByID(ID)

	var errorBags []error

	if err != nil {
		return nil, err
	}

	if Name != nil {
		user.Name = *Name
	}

	if Username != nil {
		existingUser, _ := s.Repo.GetUserByUsername(*Username)

		if existingUser != nil && existingUser.ID != user.ID {
			errorBags = append(errorBags, errors.New("Username already exists"))
		}

		user.Username = *Username
	}

	if Email != nil {
		existingUser, _ := s.Repo.GetUserByEmail(*Email)

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

	return user, s.Repo.Update(user)
}

func (s *UserService) DeleteUser(ID uint64) error {
	user, err := s.Repo.GetUserByID(ID)

	if err != nil {
		return err
	}

	return s.Repo.Delete(user)
}
