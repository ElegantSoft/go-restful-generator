package crud

import (
	"encoding/json"
	"strings"
)

type Service[T any] struct {
	Repo Repo[T]
	Qtb  *QueryToDBConverter
}

func (svc *Service[T]) Find(api GetAllRequest, result interface{}, totalRows *int64) error {
	var s map[string]interface{}
	if len(api.S) > 0 {
		err := json.Unmarshal([]byte(api.S), &s)
		if err != nil {
			return err
		}
	}

	tx := svc.Repo.getTx()
	if len(api.Fields) > 0 {
		fields := strings.Split(api.Fields, ",")
		tx.Select(fields)
	}
	if len(api.Join) > 0 {
		svc.Qtb.relationsMapper(api.Join, tx)
	}

	if len(api.Filter) > 0 {
		svc.Qtb.filterMapper(api.Filter, tx)
	}

	if len(api.Sort) > 0 {
		svc.Qtb.sortMapper(api.Sort, tx)
	}

	err := svc.Qtb.searchMapper(s, tx)
	if err != nil {
		return err
	}
	tx.Count(totalRows)

	tx.Limit(api.Limit)

	if api.Page > 0 {
		tx.Offset((api.Page - 1) * api.Limit)
	}
	return tx.Find(result).Error
}

func (svc *Service[T]) FindOne(api GetAllRequest, result interface{}) error {
	var s map[string]interface{}
	if len(api.S) > 0 {
		err := json.Unmarshal([]byte(api.S), &s)
		if err != nil {
			return err
		}
	}

	tx := svc.Repo.getTx()

	if len(api.Fields) > 0 {
		fields := strings.Split(api.Fields, ",")
		tx.Select(fields)
	}
	if len(api.Join) > 0 {
		svc.Qtb.relationsMapper(api.Join, tx)
	}

	if len(api.Filter) > 0 {
		svc.Qtb.filterMapper(api.Filter, tx)
	}

	if len(api.Sort) > 0 {
		svc.Qtb.sortMapper(api.Sort, tx)
	}

	err := svc.Qtb.searchMapper(s, tx)
	if err != nil {
		return err
	}
	return tx.First(result).Error
}

func (svc *Service[T]) Create(data *T) error {
	return svc.Repo.Create(data)
}

func (svc *Service[T]) Delete(cond *T) error {
	return svc.Repo.Delete(cond)
}

func (svc *Service[T]) Update(cond *T, updatedColumns *T) error {
	return svc.Repo.Update(cond, updatedColumns)
}

func NewService[T any](repo Repo[T]) *Service[T] {
	return &Service[T]{
		Repo: repo,
		Qtb:  &QueryToDBConverter{},
	}
}
