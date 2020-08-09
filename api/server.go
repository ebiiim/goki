package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/ebiiim/goki"
	"github.com/ebiiim/goki/app"
	"github.com/ebiiim/goki/config"
	"github.com/ebiiim/goki/model"

	"github.com/dghubble/gologin/v2/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
	"github.com/ebiiim/logo"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

// Log is the Logo logger for this package.
var Log = logo.New(logo.DEBUG, nil)

// Server contains everything to serve the web service.
type Server struct {
	A *app.App
	R *mux.Router
	S sessions.Store
}

// ctxKey represents keys used in context.WithValue.
type ctxKey int

const (
	ctxLoginUser ctxKey = iota
)

// check session values
func valuesExist(sess *sessions.Session, key ...string) bool {
	for _, k := range key {
		if sess.Values[k] == nil {
			return false
		}
	}
	return true
}

// aliases
var base = config.Params.Server.BasePath

// NewServer initializes a Server.
func NewServer(ap *app.App) *Server {
	s := &Server{}
	s.A = ap
	s.R = mux.NewRouter()
	s.S = sessions.NewFilesystemStore("./sessions", []byte(config.Params.Session.Key))
	s.R.HandleFunc(base, s.checkLogin(s.serveTop))
	s.R.HandleFunc(path.Join(base, "me"), s.checkLogin(s.notLoggedInGoTop(s.serveMe)))
	s.R.HandleFunc(path.Join(base, "logout"), s.serveLogout)
	// Twitter login
	cb := fmt.Sprintf("%s://%s%s", config.Params.Server.Scheme, config.Params.Server.Address, config.Params.Twitter.CallbackPath)
	oauth1Config := &oauth1.Config{
		ConsumerKey:    config.Params.Twitter.Key,
		ConsumerSecret: config.Params.Twitter.Secret,
		CallbackURL:    cb,
		Endpoint:       twitterOAuth1.AuthorizeEndpoint,
	}
	s.R.Handle(path.Join(base, "login/twitter"), twitter.LoginHandler(oauth1Config, nil))
	s.R.Handle(path.Join(base, config.Params.Twitter.CallbackPath), twitter.CallbackHandler(oauth1Config, s.twitterLogin(), nil))
	return s
}

// checkLogin middleware checks login.
// - Get the Goki user ID from session and verify it.
//   - (A) Success: put the user into context value `ctxLoginUser` and go next
//   - (B) Error: just go next
//   - (X) Unexpected error: 500
func (s *Server) checkLogin(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		Log.D("checkLogin")
		sess, err := s.S.Get(r, config.SessionName)
		if err != nil {
			Log.D("checkLogin: error while getting session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return // (X)
		}
		if !valuesExist(sess, config.SessionUserID) {
			Log.D("checkLogin: no data in session so just go next")
			next(w, r)
			return // (B)
		}
		Log.D("checkLogin: check the user id")
		userID, _ := sess.Values[config.SessionUserID].(string) // already validated
		u, err := s.A.GetUser(userID)
		if err != nil {
			if errors.Is(err, goki.ErrUserNotFound) {
				//invalid user: delete the session and go next
				Log.D("checkLogin: user not found (invalid user id in the session) so delete session and go next")
				sess.Options.MaxAge = -1
				if err := sess.Save(r, w); err != nil {
					Log.E("checkLogin: error while saving session")
				}
				// anyway, go next
				next(w, r)
				return // (B)
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return // (X)
		}
		Log.D("checkLogin: ok! set context")
		ctx := context.WithValue(context.Background(), ctxLoginUser, u)
		r = r.WithContext(ctx)
		next(w, r)
	}
}

// notLoggedInGoTop middleware redirects unauthenticated users to the top page.
func (s *Server) notLoggedInGoTop(next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		Log.D("notLoggedInGoTop")
		u, ok := r.Context().Value(ctxLoginUser).(*model.User)
		if !ok || u == nil {
			Log.D("notLoggedInGoTop: go top")
			http.Redirect(w, r, path.Join(base, "login"), http.StatusFound)
			return
		}
		next(w, r)
	}
}

// twitterLogin handles Twitter OAuth1 callback.
// - Check Twitter user.
//   - (A) Error: redirect to the login page.
//   - (B) New Twitter user: create a new Goki user and login.
//   - (C) Known Twitter user: login with the associated Goki user and login.
//   - (X) Unexpected error:  500
func (s *Server) twitterLogin() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		Log.D("twitterLogin")
		ctx := r.Context()
		twitterUser, err := twitter.UserFromContext(ctx)
		if err != nil {
			Log.D("twitterLogin: twitter oauth failed")
			http.Redirect(w, r, path.Join(base, "login"), http.StatusFound)
			return // (A)
		}
		// check Twitter User
		Log.D("twitterLogin: check twitter user")
		user, err := s.A.GetUserByTwitterID(twitterUser.IDStr)
		if err != nil {
			if errors.Is(err, goki.ErrUserNotFound) {
				Log.D("twitterLogin: create a new Goki user for twitter user %v", twitterUser.IDStr)
				iU, iErr := s.A.AddUser(goki.NewID(), twitterUser.Name, twitterUser.IDStr)
				if iErr != nil {
					// somehow failed to create a new user
					Log.D("twitterLogin: failed to create a new Goki user")
					http.Error(w, iErr.Error(), http.StatusInternalServerError)
					return // (X)
				}
				user = iU
			} else {
				//	// some error from App or UserDB
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return // (X)
			}
		}
		// make session
		Log.D("twitterLogin: make session")
		sess, err := s.S.New(r, config.SessionName)
		if err != nil {
			Log.D("twitterLogin: failed to make session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return // (X)
		}
		sess.Values[config.SessionUserID] = user.ID
		Log.D("twitterLogin: save session")
		if err := sess.Save(r, w); err != nil {
			Log.E("twitterLogin: failed to save session")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return // (X)
		}
		Log.D("twitterLogin: redirect to /me")
		http.Redirect(w, r, path.Join(base, "me"), http.StatusFound)
		return // (B) or (C)
	}
	return http.HandlerFunc(fn)
}

// serveLogout handles logout page.
// - Delete the session.
//   - (A) Success: redirect to the top page
//   - (X) Unexpected error: 500
func (s *Server) serveLogout(w http.ResponseWriter, r *http.Request) {
	Log.D("serveLogout")
	sess, err := s.S.Get(r, config.SessionName)
	if err != nil {
		Log.E("serveLogout: error while getting session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return // (X)
	}
	if sess.ID == "" {
		Log.D("serveLogout: no session so redirect to /top")
		http.Redirect(w, r, base, http.StatusFound)
		return // (A)
	}
	Log.D("serveLogout: delete session")
	sess.Options.MaxAge = -1 // delete session
	if err := sess.Save(r, w); err != nil {
		Log.E("serveLogout: error while saving session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return // (X)
	}
	Log.D("serveLogin: ok! now redirect to /top")
	http.Redirect(w, r, base, http.StatusFound)
	return // (A)
}

func (s *Server) serveTop(w http.ResponseWriter, r *http.Request) {
	Log.D("serveTop")
	u, ok := r.Context().Value(ctxLoginUser).(*model.User)
	isLoggedIn := true
	if !ok || u == nil {
		Log.D("serveTop: not login")
		isLoggedIn = false
	}
	if isLoggedIn {
		fmt.Fprintf(w, "<html><body>Gやっつけた！<br>今まで駆除したGの数を覚えていますか？<br>%vさんですね！ <a href='/me'>マイページ</a> <a href='/logout'>ログアウト</a></body>", u.Name)
		return
	}
	fmt.Fprintf(w, "<html><body>Gやっつけた！<br>今まで駆除したGの数を覚えていますか？<br><a href='/login/twitter'>Login with Twitter</a></body>")
	return
}

func (s *Server) serveMe(w http.ResponseWriter, r *http.Request) {
	Log.D("serveMe")
	u, _ := r.Context().Value(ctxLoginUser).(*model.User)
	g, err := s.A.CountByYear(u.ID, time.Now().Year(), time.Local)
	if err != nil {
		fmt.Fprintf(w, "err")
		return
	}
	fmt.Fprintf(w, "<html><body>Gやっつけた！<br>%vさん、こんにちは！<br>今年の戦績: %v<br><a href='/'>トップ</a> <a href='/logout'>ログアウト</a></body>", u.Name, g)
	return
}
