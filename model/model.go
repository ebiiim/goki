package model

import "time"

// User contains user information.
type User struct {
	ID      string
	Name    string
	Twitter struct {
		ID string
	}
}

// NewUser initializes an User.
func NewUser(id, name, twitterID string) *User {
	return &User{
		ID:   id,
		Name: name,
		Twitter: struct {
			ID string
		}{
			ID: twitterID,
		},
	}
}

// Goki contains roaches.
type Goki struct {
	// S represents a small size roach.
	S int
	// M represents a middle size roach.
	M int
	// L represents a large size roach.
	L int
}

// Activity contains an activity.
type Activity struct {
	UserID  string
	TimeUTC time.Time
	// The number of roaches eliminated by this activity.
	G *Goki
}

// NewActivity initializes an Activity.
func NewActivity(userID string, timeUTC time.Time, numS, numM, numL int) *Activity {
	return &Activity{
		UserID:  userID,
		TimeUTC: timeUTC,
		G: &Goki{
			S: numS,
			M: numM,
			L: numL,
		},
	}
}
