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
				log.Printf("field -> %+v", field)
				fieldQuery, ok := field.(map[string]map[string]interface{})
				if ok {
					log.Printf("field is ok-> %+v", field)
					for whereField, whereValStruct := range fieldQuery {
						for whereOperator, whereValue := range whereValStruct {
							operator, ok := filterConditions[whereOperator]
							if ok {
								log.Printf("whereField -> %+v whereValue -> %+v operator -> %+v", whereField, whereValue, operator)
								//tx = tx.Where(fmt.Sprintf("%s %s ?", whereField, operator), whereValue)
							}
						}
					}
				} else {
					keyAndVal, ok := field.(map[string]interface{})
					if ok {
						for whereField, whereVal := range keyAndVal {
							whereValMap, ok := whereVal.(map[string]interface{})
							if ok {
								log.Printf("dynamic -> %+v", whereValMap)
								for operatorKey, value := range whereValMap {
									operator, ok := filterConditions[operatorKey]
									if ok {
										tx = tx.Where(fmt.Sprintf("%s %s ?", whereField, operator), value)
									}
								}

							} else {

								tx = tx.Where(whereField, whereVal)
							}
						}
					}
				}
			}
		}
	}
	return nil, tx
}
