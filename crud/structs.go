package crud

const (
	AND       = "$and"
	OR        = "$or"
	SEPARATOR = "||"
)

type GetAll struct {
	Page   uint     `json:"page" form:"page"`
	Limit  uint     `json:"limit" form:"limit"`
	Join   []string `json:"join" form:"join"`
	S      string   `json:"s" form:"s"`
	Fields []string `json:"fields" form:"fields"`
	Sort   []string `json:"sort" form:"sort"`
}

var filterConditions = map[string]string{
	"$eq":   "=",
	"$ne":   "!=",
	"$gt":   ">",
	"$lt":   "<",
	"$gte":  ">=",
	"$lte":  "<=",
	"$cont": "LIKE",
}
