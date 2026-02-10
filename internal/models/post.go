package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text"`
	UserID    uint           `json:"user_id" gorm:"index;not null"`
	User      *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt time.Time      `json:"created_at" gorm:"index:,sort:desc"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,min=3,max=200"`
	Content string `json:"content" binding:"required"`
}

type PostResponse struct {
	ID        uint          `json:"id"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	UserID    uint          `json:"user_id"`
	Author    *UserResponse `json:"author,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// ToResponse converts Post to PostResponse
func (p *Post) ToResponse() PostResponse {
	resp := PostResponse{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		UserID:    p.UserID,
		CreatedAt: p.CreatedAt,
	}

	if p.User != nil {
		author := p.User.ToResponse()
		resp.Author = &author
	}

	return resp
}
