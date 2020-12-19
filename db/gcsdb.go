package db

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/ebiiim/goki"
	"github.com/ebiiim/goki/model"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var gcsClient *storage.Client // client.Close need not be called at exit
var gcsAccessTimeout = 10 * time.Second

func initClientIfNeeded() error {
	if gcsClient != nil {
		return nil
	}
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, gcsAccessTimeout)
	defer cancelFunc()
	creds, err := google.FindDefaultCredentials(ctx, storage.ScopeReadWrite)
	if err != nil {
		return err
	}
	gcsClient, err = storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return err
	}
	return nil
}

// GCSUserDB is an easy UserDB stores data in a JSON file and saves it in GCS.
// Cannot be read from multiple app instances.
type GCSUserDB struct {
	bucket string
	file   string
	client *storage.Client

	db map[string]*model.User
	mu sync.Mutex
}

var _ UserDB = (*GCSUserDB)(nil)

// NewGCSUserDB initializes a GCSUserDB.
// A UserDB must be created in GCS.
func NewGCSUserDB(bucket, file string) (*GCSUserDB, error) {
	// init shared GCS Client
	if err := initClientIfNeeded(); err != nil {
		return nil, err
	}

	d := &GCSUserDB{
		bucket: bucket,
		file:   file,
		client: gcsClient,
		db:     map[string]*model.User{},
	}
	if err := d.load(); err != nil {
		return nil, goki.ErrWrap(goki.ErrDBOpen, err)
	}
	return d, nil
}

func (d *GCSUserDB) load() error {
	// Load JSON.
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, gcsAccessTimeout)
	defer cancelFunc()
	// logf(DEBUG, "GCSUserDB.load: read GCS")
	obj := d.client.Bucket(d.bucket).Object(d.file)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()
	var db map[string]*model.User
	if err := json.NewDecoder(reader).Decode(&db); err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *GCSUserDB) save() error {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, gcsAccessTimeout)
	defer cancelFunc()
	// logf(DEBUG, "GCSUserDB.save: write GCS")
	obj := d.client.Bucket(d.bucket).Object(d.file)
	writer := obj.NewWriter(ctx)
	if err := json.NewEncoder(writer).Encode(&d.db); err != nil {
		return err
	}
	// Writes happen asynchronously!
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

// Close saves data to the database JSON file.
func (d *GCSUserDB) Close() error {
	if err := d.save(); err != nil {
		return goki.ErrWrap(goki.ErrDBClose, err)
	}
	return nil
}

// Get gets an user or error.
func (d *GCSUserDB) Get(userID string) (*model.User, error) {
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
func (d *GCSUserDB) GetByTwitterID(twitterID string) (*model.User, error) {
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
func (d *GCSUserDB) Add(user *model.User) error {
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

// GCSActivityDB is an easy ActivityDB stores data in a JSON file and saves it in GCS.
// Cannot be read from multiple app instances.
type GCSActivityDB struct {
	bucket string
	file   string
	client *storage.Client

	// UserID -> time.Unix -> Activity
	db map[string]map[int64]*model.Activity
	mu sync.Mutex
}

var _ ActivityDB = (*GCSActivityDB)(nil)

// NewGCSActivityDB initializes a GCSActivityDB
// An ActivityDB must be created in GCS.
func NewGCSActivityDB(bucket, file string) (*GCSActivityDB, error) {
	// init shared GCS Client
	if err := initClientIfNeeded(); err != nil {
		return nil, err
	}

	d := &GCSActivityDB{
		bucket: bucket,
		file:   file,
		client: gcsClient,
		db:     map[string]map[int64]*model.Activity{},
	}
	if err := d.load(); err != nil {
		return nil, goki.ErrWrap(goki.ErrDBOpen, err)
	}
	return d, nil
}

func (d *GCSActivityDB) load() error {
	// Load JSON.
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, gcsAccessTimeout)
	defer cancelFunc()
	// logf(DEBUG, "GCSActivityDB.load: read GCS")
	obj := d.client.Bucket(d.bucket).Object(d.file)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()
	var db map[string]map[int64]*model.Activity
	if err := json.NewDecoder(reader).Decode(&db); err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *GCSActivityDB) save() error {
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, gcsAccessTimeout)
	defer cancelFunc()
	// logf(DEBUG, "GCSActivityDB.save: write GCS")
	obj := d.client.Bucket(d.bucket).Object(d.file)
	writer := obj.NewWriter(ctx)
	if err := json.NewEncoder(writer).Encode(&d.db); err != nil {
		return err
	}
	// Writes happen asynchronously!
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

// Close saves data to the database JSON file.
func (d *GCSActivityDB) Close() error {
	if err := d.save(); err != nil {
		return goki.ErrWrap(goki.ErrDBClose, err)
	}
	return nil
}

// Add adds an activity.
// This method DOES NOT validate Activity.UserID in the given activity.
// In this UserDB implementation, if an activity in DB has same timestamp with the given activity, then store the given one with timestamp++.
// Always returns nil
func (d *GCSActivityDB) Add(act *model.Activity) error {
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
func (d *GCSActivityDB) Query(userID string, queryFn func(a *model.Activity) bool) ([]*model.Activity, error) {
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
