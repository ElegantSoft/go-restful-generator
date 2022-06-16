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

func searchMapper(s map[string]interface{}, tx *gorm.DB) error {
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
									tx = tx.Where(fmt.Sprintf("%s %s ?", whereField, operator), value)
								}
							}

						} else {

							tx = tx.Where(whereField, whereVal)
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
										tx = tx.Where(fmt.Sprintf("%s %s ?", whereField, operator), value)
									} else {
										tx = tx.Or(fmt.Sprintf("%s %s ?", whereField, operator), value)
									}
								}
							}

						} else {
							if i == 0 {
								tx = tx.Where(whereField, whereVal)
							} else {
								tx = tx.Or(whereField, whereVal)
							}
						}
					}
				}
			}

		}

	}
	return nil
}

func relationsMapper(joinString string, tx *gorm.DB) {
	relations := strings.Split(joinString, ",")
	for _, relation := range relations {
		log.Printf("will preload -> %v", strcase.ToCamel(relation))
		tx.Preload(strcase.ToCamel(relation))
	}
}
