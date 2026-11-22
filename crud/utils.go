package crud

import (
	"errors"
	"fmt"
	"reflect"
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
	// MaxJoinDepth caps the number of nested levels in a join path. <=0 uses DefaultMaxJoinDepth.
	MaxJoinDepth int
	// AllowedJoins optionally restricts joins to these relation paths (or their parents).
	// Empty means any relation of the model may be joined.
	AllowedJoins []string
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

func (q *QueryToDBConverter) maxJoinDepth() int {
	if q.MaxJoinDepth > 0 {
		return q.MaxJoinDepth
	}
	return DefaultMaxJoinDepth
}

// normalizeJoinPath splits a dotted relation path and title-cases each segment,
// matching how relation field names are written on the model.
func normalizeJoinPath(relation string) ([]string, error) {
	segments := strings.Split(relation, ".")
	titled := make([]string, len(segments))
	for i, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			return nil, fmt.Errorf("%w: %q has an empty segment", ErrInvalidJoin, relation)
		}
		titled[i] = cases.Title(language.English, cases.NoLower).String(segment)
	}
	return titled, nil
}

// joinAllowed reports whether path equals an allowlist entry or is a parent of one.
func (q *QueryToDBConverter) joinAllowed(path string) bool {
	for _, entry := range q.AllowedJoins {
		segments, err := normalizeJoinPath(entry)
		if err != nil {
			continue
		}
		allowed := strings.Join(segments, ".")
		if allowed == path || strings.HasPrefix(allowed, path+".") {
			return true
		}
	}
	return false
}

// relationKey identifies one relation by the model that owns it and its name.
type relationKey struct {
	owner reflect.Type
	name  string
}

// validateJoin checks one join path against the depth cap, the model's relation
// graph, the optional allowlist and the rule that no relation may repeat within
// the path, and returns the canonical path to preload.
func (q *QueryToDBConverter) validateJoin(tx *gorm.DB, relation string) (string, error) {
	segments, err := normalizeJoinPath(relation)
	if err != nil {
		return "", err
	}
	if maxDepth := q.maxJoinDepth(); len(segments) > maxDepth {
		return "", fmt.Errorf("%w: %q has %d levels, max is %d", ErrJoinDepthExceeded, relation, len(segments), maxDepth)
	}
	current := ensureSchema(tx)
	if current == nil {
		return "", fmt.Errorf("%w: %q: model schema unavailable", ErrInvalidJoin, relation)
	}
	visited := make(map[relationKey]bool, len(segments))
	for _, name := range segments {
		rel, ok := current.Relationships.Relations[name]
		if !ok || rel.FieldSchema == nil {
			return "", fmt.Errorf("%w: %q is not a relation of the model", ErrInvalidJoin, relation)
		}
		key := relationKey{owner: current.ModelType, name: name}
		if visited[key] {
			return "", fmt.Errorf("%w: %q uses relation %s more than once", ErrInvalidJoin, relation, name)
		}
		visited[key] = true
		current = rel.FieldSchema
	}
	path := strings.Join(segments, ".")
	if len(q.AllowedJoins) > 0 && !q.joinAllowed(path) {
		return "", fmt.Errorf("%w: %q is not allowed", ErrInvalidJoin, relation)
	}
	return path, nil
}

// relationsMapper validates every comma-separated join entry before preloading any
// of them. On error it also records the error on tx, so the query cannot run.
func (q *QueryToDBConverter) relationsMapper(joinString string, tx *gorm.DB) error {
	var paths []string
	for _, relation := range strings.Split(joinString, ",") {
		relation = strings.TrimSpace(relation)
		if relation == "" {
			continue
		}
		path, err := q.validateJoin(tx, relation)
		if err != nil {
			tx.AddError(err)
			return err
		}
		paths = append(paths, path)
	}
	for _, path := range paths {
		tx.Preload(path)
	}
	return nil
}

func (q *QueryToDBConverter) filterMapper(filters []string, tx *gorm.DB) {
	for _, filter := range filters {
		// limit to 3 parts so the value itself may contain the separator
		// (e.g. "guest_id||eq||x||y" -> value is "x||y")
		filterParams := strings.SplitN(filter, SEPARATOR, 3)
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
