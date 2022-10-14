package common

import "github.com/google/uuid"

type ById struct {
	ID uuid.UUID `uri:"id" binding:"required"`
}
