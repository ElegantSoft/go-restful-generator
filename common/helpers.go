package common

import (
	"golang.org/x/mod/modfile"
	"os"
	"reflect"
)

func Unique(intSlice []uint) []uint {
	keys := make(map[uint]bool)
	var list []uint
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func GetIdFromCtx(id interface{}) uint {
	idFloat, _ := id.(float64)
	return uint(idFloat)
}

func Contains(val interface{}, slice interface{}) bool {
	found := false
	for _, v := range slice.([]uint) {
		if v == val {
			found = true
		}
	}
	return found
}
func StringsContains(val string, slice []string) bool {
	found := false
	for _, v := range slice {
		if v == val {
			found = true
		}
	}
	return found
}

func HashIntersection(a interface{}, b interface{}) []interface{} {
	set := make([]interface{}, 0)
	hash := make(map[interface{}]bool)
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	for i := 0; i < av.Len(); i++ {
		el := av.Index(i).Interface()
		hash[el] = true
	}

	for i := 0; i < bv.Len(); i++ {
		el := bv.Index(i).Interface()
		if _, found := hash[el]; found {
			set = append(set, el)
		}
	}

	return set
}

func removeDuplicateAdjacent(checkText string) string {
	newText := ""
	for i := range []rune(checkText) {
		if i+1 < len([]rune(checkText)) {
			if []rune(checkText)[i] != []rune(checkText)[i+1] {
				newText += string([]rune(checkText)[i])
			}

		} else {
			newText += string([]rune(checkText)[i])
		}

	}
	return newText
}

func GetModuleName() string {
	goModBytes, err := os.ReadFile("../go.mod")
	if err != nil {
		panic(err)
	}

	modName := modfile.ModulePath(goModBytes)

	return modName
}
