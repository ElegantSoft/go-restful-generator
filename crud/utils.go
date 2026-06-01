package crud

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ContainOperator = "cont"
	NotNullOperator = "notnull"
	IsNullOperator  = "isnull"
	InOperator      = "$in"
)

var AndValueNotSlice = errors.New("the value of $and or $or not array")

type QueryToDBConverter struct {
}

// where applies a condition as AND (or==false) or OR (or==true).
func (q *QueryToDBConverter) where(tx *gorm.DB, or bool, query interface{}, args ...interface{}) {
	if or {
		tx.Or(query, args...)
	} else {
		tx.Where(query, args...)
	}
}

// applyCond validates the column and operator, then applies a parameterized
// condition. Unknown columns or operators are dropped instead of being
// interpolated into raw SQL.
func (q *QueryToDBConverter) applyCond(tx *gorm.DB, or bool, field, operatorKey string, value interface{}) {
	col, ok := resolveColumn(tx, field)
	if !ok {
		return
	}
	operator, known := filterConditions[operatorKey]
	if !known {
		return
	}
	qc := tx.Statement.Quote(col)

	switch operatorKey {
	case NotNullOperator, IsNullOperator:
		q.where(tx, or, fmt.Sprintf("%s %s", qc, operator))
	case InOperator:
		str, ok := value.(string)
		if !ok {
			return
		}
		q.where(tx, or, fmt.Sprintf("%s IN ?", qc), strings.Split(str, ","))
	case ContainOperator:
		q.where(tx, or, fmt.Sprintf("%s %s ?", qc, operator), fmt.Sprintf("%%%v%%", value))
	default:
		q.where(tx, or, fmt.Sprintf("%s %s ?", qc, operator), value)
	}
}

func (q *QueryToDBConverter) searchMapper(s map[string]interface{}, tx *gorm.DB) error {
	for k := range s {
		var or bool
		switch k {
		case AND:
			or = false
		case OR:
			or = true
		default:
			continue
		}

		vals, ok := s[k].([]interface{})
		if !ok {
			return AndValueNotSlice
		}

		for i, field := range vals {
			keyAndVal, ok := field.(map[string]interface{})
			if !ok {
				continue
			}
			for whereField, whereVal := range keyAndVal {
				useOr := or && i > 0
				if whereValMap, ok := whereVal.(map[string]interface{}); ok {
					for operatorKey, value := range whereValMap {
						q.applyCond(tx, useOr, whereField, operatorKey, value)
					}
				} else {
					// equality shorthand: {"field": "value"}
					if col, ok := resolveColumn(tx, whereField); ok {
						q.where(tx, useOr, fmt.Sprintf("%s = ?", tx.Statement.Quote(col)), whereVal)
					}
				}
			}
		}
	}
	return nil
}

func (q *QueryToDBConverter) relationsMapper(joinString string, tx *gorm.DB) {
	relations := strings.Split(joinString, ",")
	for _, relation := range relations {
		nestedRelationsSlice := strings.Split(relation, ".")
		titledSlice := make([]string, len(nestedRelationsSlice))
		for i, relation := range nestedRelationsSlice {
			titledSlice[i] = cases.Title(language.English, cases.NoLower).String(relation)
		}
		nestedRelation := strings.Join(titledSlice, ".")
		if len(nestedRelation) > 0 {
			tx.Preload(nestedRelation)
		}
	}
}

func (q *QueryToDBConverter) filterMapper(filters []string, tx *gorm.DB) {
	for _, filter := range filters {
		filterParams := strings.Split(filter, SEPARATOR)
		if len(filterParams) < 2 {
			continue
		}
		operator, ok := filterConditions[filterParams[1]]
		if !ok {
			continue
		}
		col, ok := resolveColumn(tx, filterParams[0])
		if !ok {
			continue
		}
		qc := tx.Statement.Quote(col)

		switch filterParams[1] {
		case NotNullOperator, IsNullOperator:
			tx.Where(fmt.Sprintf("%s %s", qc, operator))
		default:
			if len(filterParams) != 3 {
				continue
			}
			switch filterParams[1] {
			case ContainOperator:
				tx.Where(fmt.Sprintf("%s %s ?", qc, operator), fmt.Sprintf("%%%s%%", filterParams[2]))
			case InOperator:
				tx.Where(fmt.Sprintf("%s IN ?", qc), strings.Split(filterParams[2], ","))
			default:
				tx.Where(fmt.Sprintf("%s %s ?", qc, operator), filterParams[2])
			}
		}
	}
}

func (q *QueryToDBConverter) sortMapper(sorts []string, tx *gorm.DB) {
	for _, sort := range sorts {
		sortParams := strings.Split(sort, SortSeparator)
		col, ok := resolveColumn(tx, sortParams[0])
		if !ok {
			continue
		}
		desc := true
		if len(sortParams) == 2 && strings.EqualFold(sortParams[1], "asc") {
			desc = false
		}
		tx.Order(clause.OrderByColumn{
			Column: clause.Column{Name: col},
			Desc:   desc,
		})
	}
}
