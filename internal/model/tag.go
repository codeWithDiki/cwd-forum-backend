package model

import "time"

type Tag struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug" gorm:"unique index"`
	Color      string    `json:"color"`
	UsageCount int       `json:"usage_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Associations
	Threads []Thread `gorm:"many2many:thread_tags" json:"threads,omitempty"`
}
