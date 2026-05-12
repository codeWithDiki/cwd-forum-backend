package model

import "time"

type Badge struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	IconUrl         string    `json:"icon_url"`
	CriteriaType    string    `json:"criteria_type"`  // e.g., "post_count", "like_count"
	CriteriaValue   int       `json:"criteria_value"` // e.g., 100 for post_count
	FontColor       string    `json:"font_color"`
	BackgroundColor string    `json:"background_color"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Associations
	Users []User `gorm:"many2many:user_badges" json:"users,omitempty"`
}
