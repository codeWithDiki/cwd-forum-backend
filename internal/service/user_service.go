package service

import (
	"encoding/json"
	"errors"
	"gin-quickstart/internal/enum"
	"gin-quickstart/internal/model"
	"gin-quickstart/internal/repository"
	"gin-quickstart/pkg/logger"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	getStatus := s.Repo.RedisClient.Get(ctx, "users")

	if getStatus.Err() == nil {
		s.log.Debug(ctx, "Service GetAllUsers Cache Hit")
		var users []model.User
		err := json.Unmarshal([]byte(getStatus.Val()), &users)

		if err != nil {
			s.log.Error(ctx, "Service GetAllUsers JSON Unmarshal Error", err)
			return nil, err
		}

		return users, nil
	}

	if getStatus.Err() != nil && getStatus.Err() != redis.Nil {
		s.log.Error(ctx, "Get Status Error", getStatus.Err())
		return nil, getStatus.Err()
	}

	users, err := s.Repo.GetAllUsers(ctx)

	if err != nil {
		s.log.Error(ctx, "Repo GetAllUsers Error", err)
		return nil, err
	}

	json, err := json.Marshal(users)

	if err != nil {
		s.log.Error(ctx, "JSON Marshal Error", err)
		return nil, err
	}

	cmdStatus := s.Repo.RedisClient.Set(ctx, "users", json, time.Hour)

	if cmdStatus.Err() != nil {
		s.log.Error(ctx, "Set Status Error", cmdStatus.Err())
		return nil, cmdStatus.Err()
	}

	return users, nil
}

func (s UserService) GetUserByID(ctx *gin.Context, id uint64) (*model.User, error) {
	getStatus := s.Repo.RedisClient.Get(ctx, "user:"+strconv.FormatUint(id, 10))
	s.log.Debug(ctx, "Service GetUserByID Called")

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

	user, err := s.Repo.GetUserByID(ctx, id)

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

func (s UserService) GetUserByUsername(ctx *gin.Context, username string) (*model.User, error) {
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

	user, err := s.Repo.GetUserByUsername(ctx, username)

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

func (s UserService) GetUserByEmail(ctx *gin.Context, email string) (*model.User, error) {
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

	user, err := s.Repo.GetUserByEmail(ctx, email)

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

func (s UserService) GetFollowers(ctx *gin.Context, userID uint64) ([]model.User, error) {
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

	followers, err := s.Repo.GetFollowers(ctx, userID)

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

func (s UserService) GetFollowing(ctx *gin.Context, userID uint64) ([]model.User, error) {
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

	following, err := s.Repo.GetFollowing(ctx, userID)

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

	delStatus := s.Repo.RedisClient.Del(ctx, "users", "user:"+strconv.FormatUint(uint64(user.ID), 10), "user:username:"+user.Username, "user:email:"+user.Email)

	if delStatus.Err() != nil {
		return nil, delStatus.Err()
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

	delStatus := s.Repo.RedisClient.Del(ctx, "users", "user:"+strconv.FormatUint(uint64(user.ID), 10), "user:username:"+user.Username, "user:email:"+user.Email)

	if delStatus.Err() != nil {
		return nil, delStatus.Err()
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

	delStatus := s.Repo.RedisClient.Del(ctx, "users", "user:"+strconv.FormatUint(uint64(user.ID), 10), "user:username:"+user.Username, "user:email:"+user.Email)

	if delStatus.Err() != nil {
		return delStatus.Err()
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

	delStatus := s.Repo.RedisClient.Del(ctx, "user:"+strconv.FormatUint(userID, 10)+":following", "user:"+strconv.FormatUint(targetUserID, 10)+":followers")

	if delStatus.Err() != nil {
		return delStatus.Err()
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

	delStatus := s.Repo.RedisClient.Del(ctx, "user:"+strconv.FormatUint(userID, 10)+":following", "user:"+strconv.FormatUint(targetUserID, 10)+":followers")

	if delStatus.Err() != nil {
		return delStatus.Err()
	}

	return nil
}
