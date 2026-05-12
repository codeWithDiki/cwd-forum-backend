package model

import "time"

type ModerationLog struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ModeratorId  uint       `json:"moderator_id"`
	TargetUserId *uint      `json:"target_user_id"`
	TargetPostId *uint      `json:"target_post_id"`
	Action       string     `json:"action"` // e.g., "delete_post", "ban_user"
	Reason       string     `json:"reason"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Associations
	Moderator  User  `gorm:"foreignKey:ModeratorId" json:"moderator"`
	TargetUser *User `gorm:"foreignKey:TargetUserId" json:"target_user,omitempty"`
	TargetPost *Post `gorm:"foreignKey:TargetPostId" json:"target_post,omitempty"`
}
