package app_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ebiiim/goki/app"
	"github.com/ebiiim/goki/db"
)

const testdataDir = "./testdata"

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	b, err := ioutil.ReadFile(src)
	if err != nil {
		t.Error(err)
	}
	if err := ioutil.WriteFile(dst, b, 0755); err != nil {
		t.Error(err)
	}
}

func removeFile(t *testing.T, target string) {
	t.Helper()
	if err := os.Remove(target); err != nil {
		t.Error(err)
	}
}

func setupAppWithJSONDB(t *testing.T) (_ *app.App, cleanupFn func()) {
	t.Helper()

	udbGoldenFile := filepath.Join(testdataDir, "JSONUserDB.json")
	udbFile := filepath.Join(testdataDir, fmt.Sprintf("JSONUserDB_%s.json", t.Name()))
	copyFile(t, udbGoldenFile, udbFile)
	udb, err := db.NewJSONUserDB(udbFile)
	if err != nil {
		t.Error(err)
	}

	adbGoldenFile := filepath.Join(testdataDir, "JSONActivityDB.json")
	adbFile := filepath.Join(testdataDir, fmt.Sprintf("JSONActivityDB_%s.json", t.Name()))
	copyFile(t, adbGoldenFile, adbFile)
	adb, err := db.NewJSONActivityDB(adbFile)
	if err != nil {
		t.Error(err)
	}

	cleanupFn = func() {
		removeFile(t, udbFile)
		removeFile(t, adbFile)
	}

	return app.NewApp(udb, adb), cleanupFn
}

func TestApp_Close(t *testing.T) {
	a, cleanupFn := setupAppWithJSONDB(t)
	defer cleanupFn()
	if err := a.Close(); err != nil {
		t.Error(err)
	}
}

func TestApp_GetUser(t *testing.T) {
	a, cleanupFn := setupAppWithJSONDB(t)
	defer cleanupFn()
	defer func() {
		if err := a.Close(); err != nil {
			t.Error(err)
		}
	}()
	// Just try to use the database so complicated tests are not needed.
	u, err := a.GetUser("123")
	if err != nil {
		t.Error(err)
	}
	if u.Name != "alice" {
		t.Error("err")
	}
}

func TestApp_AddUser(t *testing.T) {
	a, cleanupFn := setupAppWithJSONDB(t)
	defer cleanupFn()
	defer func() {
		if err := a.Close(); err != nil {
			t.Error(err)
		}
	}()
	// Just try to use the database so complicated tests are not needed.
	u, err := a.AddUser("000", "taro", "00000000")
	if err != nil {
		t.Error(err)
	}
	if u.Name != "taro" {
		t.Error("err")
	}
}

func TestApp_Action(t *testing.T) {
	a, cleanupFn := setupAppWithJSONDB(t)
	defer cleanupFn()
	defer func() {
		if err := a.Close(); err != nil {
			t.Error(err)
		}
	}()
	// Just try to use the database so complicated tests are not needed.
	u, _ := a.GetUser("123") // alice
	act, err := a.Action(u, 1, 10, 100)
	if err != nil {
		t.Error(err)
	}
	if act.G.S+act.G.M+act.G.L != 111 {
		t.Error("err")
	}
}
