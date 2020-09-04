package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const (
	SessionName           = "goki"
	SessionUserID         = "user_id"
	ServerWriteTimeout    = 15 * time.Second
	ServerReadTimeout     = 15 * time.Second
	ServerIdleTimeout     = 60 * time.Second
	ServerShutdownTimeout = 60 * time.Second
)

type config struct {
	Server struct {
		Scheme   string `json:"scheme"`
		Address  string `json:"address"`
		BasePath string `json:"base_path"`
	} `json:"server"`
	Web struct {
		TemplateDir string `json:"template_dir"`
		StaticDir   string `json:"static_dir"`
		ServeStatic bool   `json:"serve_static"`
	} `json:"web"`
	Session struct {
		Key string `json:"key"`
	} `json:"session"`
	Twitter struct {
		Key             string `json:"key"`
		Secret          string `json:"secret"`
		RequestURL      string `json:"request_url"`
		AuthorizeURL    string `json:"authorize_url"`
		TokenRequestURL string `json:"token_request_url"`
		// CallbackPath does not consider config.Server.BasePath.
		// {config.Server.Scheme}://{config.Server.Address}{CallbackPath}
		CallbackPath string `json:"callback_path"`
	} `json:"twitter"`
}

var Params config

func init() {
	p, ok := os.LookupEnv("GOKI_CONFIG")
	if !ok {
		p = "./config.json"
	}
	f, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatalf("[FATAL] could not load config file %v: %v", p, err)
	}
	if err := json.Unmarshal(f, &Params); err != nil {
		log.Fatalf("[FATAL] could not decode config file %v: %v", p, err)
	}
	// override credentials if env is set.
	tk, ok := os.LookupEnv("TWITTER_CONSUMER_KEY")
	if ok {
		Params.Twitter.Key = tk
	}
	ts, ok := os.LookupEnv("TWITTER_CONSUMER_SECRET")
	if ok {
		Params.Twitter.Secret = ts
	}
}
