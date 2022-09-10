package crud

import (
	"gorm.io/gorm"
)

type Repo[T any] interface {
	FindOne(cond *T, dest *T) error
	Update(cond *T, updatedColumns interface{}) error
	Delete(cond *T) error
	Create(data *T) error
	getTx() *gorm.DB
}

type Repository[T any] struct {
	DB    *gorm.DB
	Model *T
}

func (r *Repository[T]) FindOne(cond *T, dest *T) error {
	return r.DB.Where(cond).First(dest).Error
}

func (r *Repository[T]) Update(cond *T, updatedColumns interface{}) error {
	return r.DB.Model(r.Model).Where(cond).UpdateColumns(updatedColumns).Error
}

func (r *Repository[T]) Delete(cond *T) error {
	if err := r.DB.Model(r.Model).Delete(cond); err != nil {
		return err.Error
	}
	return nil
}

func (r *Repository[T]) Create(data *T) error {
	return r.DB.Create(data).Error
}

func (r *Repository[T]) getTx() *gorm.DB {
	return r.DB.Model(r.Model)
}

func NewRepository[T any](db *gorm.DB, model *T) Repo[T] {
	return &Repository[T]{
		DB:    db,
		Model: model,
	}
}
