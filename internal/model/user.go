package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name" binding:"required"`
	Username    string         `json:"username" gorm:"unique index" binding:"required,alphanum"`
	Email       string         `json:"email" gorm:"unique index" binding:"required,email"`
	Password    string         `json:"-" binding:"required,min=8"` // Exclude from JSON responses
	Avatar      string         `json:"avatar" binding:"omitempty,url"`
	Bio         string         `json:"bio" binding:"omitempty,max=500"`
	Role        string         `json:"role"` // e.g., "user", "moderator", "admin"
	IsBanned    bool           `json:"is_banned"`
	IsVerified  bool           `json:"is_verified"`
	LastSeenAt  *time.Time     `json:"last_seen_at"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UpdatedAt   time.Time      `json:"updated_at"`

	Threads        []Thread        `gorm:"foreignKey:AuthorId" json:"threads,omitempty"`
	Posts          []Post          `gorm:"foreignKey:AuthorId" json:"posts,omitempty"`
	Notifications  []Notification  `gorm:"foreignKey:UserId" json:"notifications,omitempty"`
	Badges         []Badge         `gorm:"many2many:user_badges" json:"badges,omitempty"`
	ModerationLogs []ModerationLog `gorm:"foreignKey:ModeratorId" json:"moderation_logs,omitempty"`
	Followings     []User          `gorm:"many2many:user_users;foreignKey:ID;joinForeignKey:FollowerID;References:ID;joinReferences:FollowedID"`
	Followers      []User          `gorm:"many2many:user_users;foreignKey:ID;joinForeignKey:FollowedID;References:ID;joinReferences:FollowerID"`
}

type UserUser struct {
	FollowerID           uint `gorm:"primaryKey"`
	FollowedID           uint `gorm:"primaryKey"`
	ActivateNotification bool `json:"activate_notification"` // Whether the follower wants notifications about the followed user's activity
	CreatedAt            time.Time
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {

	// If the password field is being updated, hash the new password

	return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Hash the password before saving to the database
	err := u.HashPassword()
	if err != nil {
		return err
	}

	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}
