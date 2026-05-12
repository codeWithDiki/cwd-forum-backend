package model

import "time"

type Thread struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	CategoryID uint       `json:"category_id"`
	AuthorID   uint       `json:"author_id"`
	PinnedBy   *uint      `json:"pinned_by"` // Nullable, indicates who pinned the thread
	Title      string     `json:"title" gorm:"index"`
	Slug       string     `json:"slug" gorm:"unique index"`
	IsPinned   bool       `json:"is_pinned"`
	IsLocked   bool       `json:"is_locked"`
	IsSolved   bool       `json:"is_solved"`
	LastPostAt *time.Time `json:"last_post_at"` // Nullable, indicates the time of the last post in the thread
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	// Associations
	Posts        []Post   `gorm:"foreignKey:ThreadID" json:"posts,omitempty"`
	Category     Category `gorm:"foreignKey:CategoryID" json:"-"`
	Author       User     `gorm:"foreignKey:AuthorId" json:"-"`
	PinnedByUser *User    `gorm:"foreignKey:PinnedBy" json:"pinned_by_user,omitempty"`
	Tags         []Tag    `gorm:"many2many:thread_tags" json:"tags,omitempty"`
}
