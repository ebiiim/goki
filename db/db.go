package db

import (
	"io"
	"time"

	"github.com/ebiiim/goki/model"
)

// UserDB interface provides User operations.
type UserDB interface {
	io.Closer
	Get(userID string) (*model.User, error)
	GetByTwitterID(twitterID string) (*model.User, error)
	Add(user *model.User) error
}

// ActivityDB interface provides Activity operations.
type ActivityDB interface {
	io.Closer
	Add(activity *model.Activity) error
	Query(userID string, queryFn func(a *model.Activity) bool) ([]*model.Activity, error)
}

// QueryFuncTime returns a queryFn for ActivityDB.Query method.
func QueryFuncTime(afterUTC time.Time, beforeUTC time.Time) func(a *model.Activity) bool {
	return func(a *model.Activity) bool {
		if a.TimeUTC.After(afterUTC) && a.TimeUTC.Before(beforeUTC) {
			return true
		}
		return false
	}
}
