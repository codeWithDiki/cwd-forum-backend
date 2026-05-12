package model

import "time"

type Reaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostId    uint      `json:"post_id"`
	UserId    uint      `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Associations
	Post Post `gorm:"foreignKey:PostId" json:"post"`
	User User `gorm:"foreignKey:UserId" json:"user"`
}
