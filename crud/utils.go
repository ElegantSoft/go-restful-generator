package crud

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
)

var AndValueNotSlice = errors.New("the value of $and not array")

func searchToQuery(s map[string]interface{}, tx *gorm.DB) (error, *gorm.DB) {
	for k := range s {
		if k == AND {
			vals, ok := s[k].([]interface{})
			if !ok {
				return AndValueNotSlice, nil
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
				return AndValueNotSlice, nil
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
	return nil, tx
}
