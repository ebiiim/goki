module github.com/ebiiim/goki

go 1.15

replace github.com/GoogleCloudPlatform/firestore-gorilla-sessions v0.1.0 => github.com/ebiiim/firestore-gorilla-sessions v0.1.1

require (
	cloud.google.com/go/firestore v1.4.0
	cloud.google.com/go/storage v1.10.0
	github.com/GoogleCloudPlatform/firestore-gorilla-sessions v0.1.0
	github.com/dghubble/gologin/v2 v2.2.0
	github.com/dghubble/oauth1 v0.6.0
	github.com/ebiiim/logo v0.1.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.1
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58
	google.golang.org/api v0.36.0
)
