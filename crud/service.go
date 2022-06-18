package crud

import (
	"gorm.io/gorm"
	"strings"
)

type Service[T any] struct {
	db    *gorm.DB
	model *T
	qtb   *queryToDBConverter
}

//func (r Service[T]) FindOne(cond *T, dest *T) error {
//	return r.db.Where(cond).First(dest).Error
//}

func (r Service[T]) Find(api GetAllRequest, result *[]T) error {
	var s map[string]interface{}

	tx := r.db.Model(r.model)

	if len(api.Fields) > 0 {
		fields := strings.Split(api.Fields, ",")
		tx.Select(fields)
	}
	if len(api.Join) > 0 {
		r.qtb.relationsMapper(api.Join, tx)
	}
	if api.Page > 0 {
		tx.Limit(int(api.Limit)).Offset(int((api.Page - 1) * api.Limit))
	}

	if len(api.Filter) > 0 {
		r.qtb.filterMapper(api.Filter, tx)
	}

	if len(api.Sort) > 0 {
		r.qtb.sortMapper(api.Sort, tx)
	}

	err := r.qtb.searchMapper(s, tx)
	if err != nil {
		return err
	}
	tx.Find(&result)
	return nil
}

func (r Service[T]) FindOne(api GetAllRequest, result *T) error {
	var s map[string]interface{}

	tx := r.db.Model(r.model)

	if len(api.Fields) > 0 {
		fields := strings.Split(api.Fields, ",")
		tx.Select(fields)
	}
	if len(api.Join) > 0 {
		r.qtb.relationsMapper(api.Join, tx)
	}

	if len(api.Filter) > 0 {
		r.qtb.filterMapper(api.Filter, tx)
	}

	if len(api.Sort) > 0 {
		r.qtb.sortMapper(api.Sort, tx)
	}

	err := r.qtb.searchMapper(s, tx)
	if err != nil {
		return err
	}
	tx.First(&result)
	return nil
}

func NewRepository[T any](db *gorm.DB, model *T) *Service[T] {
	return &Service[T]{
		db:    db,
		model: model,
		qtb:   &queryToDBConverter{},
	}
}
