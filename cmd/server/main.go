package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ebiiim/goki/api"
	"github.com/ebiiim/goki/app"
	"github.com/ebiiim/goki/config"
	"github.com/ebiiim/goki/db"
)

func main() {
	fmt.Println(config.Params)
	udb, err := db.NewJSONUserDB("./userDB.json")
	if err != nil {
		panic(err)
	}
	adb, err := db.NewJSONActivityDB("./activityDB.json")
	if err != nil {
		panic(err)
	}
	ap := app.NewApp(udb, adb)
	s := api.NewServer(ap)
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	done := make(chan struct{})
	go func() {
		if err := http.ListenAndServe(config.Params.Server.Address, s.R); err != nil {
			panic(err)
		}
	}()
	fmt.Printf("%s://%s\n", config.Params.Server.Scheme, config.Params.Server.Address)
	go func() {
		<-sigCh
		fmt.Println("SIGINT Received!")
		ap.Close()
		close(done)
	}()
	<-done
}
