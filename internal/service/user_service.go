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
func (s UserService) GetAllUsers(ctx *gin.Context) ([]model.User, error) {
	getStatus := s.Repo.RedisClient.Get(ctx, "users")

	if getStatus.Err() == nil {
		var users []model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &users)

		if err != nil {
			return nil, err
		}

		return users, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	users, err := s.Repo.GetAllUsers()

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(users)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.Repo.RedisClient.Set(ctx, "users", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return users, nil
}

func (s UserService) GetUserByID(id uint64, ctx *gin.Context) (*model.User, error) {
	getStatus := s.Repo.RedisClient.Get(ctx, "user:"+strconv.FormatUint(id, 10))

	if getStatus.Err() == nil {
		var user model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &user)

		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	user, err := s.Repo.GetUserByID(id)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(user)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.Repo.RedisClient.Set(ctx, "user:"+strconv.FormatUint(id, 10), json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return user, nil
}

func (s UserService) GetUserByUsername(username string, ctx *gin.Context) (*model.User, error) {
	getStatus := s.Repo.RedisClient.Get(ctx, "user:username:"+username)

	if getStatus.Err() == nil {
		var user model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &user)

		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	user, err := s.Repo.GetUserByUsername(username)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(user)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.Repo.RedisClient.Set(ctx, "user:username:"+username, json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return user, nil
}

func (s UserService) GetUserByEmail(email string, ctx *gin.Context) (*model.User, error) {
	getStatus := s.Repo.RedisClient.Get(ctx, "user:email:"+email)

	if getStatus.Err() == nil {
		var user model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &user)

		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	user, err := s.Repo.GetUserByEmail(email)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(user)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.Repo.RedisClient.Set(ctx, "user:email:"+email, json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return user, nil
}

func (s UserService) GetFollowers(userID uint64, ctx *gin.Context) ([]model.User, error) {
	getStatus := s.Repo.RedisClient.Get(ctx, "user:"+strconv.FormatUint(userID, 10)+":followers")

	if getStatus.Err() == nil {
		var followers []model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &followers)

		if err != nil {
			return nil, err
		}

		return followers, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	followers, err := s.Repo.GetFollowers(userID)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(followers)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.Repo.RedisClient.Set(ctx, "user:"+strconv.FormatUint(userID, 10)+":followers", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return followers, nil
}

func (s UserService) GetFollowing(userID uint64, ctx *gin.Context) ([]model.User, error) {
	getStatus := s.Repo.RedisClient.Get(ctx, "user:"+strconv.FormatUint(userID, 10)+":following")

	if getStatus.Err() == nil {
		var following []model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &following)

		if err != nil {
			return nil, err
		}

		return following, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		return nil, getStatus.Err()
	}

	following, err := s.Repo.GetFollowing(userID)

	if err != nil {
		return nil, err
	}

	json, err := json.Marshal(following)

	if err != nil {
		return nil, err
	}

	cmdStatus := s.Repo.RedisClient.Set(ctx, "user:"+strconv.FormatUint(userID, 10)+":following", json, time.Hour)

	if cmdStatus.Err() != nil {
		return nil, cmdStatus.Err()
	}

	return following, nil
}

// SETTER
func (s *UserService) CreateUser(
	Name string,
	Username string,
	Email string,
	Password string,
	Avatar string,
	Bio string,
	ctx *gin.Context,
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

	delStatus := s.Repo.RedisClient.Del(ctx, "users", "user:"+strconv.FormatUint(uint64(user.ID), 10), "user:username:"+user.Username, "user:email:"+user.Email)

	if delStatus.Err() != nil {
		return nil, delStatus.Err()
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
	ctx *gin.Context,
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

	uErr := s.Repo.Update(user)

	if uErr != nil {
		return nil, uErr
	}

	delStatus := s.Repo.RedisClient.Del(ctx, "users", "user:"+strconv.FormatUint(uint64(user.ID), 10), "user:username:"+user.Username, "user:email:"+user.Email)

	if delStatus.Err() != nil {
		return nil, delStatus.Err()
	}

	return user, nil
}

func (s *UserService) DeleteUser(ID uint64, ctx *gin.Context) error {
	user, err := s.Repo.GetUserByID(ID)

	if err != nil {
		return err
	}

	dErr := s.Repo.Delete(user)

	if dErr != nil {
		return dErr
	}

	delStatus := s.Repo.RedisClient.Del(ctx, "users", "user:"+strconv.FormatUint(uint64(user.ID), 10), "user:username:"+user.Username, "user:email:"+user.Email)

	if delStatus.Err() != nil {
		return delStatus.Err()
	}

	return nil
}
