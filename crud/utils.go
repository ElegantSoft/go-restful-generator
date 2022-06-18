package crud

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"gorm.io/gorm"
	"log"
	"strings"
)

var AndValueNotSlice = errors.New("the value of $and or $or not array")

type QueryToDBConverter struct {
}

func (q *QueryToDBConverter) searchMapper(s map[string]interface{}, tx *gorm.DB) error {
	for k := range s {
		if k == AND {
			vals, ok := s[k].([]interface{})
			if !ok {
				return AndValueNotSlice
			}
			for _, field := range vals {
				keyAndVal, ok := field.(map[string]interface{})
				if ok {
					for whereField, whereVal := range keyAndVal {
						whereValMap, ok := whereVal.(map[string]interface{})
						if ok {
							for operatorKey, value := range whereValMap {
								operator, ok := filterConditions[operatorKey]
								if ok {
									if operatorKey == "$cont" {
										value = fmt.Sprintf("%%%s%%", value)
										log.Println(value)
									}
									tx.Where(fmt.Sprintf("%s %s ?", whereField, operator), value)
								}
							}

						} else {

							tx.Where(whereField, whereVal)
						}
					}
				}
			}
		} else if k == OR {
			vals, ok := s[k].([]interface{})
			if !ok {
				return AndValueNotSlice
			}
			for i, field := range vals {
				keyAndVal, ok := field.(map[string]interface{})
				if ok {
					for whereField, whereVal := range keyAndVal {
						whereValMap, ok := whereVal.(map[string]interface{})
						if ok {
							for operatorKey, value := range whereValMap {
								operator, ok := filterConditions[operatorKey]
								if ok {
									if operatorKey == "$cont" {
										value = fmt.Sprintf("%%%s%%", value)
										log.Println(value)
									}
									if i == 0 {
										tx.Where(fmt.Sprintf("%s %s ?", whereField, operator), value)
									} else {
										tx.Or(fmt.Sprintf("%s %s ?", whereField, operator), value)
									}
								}
							}

						} else {
							if i == 0 {
								tx.Where(whereField, whereVal)
							} else {
								tx.Or(whereField, whereVal)
							}
						}
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
		tx.Preload(strcase.ToCamel(relation))
	}
}

func (q *QueryToDBConverter) filterMapper(filters []string, tx *gorm.DB) {
	for _, filter := range filters {
		filterParams := strings.Split(filter, SEPARATOR)
		if len(filterParams) == 3 {
			operator, ok := filterConditions[filterParams[1]]
			if ok {
				if filterParams[1] == "$cont" {
					tx.Where(fmt.Sprintf("%s %s ?", filterParams[0], operator), fmt.Sprintf("%%%s%%", filterParams[2]))
				} else {
					tx.Where(fmt.Sprintf("%s %s ?", filterParams[0], operator), filterParams[2])

				}
			}
		}
	}
}

func (q *QueryToDBConverter) sortMapper(sorts []string, tx *gorm.DB) {
	for _, sort := range sorts {
		sortParams := strings.Split(sort, SortSeparator)
		if len(sortParams) == 2 {
			tx.Order(fmt.Sprintf("%s %s", sortParams[0], strings.ToLower(sortParams[1])))
		} else {
			tx.Order(fmt.Sprintf("%s desc", sortParams[0]))
		}
	}
}
