package app_test

import (
	"path/filepath"
	"testing"

	"github.com/ebiiim/goki/app"
	"github.com/ebiiim/goki/db"
)

const testdataDir = "./testdata"

func setupAppWithJSONDB(t *testing.T) *app.App {
	t.Helper()
	udb, err := db.NewJSONUserDB(filepath.Join(testdataDir, "JSONUserDB.json"))
	if err != nil {
		t.Error(err)
	}
	adb, err := db.NewJSONActivityDB(filepath.Join(testdataDir, "JSONActivityDB.json"))
	if err != nil {
		t.Error(err)
	}
	return app.NewApp(udb, adb)
}

func TestApp_Close(t *testing.T) {
	a := setupAppWithJSONDB(t)
	if err := a.Close(); err != nil {
		t.Error(err)
	}
}

func TestApp_GetUser(t *testing.T) {
	a := setupAppWithJSONDB(t)
	defer func() {
		if err := a.Close(); err != nil {
			t.Error(err)
		}
	}()
	// Just use the database so complicated tests are not needed.
	u, err := a.GetUser("123")
	if err != nil {
		t.Error(err)
	}
	if u.Name != "alice" {
		t.Error("err")
	}
}

func TestApp_AddUser(t *testing.T) {
	a := setupAppWithJSONDB(t) // Do not run a.Close to avoid save DB.
	// Just use the database so complicated tests are not needed.
	u, err := a.AddUser("000", "taro", "00000000")
	if err != nil {
		t.Error(err)
	}
	if u.Name != "taro" {
		t.Error("err")
	}
}

func TestApp_Action(t *testing.T) {
	a := setupAppWithJSONDB(t) // Do not run a.Close to avoid save DB.
	// Just use the database so complicated tests are not needed.
	u, _ := a.GetUser("123") // alice
	act, err := a.Action(u, 1, 10, 100)
	if err != nil {
		t.Error(err)
	}
	if act.G.S+act.G.M+act.G.L != 111 {
		t.Error("err")
	}
}
