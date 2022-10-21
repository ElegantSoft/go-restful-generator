package crud

const (
	AND           = "$and"
	OR            = "$or"
	SEPARATOR     = "||"
	SortSeparator = ","
)

type GetAllRequest struct {
	Page   int      `json:"page" form:"page"`
	Limit  int      `json:"limit" form:"limit"`
	Join   string   `json:"join" form:"join"`
	S      string   `json:"s" form:"s"`
	Fields string   `json:"fields" form:"fields"`
	Filter []string `json:"filter" form:"filter"`
	Sort   []string `json:"sort" form:"sort"`
}

var filterConditions = map[string]string{
	"eq":      "=",
	"ne":      "!=",
	"gt":      ">",
	"lt":      "<",
	"gte":     ">=",
	"lte":     "<=",
	"$in":     "in",
	"cont":    "ILIKE",
	"isnull":  "IS NULL",
	"notnull": "IS NOT NULL",
}

type ById struct {
	ID string `uri:"id" binding:"required"`
}
