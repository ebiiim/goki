package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/firestore"
	sessions "github.com/GoogleCloudPlatform/firestore-gorilla-sessions"
	"github.com/ebiiim/goki/app"
	"github.com/ebiiim/goki/config"
	"github.com/ebiiim/goki/db"
	"github.com/ebiiim/goki/server"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var (
	userDB                    string = "userDB.json"
	activityDB                string = "activityDB.json"
	gcpProjectID              string = os.Getenv("GCP_ID")
	dbBucket                  string = os.Getenv("GCS_DB_BUCKET")
	twitterCallbackServerName string = os.Getenv("TWITTER_CALLBACK_SERVER_NAME")
)

func main() {
	// HACK
	server.UrlTwitterCallback = twitterCallbackServerName + config.Params.Twitter.CallbackPath

	udb, err := db.NewGCSUserDB(dbBucket, userDB)
	if err != nil {
		log.Fatalf("could not load user database: %v", err)
	}
	adb, err := db.NewGCSActivityDB(dbBucket, activityDB)
	if err != nil {
		log.Fatalf("could not load activity database: %v", err)
	}
	ap := app.NewApp(udb, adb)

	// session store
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFunc()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/datastore")
	if err != nil {
		log.Fatalf("could not get GCP credentials: %v", err)
	}
	client, err := firestore.NewClient(ctx, gcpProjectID, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("could not open firestore: %v", err)
	}
	defer client.Close()
	ss, err := sessions.New(ctx, client)
	if err != nil {
		log.Fatalf("could not init sessions: %v", err)
	}

	// server
	s := server.NewServer(config.Params.Server.Address, ap, ss)
	go func() {
		switch scheme := config.Params.Server.Scheme; scheme {
		case "http":
			if err := s.ListenAndServe(); err != nil { // err will be returned when call s.Shutdown
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
