package goki

import (
	"time"

	"github.com/google/uuid"
)

// NewID generates a new random ID.
func NewID() string {
	return uuid.New().String()
}

// TimeNow is an alias of time.Now by default.
var TimeNow = time.Now
