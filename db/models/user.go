package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	ID          uint     `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CategoryID  uint     `json:"category_id"`
	Category    Category `json:"category"`
}

type Category struct {
	gorm.Model
	Name  string `json:"name,omitempty"`
	Posts []Post `json:"posts,omitempty"`
}
