package db_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ebiiim/goki/db"
	"github.com/ebiiim/goki/model"
)

const testdataDir = "./testdata"

func removeFile(t *testing.T, f string) {
	t.Helper()
	if err := os.Remove(f); err != nil {
		t.Fatal(err)
	}
}

var (
	JST, _         = time.LoadLocation("Asia/Tokyo")
	U1             = model.NewUser("123", "alice", "12345678")
	U2             = model.NewUser("456", "bob", "87654321")
	A1t            = time.Date(2020, 8, 1, 9, 0, 0, 0, time.UTC)
	A1             = model.NewActivity(U1.ID, A1t, 3, 0, 0)
	A2t            = time.Date(2020, 8, 2, 10, 10, 10, 10, time.UTC)
	A2             = model.NewActivity(U1.ID, A2t, 3, 3, 0)
	A3t            = time.Date(2020, 8, 1, 9, 0, 0, 0, time.UTC)
	A3             = model.NewActivity(U2.ID, A3t, 0, 0, 12345678)
	UTC202008Begin = time.Date(2020, 8, 1, 0, 0, 0, 0, time.UTC)
	UTC202009Begin = time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC)
	JST202008Begin = time.Date(2020, 8, 1, 0, 0, 0, 0, JST).In(time.UTC)
	JST202009Begin = time.Date(2020, 9, 1, 0, 0, 0, 0, JST).In(time.UTC)
	UTC202109Begin = time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC)
	UTC202110Begin = time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC)
	JST202109Begin = time.Date(2021, 9, 1, 0, 0, 0, 0, JST).In(time.UTC)
	JST202110Begin = time.Date(2021, 10, 1, 0, 0, 0, 0, JST).In(time.UTC)
)

func TestNewJSONUserDB_NewFile(t *testing.T) {
	var testDBPath = "JSONUserDB_NewFile.json"
	defer removeFile(t, testDBPath)
	d, err := db.NewJSONUserDB(testDBPath)
	if err != nil {
		t.Error(err)
		return
	}
	if err := d.Close(); err != nil {
		t.Error(err)
	}
}

func TestNewJSONUserDB_OpenFile(t *testing.T) {
	var testDBPath = filepath.Join(testdataDir, "JSONUserDB_OpenFile.json")
	d, err := db.NewJSONUserDB(testDBPath)
	if err != nil {
		t.Error(t)
		return
	}
	if err := d.Close(); err != nil {
		t.Error(err)
	}
}

func TestJSONUserDB_Get(t *testing.T) {
	var testDBPath = filepath.Join(testdataDir, "JSONUserDB_Get.json")
	d, err := db.NewJSONUserDB(testDBPath)
	if err != nil {
		t.Error(t)
		return
	}
	if err := d.Close(); err != nil {
		t.Error(err)
	}
	cases := []struct {
		name     string
		d        *db.JSONUserDB
		userID   string
		userName string
		isErr    bool
	}{
		{"alice", d, U1.ID, U1.Name, false},
		{"bob", d, U2.ID, U2.Name, false},
		{"F_taro", d, "000", "taro", true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			u, err := c.d.Get(c.userID)
			if c.isErr {
				if err == nil {
					t.Error("expected err")
					return
				}
				return
			}
			if err != nil {
				t.Error("err")
				return
			}
			if u.ID != c.userID || u.Name != c.userName {
				t.Error("data")
				return
			}
		})
	}
}

func TestJSONUserDB_GetByTwitterID(t *testing.T) {
	var testDBPath = filepath.Join(testdataDir, "JSONUserDB_Get.json")
	d, err := db.NewJSONUserDB(testDBPath)
	if err != nil {
		t.Error(t)
		return
	}
	if err := d.Close(); err != nil {
		t.Error(err)
	}
	cases := []struct {
		name      string
		d         *db.JSONUserDB
		twitterID string
		userID    string
		isErr     bool
	}{
		{"alice", d, U1.Twitter.ID, U1.ID, false},
		{"bob", d, U2.Twitter.ID, U2.ID, false},
		{"F_taro", d, "00000000", "000", true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			u, err := c.d.GetByTwitterID(c.twitterID)
			if c.isErr {
				if err == nil {
					t.Error("expected err")
					return
				}
				return
			}
			if err != nil {
				t.Error("err")
				return
			}
			if u.ID != c.userID {
				t.Error("data")
				return
			}
		})
	}
}

func TestJSONUserDB_Add(t *testing.T) {
	var testDBPath = "JSONUserDB_Add.json"
	defer removeFile(t, testDBPath)
	d, err := db.NewJSONUserDB(testDBPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := d.Close(); err != nil {
			t.Error(err)
		}
	}()
	cases := []struct {
		name  string
		d     *db.JSONUserDB
		user  *model.User
		isErr bool
	}{
		{"alice", d, U1, false},
		{"bob", d, U2, false},
		{"F_alice2", d, U1, true},
		{"F_taro_Twitter_duplicated", d, model.NewUser("000", "taro", "12345678"), true},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			err := c.d.Add(c.user)
			if c.isErr {
				if err == nil {
					t.Error("expected err")
					return
				}
				return
			}
			if err != nil {
				t.Error("err")
				return
			}
		})
	}
}

func TestNewJSONActivityDB_NewFile(t *testing.T) {
	var testDBPath = "JSONActivityDB_NewFile.json"
	defer removeFile(t, testDBPath)
	d, err := db.NewJSONActivityDB(testDBPath)
	if err != nil {
		t.Error(err)
		return
	}
	if err := d.Close(); err != nil {
		t.Error(err)
	}
}

func TestNewJSONActivityDB_OpenFile(t *testing.T) {
	var testDBPath = filepath.Join(testdataDir, "JSONActivityDB_OpenFile.json")
	d, err := db.NewJSONActivityDB(testDBPath)
	if err != nil {
		t.Error(t)
		return
	}
	if err := d.Close(); err != nil {
		t.Error(err)
	}
}

func TestJSONActivityDB_Add(t *testing.T) {
	var testDBPath = "JSONActivityDB_Add.json"
	defer removeFile(t, testDBPath)
	d, err := db.NewJSONActivityDB(testDBPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := d.Close(); err != nil {
			t.Error(err)
		}
	}()
	cases := []struct {
		name  string
		d     *db.JSONActivityDB
		a     *model.Activity
		isErr bool
	}{
		{"alice1", d, A1, false},
		{"alice2", d, A2, false},
		{"alice2_same_time", d, A2, false},
		{"bob1", d, A3, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			err := c.d.Add(c.a)
			if c.isErr {
				if err == nil {
					t.Error("expected err")
					return
				}
				return
			}
			if err != nil {
				t.Error("err")
				return
			}
		})
	}
}

func TestJSONActivityDB_Query(t *testing.T) {
	var testDBPath = filepath.Join(testdataDir, "JSONActivityDB_Query.json")
	d, err := db.NewJSONActivityDB(testDBPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		if err := d.Close(); err != nil {
			t.Error(err)
		}
	}()
	cases := []struct {
		name    string
		d       *db.JSONActivityDB
		userID  string
		queryFn func(a *model.Activity) bool
		expNum  int
		isErr   bool
	}{
		{"taro_invalid_user", d, "000", db.QueryFuncTime(UTC202008Begin, UTC202009Begin), 0, false},
		{"alice_UTC202008", d, U1.ID, db.QueryFuncTime(UTC202008Begin, UTC202009Begin), 3, false},
		{"alice_JST202008", d, U1.ID, db.QueryFuncTime(JST202008Begin, JST202009Begin), 2, false},
		{"alice_UTC202109", d, U1.ID, db.QueryFuncTime(UTC202109Begin, UTC202110Begin), 0, false},
		{"alice_JST202109", d, U1.ID, db.QueryFuncTime(JST202109Begin, JST202110Begin), 1, false},
		{"bob_UTC202008", d, U2.ID, db.QueryFuncTime(UTC202008Begin, UTC202009Begin), 1, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			res, err := c.d.Query(c.userID, c.queryFn)
			if c.isErr {
				if err == nil {
					t.Error("expected err")
					return
				}
				return
			}
			if err != nil {
				t.Error("err")
				return
			}
			if len(res) != c.expNum {
				t.Errorf("want %v but got %v", c.expNum, len(res))
				return
			}
		})
	}
}
