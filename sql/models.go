package sql

import (
	"github.com/google/uuid"
)

type Authentication struct {
	ID        uuid.UUID `json:"id"`
	Prefix    string    `json:"prefix"`
	Key       string    `json:"key"`
	ExpiresIn int       `json:"expires_in"`
}
