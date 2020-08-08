package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/ebiiim/goki/model"
)

// UserDB interface provides User operations.
type UserDB interface {
	io.Closer
	Get(userID string) (*model.User, error)
	Add(user *model.User) error
}

// ActivityDB interface provides Activity operations.
type ActivityDB interface {
	io.Closer
	Add(activity *model.Activity) error
	Query(userID string, beginUTC, endUTC time.Time) ([]*model.Activity, error)
}

// JSONUserDB is an easy UserDB stores data in a JSON file.
// Cannot be read from multiple app instances.
type JSONUserDB struct {
	filePath string
	db       map[string]*model.User
	mu       sync.Mutex
}

// NewJSONUserDB initializes a JSONUserDB
func NewJSONUserDB(filePath string) (*JSONUserDB, error) {
	d := &JSONUserDB{
		filePath: filePath,
		db:       map[string]*model.User{},
	}
	if isFile(d.filePath) {
		if err := d.load(); err != nil {
			return nil, err
		}
	} else {
		if err := d.save(); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (d *JSONUserDB) load() error {
	// Load JSON.
	f, err := ioutil.ReadFile(d.filePath)
	if err != nil {
		return err
	}
	var db map[string]*model.User
	if err := json.Unmarshal(f, &db); err != nil {
		return err
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

func (d *JSONUserDB) Close() error {
	if err := d.save(); err != nil {
		return err
	}
	return nil
}

// Get gets an user or error.
func (d *JSONUserDB) Get(userID string) (*model.User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	u, ok := d.db[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

// Add adds an user.
func (d *JSONUserDB) Add(user *model.User) error {
	if _, err := d.Get(user.ID); err == nil {
		return errors.New("user already exist")
	}
	u := model.NewUser(user.ID, user.Name, user.Twitter.ID)
	d.mu.Lock()
	defer d.mu.Unlock()
	d.db[user.ID] = u
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

// NewJSONActivityDB initializes a JSONActivityDB
func NewJSONActivityDB(filePath string) (*JSONActivityDB, error) {
	d := &JSONActivityDB{
		filePath: filePath,
		db:       map[string]map[int64]*model.Activity{},
	}
	if isFile(d.filePath) {
		if err := d.load(); err != nil {
			return nil, err
		}
	} else {
		if err := d.save(); err != nil {
			return nil, err
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

func (d *JSONActivityDB) Close() error {
	if err := d.save(); err != nil {
		return err
	}
	return nil
}

// Add adds an activity.
// This method DOES NOT validate Activity.UserID in the given activity.
// If an activity in DB has same timestamp with the given activity, store the given one with timestamp++.
func (d *JSONActivityDB) Add(act *model.Activity) error {
	// init
	if chk, ok := d.db[act.UserID]; !ok || chk == nil {
		d.db[act.UserID] = map[int64]*model.Activity{}
	}
	ut := act.TimeUTC.Unix()
	d.mu.Lock()
	defer d.mu.Unlock()
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
	return nil
}

func (d *JSONActivityDB) Query(userID string, queryFn func(a *model.Activity) bool) ([]*model.Activity, error) {
	al, ok := d.db[userID]
	if !ok {
		return nil, errors.New("invalid user or the user has no activities")
	}
	var ret []*model.Activity
	for _, v := range al {
		if queryFn(v) {
			ret = append(ret, v)
		}
	}
	return ret, nil
}

func QueryFuncTime(afterUTC time.Time, beforeUTC time.Time) func(a *model.Activity) bool {
	return func(a *model.Activity) bool {
		if a.TimeUTC.After(afterUTC) && a.TimeUTC.Before(beforeUTC) {
			fmt.Printf("T: %v - %v | %v\n", afterUTC.String(), beforeUTC.String(), a.TimeUTC.String())
			return true
		}
		fmt.Printf("F: %v - %v | %v\n", afterUTC.String(), beforeUTC.String(), a.TimeUTC.String())
		return false
	}
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
