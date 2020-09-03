package api

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/dghubble/gologin/v2/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
	"github.com/ebiiim/logo"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/ebiiim/goki"
	"github.com/ebiiim/goki/app"
	"github.com/ebiiim/goki/config"
	"github.com/ebiiim/goki/model"
)

// Log is the Logo logger for this package.
var Log = logo.New(logo.DEBUG, nil)

// ctxKey identifies the key used in context.WithValue.
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

// tmplKey identifies the page.
type tmplKey int

const (
	tmplTop tmplKey = iota
	tmplMe
	tmplDo
)

// template helper
func (s *Server) mustTmpl(key tmplKey, filenames ...string) {
	t := template.Must(template.ParseFiles(filenames...))
	s.T[key] = t
}

// aliases
var (
	pathBase            = config.Params.Server.BasePath
	pathStatic          = path.Join(pathBase, "static/")
	pathTop             = path.Join(pathBase, "")
	pathMe              = path.Join(pathBase, "me")
	pathDo              = path.Join(pathBase, "do")
	pathLogout          = path.Join(pathBase, "logout")
	pathTwitterLogin    = path.Join(pathBase, "login/twitter")
	pathTwitterCallback = config.Params.Twitter.CallbackPath
	urlTwitterCallback  = fmt.Sprintf("%s://%s%s", config.Params.Server.Scheme, config.Params.Server.Address, config.Params.Twitter.CallbackPath)
	dirTmpl             = config.Params.Web.TemplateDir
	dirStatic           = config.Params.Web.StaticDir
)

// Server contains everything to serve the web service.
type Server struct {
	http.Server
	A *app.App
	S sessions.Store
	T map[tmplKey]*template.Template
}

// NewServer initializes a Server.
func NewServer(scheme, addr string, ap *app.App, ss sessions.Store) *Server {
	r := mux.NewRouter()

	s := &Server{}
	s.A = ap
	s.S = ss
	s.T = map[tmplKey]*template.Template{}
	s.Handler = r
	s.WriteTimeout = config.ServerWriteTimeout
	s.ReadTimeout = config.ServerReadTimeout
	s.IdleTimeout = config.ServerIdleTimeout
	s.Addr = addr
	if scheme == "https" {
		panic("not implemented") // TODO
	}

	// Route and Template
	if config.Params.Web.ServeStatic {
		r.PathPrefix(pathStatic).Handler(http.StripPrefix(pathStatic, http.FileServer(http.Dir(dirStatic))))
	}

	r.HandleFunc(pathTop, s.checkLogin(s.serveTop))
	s.mustTmpl(tmplTop, filepath.Join(dirTmpl, "top.html"), filepath.Join(dirTmpl, "_head.html"), filepath.Join(dirTmpl, "_footer.html"))

	r.HandleFunc(pathMe, s.checkLogin(s.notLoggedInGoTop(s.serveMe)))
	s.mustTmpl(tmplMe, filepath.Join(dirTmpl, "me.html"), filepath.Join(dirTmpl, "_head.html"), filepath.Join(dirTmpl, "_footer.html"))

	r.HandleFunc(pathDo, s.checkLogin(s.notLoggedInGoTop(s.serveDo)))
	s.mustTmpl(tmplDo, filepath.Join(dirTmpl, "do.html"), filepath.Join(dirTmpl, "_head.html"), filepath.Join(dirTmpl, "_footer.html"))

	r.HandleFunc(pathLogout, s.serveLogout)

	// Twitter login
	oauth1Config := &oauth1.Config{
		ConsumerKey:    config.Params.Twitter.Key,
		ConsumerSecret: config.Params.Twitter.Secret,
		CallbackURL:    urlTwitterCallback,
		Endpoint:       twitterOAuth1.AuthorizeEndpoint,
	}
	r.Handle(pathTwitterLogin, twitter.LoginHandler(oauth1Config, nil))
	r.Handle(pathTwitterCallback, twitter.CallbackHandler(oauth1Config, s.twitterLogin(), nil))

	return s
}

// Shutdown closes the server.
func (s *Server) Shutdown(ctx context.Context) error {
	err1 := s.Server.Shutdown(ctx)
	err2 := s.A.Close()
	if err1 != nil || err2 != nil {
		return fmt.Errorf("Server.Close: err1=%v err2=%v", err1, err2)
	}
	return nil
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
			http.Redirect(w, r, pathTop, http.StatusFound)
			return
		}
		next(w, r)
	}
}

// twitterLogin handles Twitter OAuth1 callback.
// - Check Twitter user.
//   - (A) Error: redirect to the top page.
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
			http.Redirect(w, r, pathTop, http.StatusFound)
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
		http.Redirect(w, r, pathMe, http.StatusFound)
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
		http.Redirect(w, r, pathTop, http.StatusFound)
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
	http.Redirect(w, r, pathTop, http.StatusFound)
	return // (A)
}

func (s *Server) serveTop(w http.ResponseWriter, r *http.Request) {
	Log.D("serveTop")

	tmplStruct := struct {
		IsLoggedIn bool
		UserName   string
	}{}

	u, ok := r.Context().Value(ctxLoginUser).(*model.User)
	if ok && u != nil {
		Log.D("serveTop: logged in")
		tmplStruct.IsLoggedIn = true
		tmplStruct.UserName = u.Name
	}

	if err := s.T[tmplTop].Execute(w, tmplStruct); err != nil {
		Log.D("serveTop: template.Execute error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) serveMe(w http.ResponseWriter, r *http.Request) {
	Log.D("serveMe")

	tmplStruct := struct {
		UserName string
		G        *model.Goki
		Year     int
	}{}

	u, _ := r.Context().Value(ctxLoginUser).(*model.User)
	year := time.Now().Year()
	g, err := s.A.CountByYear(u.ID, year, time.Local)
	if err != nil {
		Log.D("serveMe: could not CountByYear")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmplStruct.UserName = u.Name
	tmplStruct.G = g
	tmplStruct.Year = year

	if err := s.T[tmplMe].Execute(w, tmplStruct); err != nil {
		Log.D("serveMe: template.Execute error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) serveDo(w http.ResponseWriter, r *http.Request) {
	Log.D("serveDo")

	tmplStruct := struct {
		UserName string
		FormMax  []struct{}
	}{
		FormMax: make([]struct{}, 21), // HACK: range(0, 21)
	}

	u, _ := r.Context().Value(ctxLoginUser).(*model.User)
	tmplStruct.UserName = u.Name

	if err := s.T[tmplDo].Execute(w, tmplStruct); err != nil {
		Log.D("serveMe: template.Execute error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
