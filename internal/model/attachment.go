package model

import "time"

type Attachment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PostID     uint      `json:"post_id"`
	UploaderId uint      `json:"uploader_id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	FileSize   int64     `json:"file_size"`
	Url        string    `json:"url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Associations
	Post     Post `gorm:"foreignKey:PostId" json:"-"`
	Uploader User `gorm:"foreignKey:UploaderId" json:"-"`
}
