package crud

const (
	AND       = "$and"
	OR        = "$or"
	SEPARATOR = "||"
)

type GetAll struct {
	Page   uint     `json:"page" form:"page"`
	Limit  uint     `json:"limit" form:"limit"`
	Join   string   `json:"join" form:"join"`
	S      string   `json:"s" form:"s"`
	Fields string   `json:"fields" form:"fields"`
	Filter []string `json:"filter" form:"filter"`
	Sort   []string `json:"sort" form:"sort"`
}

var filterConditions = map[string]string{
	"$eq":   "=",
	"$ne":   "!=",
	"$gt":   ">",
	"$lt":   "<",
	"$gte":  ">=",
	"$lte":  "<=",
	"$cont": "ILIKE",
}

//type Crud[T any] interface {
//	FindAllBase(str T) T
//}
//
//type crudImpl[T any] struct {
//}
//
//func (c crudImpl[T]) FindAllBase(str T) T {
//	//TODO implement me
//	panic("implement me")
//	return str
//}
//
//func InitCrud[T any]() Crud[T] {
//	return crudImpl[T]{}
//}
