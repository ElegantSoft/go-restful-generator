package models

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid; default:uuid_generate_v4()"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CategoryID  uuid.UUID `json:"category_id"`
	Category    Category  `json:"category"`
	Price       uint32    `json:"price"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Category struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid; default:uuid_generate_v4()"`
	Name      string    `json:"name,omitempty"`
	Posts     []Post    `json:"posts,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
