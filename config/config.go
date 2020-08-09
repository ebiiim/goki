package config

import (
	"encoding/json"
	"io/ioutil"
)

const (
	SessionName          = "goki"
	SessionSecret        = "secret"
	SessionAuthenticated = "authenticated"
	SessionTwitterID     = "twitter_id"
	SessionUserID        = "user_id"
)

type config struct {
	Server struct {
		Scheme   string `json:"scheme"`
		Address  string `json:"address"`
		BasePath string `json:"base_path"`
	} `json:"server"`
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
	f, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(f, &Params); err != nil {
		panic(err)
	}
}