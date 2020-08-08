package goki

import (
	"fmt"
)

var (
	// ErrDBOpen represents could not open DB error.
	ErrDBOpen = func(err error) error {
		return fmt.Errorf("could not open DB: %w", err)
	}
	// ErrDBClose represents could not close DB error.
	ErrDBClose = func(err error) error {
		return fmt.Errorf("could not close DB: %w", err)
	}
	// ErrAppClose represents could not close App error.
	ErrAppClose = func(err error) error {
		return fmt.Errorf("could not close App: %w", err)
	}
	// ErrInvalidUser represents invalid user error.
	ErrInvalidUser = func(err error) error {
		return fmt.Errorf("invalid user: %w", err)
	}
	// ErrUserAlreadyExist represents user already exist error.
	ErrUserAlreadyExist = func(err error) error {
		return fmt.Errorf("user already exist: %w", err)
	}
)
