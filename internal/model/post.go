package model

import "time"

type Post struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ThreadID   uint      `json:"thread_id"`
	AuthorID   uint      `json:"author_id"`
	ParentID   *uint     `json:"parent_id"` // Nullable for top-level posts
	Content    string    `json:"content"`
	IsDeleted  bool      `json:"is_deleted"`
	IsEdited   bool      `json:"is_edited"`
	IsSolution bool      `json:"is_solution"`
	VoteScore  int       `json:"vote_score"`
	EditedAt   *int64    `json:"edited_at"` // Nullable, timestamp of last edit
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Thread      Thread       `gorm:"foreignKey:ThreadID" json:"-"`
	Author      User         `gorm:"foreignKey:AuthorID" json:"-"`
	Parent      *Post        `gorm:"foreignKey:ParentId" json:"parent,omitempty"`
	Posts       []Post       `gorm:"foreignKey:ParentId" json:"replies,omitempty"`
	Votes       []Vote       `gorm:"foreignKey:PostId" json:"votes,omitempty"`
	Reactions   []Reaction   `gorm:"foreignKey:PostId" json:"reactions,omitempty"`
	Attachments []Attachment `gorm:"foreignKey:PostId" json:"attachments,omitempty"`
}
