package post

import "time"

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Slug      string    `gorm:"size:255;not null;uniqueIndex" json:"slug"`
	Summary   string    `gorm:"type:text" json:"summary"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Post) TableName() string {
	return "posts"
}
