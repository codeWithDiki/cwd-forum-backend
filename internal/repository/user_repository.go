package repository

import (
	"gin-quickstart/internal/model"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	GormDB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		GormDB: db,
	}
}

// GETTER
func (r UserRepository) GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := r.GormDB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r UserRepository) GetUserByID(id uint64) (*model.User, error) {
	var user model.User
	err := r.GormDB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r UserRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.GormDB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.GormDB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SETTER
func (r *UserRepository) Create(user *model.User) error {
	return r.GormDB.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.GormDB.Save(user).Error
}

func (r *UserRepository) Delete(user *model.User) error {
	user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	return r.GormDB.Save(user).Error
}

func (r *UserRepository) HardDelete(user *model.User) error {
	return r.GormDB.Unscoped().Delete(user).Error
}

func (r *UserRepository) Restore(user *model.User) error {
	user.DeletedAt = gorm.DeletedAt{Time: time.Time{}, Valid: false}
	return r.GormDB.Save(user).Error
}
