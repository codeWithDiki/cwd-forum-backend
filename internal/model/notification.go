package model

import (
	"time"
)

type Notification struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	ThreadId  *uint      `json:"thread_id,omitempty"`
	PostId    *uint      `json:"post_id,omitempty"`
	UserId    uint       `json:"user_id"`
	Type      string     `json:"type"`
	Payload   string     `json:"payload"`
	IsRead    bool       `json:"is_read"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`

	// Associations
	User   User    `gorm:"foreignKey:UserId" json:"user"`
	Thread *Thread `gorm:"foreignKey:ThreadId" json:"thread,omitempty"`
	Post   *Post   `gorm:"foreignKey:PostId" json:"post,omitempty"`
}
