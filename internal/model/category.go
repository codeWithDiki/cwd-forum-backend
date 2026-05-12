package model

import (
	"time"
)

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ParentID    *uint     `json:"parent_id"` // Nullable for top-level categories
	Name        string    `json:"name"`
	Slug        string    `json:"slug" gorm:"unique index"`
	Description string    `json:"description"`
	IconUrl     string    `json:"icon_url"`
	SortOrder   int       `json:"sort_order"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Associations
	Threads    []Thread   `gorm:"foreignKey:CategoryID" json:"threads,omitempty"`
	Parent     *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Categories []Category `gorm:"foreignKey:ParentID" json:"subcategories,omitempty"`
}
