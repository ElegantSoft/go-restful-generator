package crud

import (
	"gorm.io/gorm"
	"time"
)

type Repository[T any] struct {
	db    *gorm.DB
	model *T
}

func (r *Repository[T]) FindOne(cond *T, dest *T) error {
	return r.db.Where(cond).First(dest).Error
}

func (r *Repository[T]) Update(cond *T, updatedColumns interface{}) error {
	return r.db.Model(r.model).Select("*").Where(cond).UpdateColumns(updatedColumns).Error
}

func (r *Repository[T]) Delete(cond *T) error {
	if err := r.db.Model(r.model).Where(cond).Update("deleted_at", time.Now()); err != nil {
		return err.Error
	}
	return nil
}

func (r *Repository[T]) Create(data *T) error {
	return r.db.Create(data).Error
}

func (r *Repository[T]) getTx() *gorm.DB {
	return r.db.Model(r.model)
}

func NewRepository[T any](db *gorm.DB, model *T) *Repository[T] {
	return &Repository[T]{
		db:    db,
		model: model,
	}
}
