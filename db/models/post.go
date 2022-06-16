package models

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	ID          uuid.UUID `json:"id,omitempty" gorm:"type:uuid; default:uuid_generate_v4()"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	CategoryID  uuid.UUID `json:"category_id,omitempty"`
	Category    Category  `json:"category,omitempty"`
	Price       uint32    `json:"price,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type Category struct {
	ID        uuid.UUID `json:"id,omitempty" gorm:"type:uuid; default:uuid_generate_v4()"`
	Name      string    `json:"name,omitempty,omitempty"`
	Posts     []Post    `json:"posts,omitempty,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}
