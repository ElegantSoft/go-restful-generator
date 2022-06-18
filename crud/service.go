package crud

import (
	"strings"
)

type Service[T any] struct {
	repo *Repository[T]
	qtb  *queryToDBConverter
}

func (svc *Service[T]) Find(api GetAllRequest, result *[]T, totalRows *int64) error {
	var s map[string]interface{}

	tx := svc.repo.getTx()

	if len(api.Fields) > 0 {
		fields := strings.Split(api.Fields, ",")
		tx.Select(fields)
	}
	if len(api.Join) > 0 {
		svc.qtb.relationsMapper(api.Join, tx)
	}
	if api.Page > 0 {
		tx.Limit(int(api.Limit)).Offset(int((api.Page - 1) * api.Limit))
	}

	if len(api.Filter) > 0 {
		svc.qtb.filterMapper(api.Filter, tx)
	}

	if len(api.Sort) > 0 {
		svc.qtb.sortMapper(api.Sort, tx)
	}

	err := svc.qtb.searchMapper(s, tx)
	if err != nil {
		return err
	}
	tx.Count(totalRows)
	tx.Find(&result)
	return nil
}

func (svc *Service[T]) FindOne(api GetAllRequest, result *T) error {
	var s map[string]interface{}

	tx := svc.repo.getTx()

	if len(api.Fields) > 0 {
		fields := strings.Split(api.Fields, ",")
		tx.Select(fields)
	}
	if len(api.Join) > 0 {
		svc.qtb.relationsMapper(api.Join, tx)
	}

	if len(api.Filter) > 0 {
		svc.qtb.filterMapper(api.Filter, tx)
	}

	if len(api.Sort) > 0 {
		svc.qtb.sortMapper(api.Sort, tx)
	}

	err := svc.qtb.searchMapper(s, tx)
	if err != nil {
		return err
	}
	tx.First(&result)
	return nil
}

func (svc *Service[T]) Create(data *T) error {
	return svc.repo.Create(data)
}

func (svc *Service[T]) Delete(cond *T) error {
	return svc.repo.Delete(cond)
}

func NewService[T any](repo *Repository[T]) *Service[T] {
	return &Service[T]{
		repo: repo,
		qtb:  &queryToDBConverter{},
	}
}
