package model

import "time"

type Vote struct {
	ID        uint      `gorm:"primaryKey"`
	PostID    uint      `json:"post_id"`
	UserID    uint      `json:"user_id"`
	Value     int       `json:"value"` // +1 for upvote, -1 for downvote
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Associations
	Post Post `gorm:"foreignKey:PostId" json:"-"`
	User User `gorm:"foreignKey:UserId" json:"user"`
}
