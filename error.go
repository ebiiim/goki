package goki

import (
	"errors"
	"fmt"
)

var (
	// ErrDBOpen represents could not open DB error.
	ErrDBOpen = errors.New("could not open DB")
	// ErrDBClose represents could not close DB error.
	ErrDBClose = errors.New("could not close DB")
	// ErrDBSave represents could not save DB error.
	ErrDBSave = errors.New("could not save DB")
	// ErrDBInternal represents DB internal error.
	ErrDBInternal = errors.New("DB internal error")
	// ErrAppClose represents could not close App error.
	ErrAppClose = errors.New("could not close App")
	// ErrUserNotFound represents user not found error.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExist represents user already exist error.
	ErrUserAlreadyExist = errors.New("user already exist")
)

// ErrWrap returns a new error.
// Zero or one cause is accepted.
func ErrWrap(err error, cause ...error) error {
	if len(cause) == 0 {
		return err
	}
	return fmt.Errorf("%v: %w", err, cause[0])
}
