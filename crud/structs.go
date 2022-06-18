package crud

const (
	AND           = "$and"
	OR            = "$or"
	SEPARATOR     = "||"
	SortSeparator = ","
)

type GetAllRequest struct {
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

type ById struct {
	ID string `uri:"id" binding:"required"`
}
