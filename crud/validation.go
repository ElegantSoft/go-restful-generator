package crud

import (
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var schemaCache = &sync.Map{}

// ensureSchema makes sure the statement schema is parsed so identifiers can be
// validated against the model's real columns.
func ensureSchema(tx *gorm.DB) *schema.Schema {
	if tx.Statement.Schema != nil {
		return tx.Statement.Schema
	}
	if tx.Statement.Model != nil {
		if s, err := schema.Parse(tx.Statement.Model, schemaCache, tx.NamingStrategy); err == nil {
			tx.Statement.Schema = s
		}
	}
	return tx.Statement.Schema
}

// resolveColumn maps a user-supplied field name to its canonical DB column name.
// It returns ok=false when the field is not a real column on the model, which is
// the signal to drop the clause instead of interpolating attacker input into SQL.
func resolveColumn(tx *gorm.DB, field string) (string, bool) {
	s := ensureSchema(tx)
	if s == nil {
		return "", false
	}
	if f := s.LookUpField(field); f != nil && f.DBName != "" {
		return f.DBName, true
	}
	return "", false
}
