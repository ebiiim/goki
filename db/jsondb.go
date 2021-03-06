package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/ebiiim/goki"
	"github.com/ebiiim/goki/model"
)

// JSONUserDB is an easy UserDB stores data in a JSON file.
// Cannot be read from multiple app instances.
type JSONUserDB struct {
	filePath string
	db       map[string]*model.User
	mu       sync.Mutex
}

var _ UserDB = (*JSONUserDB)(nil)

// NewJSONUserDB initializes a JSONUserDB
func NewJSONUserDB(filePath string) (*JSONUserDB, error) {
	d := &JSONUserDB{
		filePath: filePath,
		db:       map[string]*model.User{},
	}
	if isFile(d.filePath) {
		if err := d.load(); err != nil {
			return nil, goki.ErrWrap(goki.ErrDBOpen, err)
		}
	} else {
		if err := d.save(); err != nil {
			return nil, goki.ErrWrap(goki.ErrDBOpen, err)
		}
	}
	return d, nil
}

func (d *JSONUserDB) load() error {
	// Load JSON.
	f, err := ioutil.ReadFile(d.filePath)
	if err != nil {
		return goki.ErrWrap(goki.ErrDBOpen, err)
	}
	var db map[string]*model.User
	if err := json.Unmarshal(f, &db); err != nil {
		return goki.ErrWrap(goki.ErrDBOpen, err)
	}
	d.db = db
	return nil
}

func (d *JSONUserDB) save() error {
	b, err := json.Marshal(d.db)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(d.filePath, b, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Close saves data to the database JSON file.
func (d *JSONUserDB) Close() error {
	if err := d.save(); err != nil {
		return goki.ErrWrap(goki.ErrDBClose, err)
	}
	return nil
}

// Get gets an user or error.
func (d *JSONUserDB) Get(userID string) (*model.User, error) {
	d.mu.Lock()
	u, ok := d.db[userID]
	d.mu.Unlock()
	if !ok {
		return nil, goki.ErrUserNotFound
	}
	var uu model.User
	deepCopy(&uu, u)
	return &uu, nil
}

// GetByTwitterID gets an user by Twitter ID or error.
func (d *JSONUserDB) GetByTwitterID(twitterID string) (*model.User, error) {
	var uu model.User
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, u := range d.db {
		if u.Twitter.ID == twitterID {
			deepCopy(&uu, u)
			return &uu, nil
		}
	}
	return nil, goki.ErrUserNotFound
}

// Add adds an user.
func (d *JSONUserDB) Add(user *model.User) error {
	if _, err := d.Get(user.ID); err == nil {
		return goki.ErrUserAlreadyExist
	}
	if _, err := d.GetByTwitterID(user.Twitter.ID); err == nil {
		return goki.ErrUserAlreadyExist
	}
	u := model.NewUser(user.ID, user.Name, user.Twitter.ID)
	d.mu.Lock()
	d.db[user.ID] = u
	d.mu.Unlock()
	if err := d.save(); err != nil {
		return goki.ErrWrap(goki.ErrDBSave, err)
	}
	return nil
}

// JSONActivityDB is an easy ActivityDB stores data in a JSON file.
// Cannot be read from multiple app instances.
type JSONActivityDB struct {
	filePath string
	// UserID -> time.Unix -> Activity
	db map[string]map[int64]*model.Activity
	mu sync.Mutex
}

var _ ActivityDB = (*JSONActivityDB)(nil)

// NewJSONActivityDB initializes a JSONActivityDB
func NewJSONActivityDB(filePath string) (*JSONActivityDB, error) {
	d := &JSONActivityDB{
		filePath: filePath,
		db:       map[string]map[int64]*model.Activity{},
	}
	if isFile(d.filePath) {
		if err := d.load(); err != nil {
			return nil, goki.ErrWrap(goki.ErrDBOpen, err)
		}
	} else {
		if err := d.save(); err != nil {
			return nil, goki.ErrWrap(goki.ErrDBOpen, err)
		}
	}
	return d, nil
}

func (d *JSONActivityDB) load() error {
	// If filePath does not exist then create a new file.
	if _, err := os.Stat(d.filePath); err != nil {
		f, err := os.Create(d.filePath)
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	// Load JSON.
	f, err := ioutil.ReadFile(d.filePath)
	if err != nil {
		return err
	}
	var db map[string]map[int64]*model.Activity
	if err := json.Unmarshal(f, &db); err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *JSONActivityDB) save() error {
	b, err := json.Marshal(d.db)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(d.filePath, b, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Close saves data to the database JSON file.
func (d *JSONActivityDB) Close() error {
	if err := d.save(); err != nil {
		return goki.ErrWrap(goki.ErrDBClose, err)
	}
	return nil
}

// Add adds an activity.
// This method DOES NOT validate Activity.UserID in the given activity.
// In this UserDB implementation, if an activity in DB has same timestamp with the given activity, then store the given one with timestamp++.
// Always returns nil
func (d *JSONActivityDB) Add(act *model.Activity) error {
	// init
	if chk, ok := d.db[act.UserID]; !ok || chk == nil {
		d.db[act.UserID] = map[int64]*model.Activity{}
	}
	ut := act.TimeUTC.Unix()
	d.mu.Lock()
	for {
		_, ok := d.db[act.UserID][ut]
		if ok {
			ut++
			continue
		}
		a := model.NewActivity(act.UserID, time.Unix(ut, 0).In(time.UTC), act.G.S, act.G.M, act.G.L)
		d.db[act.UserID][ut] = a
		break
	}
	d.mu.Unlock()
	if err := d.save(); err != nil {
		return goki.ErrWrap(goki.ErrDBSave, err)
	}
	return nil
}

// Query returns a slice of Activity (may be empty).
// This method DOES NOT validate Activity.UserID in the given activity.
// Just returns an empty slice when the given userID is invalid.
// Always returns nil
func (d *JSONActivityDB) Query(userID string, queryFn func(a *model.Activity) bool) ([]*model.Activity, error) {
	var ret []*model.Activity
	d.mu.Lock()
	defer d.mu.Unlock()
	al, ok := d.db[userID]
	if !ok {
		return ret, nil
	}
	for _, v := range al {
		if queryFn(v) {
			var act model.Activity
			deepCopy(&act, v)
			ret = append(ret, &act)
		}
	}
	return ret, nil
}

func isFile(filePath string) bool {
	s, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	if s.IsDir() {
		return false
	}
	return true
}

func deepCopy(dst interface{}, src interface{}) error {
	if dst == nil || src == nil {
		return goki.ErrWrap(goki.ErrDBInternal, errors.New("deepCopy: dst or src is nil"))
	}
	b, err := json.Marshal(src)
	if err != nil {
		return goki.ErrWrap(goki.ErrDBInternal, fmt.Errorf("could not encode to JSON: %w", err))
	}
	if err := json.Unmarshal(b, dst); err != nil {
		return goki.ErrWrap(goki.ErrDBInternal, fmt.Errorf("could not decode from JSON: %w", err))
	}
	return nil
}
