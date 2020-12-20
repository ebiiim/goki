package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/acme/autocert"

	"github.com/ebiiim/goki/app"
	"github.com/ebiiim/goki/config"
	"github.com/ebiiim/goki/db"
	"github.com/ebiiim/goki/server"
)

const (
	userDBPath     = "./userDB.json"
	activityDBPath = "./activityDB.json"
	sessionDirPath = "./sessions"
)

func main() {
	udb, err := db.NewJSONUserDB(userDBPath)
	if err != nil {
		log.Fatalf("could not load user database file %s: %v", userDBPath, err)
	}
	adb, err := db.NewJSONActivityDB(activityDBPath)
	if err != nil {
		log.Fatalf("could not load activity database file %s: %v", activityDBPath, err)
	}
	ap := app.NewApp(udb, adb)
	ss := sessions.NewFilesystemStore(sessionDirPath, []byte(config.Params.Session.Key))
	s := server.NewServer(config.Params.Server.Address, ap, ss)
	go func() {
		switch scheme := config.Params.Server.Scheme; scheme {
		case "http":
			if err := s.ListenAndServe(); err != nil { // err will be returned when call s.Shutdown
				log.Printf("server closed: %v", err)
			}
		case "https":
			if err := http.Serve(autocert.NewListener(config.Params.Server.Address), s.Handler); err != nil { // err will be returned when call s.Shutdown
				log.Printf("server closed: %v", err)
			}
		default:
			log.Printf("invalid scheme: %v", scheme)
		}
	}()
	log.Printf("%s://%s\n", config.Params.Server.Scheme, config.Params.Server.Address)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("SIGINT Received!")

	ctx, cancel := context.WithTimeout(context.Background(), config.ServerShutdownTimeout)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Println(err)
	}
}
