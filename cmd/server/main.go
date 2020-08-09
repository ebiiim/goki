package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/sessions"

	"github.com/ebiiim/goki/api"
	"github.com/ebiiim/goki/app"
	"github.com/ebiiim/goki/config"
	"github.com/ebiiim/goki/db"
)

func main() {
	udb, err := db.NewJSONUserDB("./userDB.json")
	if err != nil {
		panic(err)
	}
	adb, err := db.NewJSONActivityDB("./activityDB.json")
	if err != nil {
		panic(err)
	}
	ap := app.NewApp(udb, adb)
	ss := sessions.NewFilesystemStore("./sessions", []byte(config.Params.Session.Key))
	s := api.NewServer(ap, ss)
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	done := make(chan struct{})
	go func() {
		if err := http.ListenAndServe(config.Params.Server.Address, s.R); err != nil {
			log.Fatalln(err)
		}
	}()
	log.Printf("%s://%s\n", config.Params.Server.Scheme, config.Params.Server.Address)
	go func() {
		<-sigCh
		log.Println("SIGINT Received!")
		if err := s.Close(); err != nil {
			log.Println(err)
		}
		close(done)
	}()
	<-done
}
